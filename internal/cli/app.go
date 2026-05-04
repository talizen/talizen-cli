package cli

import (
	"bysir/talizen-cli/internal/talizen"
	"context"
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

const (
	defaultAPIHostValue = "https://talizen.com"
	defaultWebHostValue = "https://talizen.com"
)

var version = "dev"

func defaultAPIHost() string {
	if v := strings.TrimSpace(os.Getenv("TALIZEN_API_HOST")); v != "" {
		return v
	}

	return defaultAPIHostValue
}

func defaultWebHost(apiHost string) string {
	if v := strings.TrimSpace(os.Getenv("TALIZEN_WEB_HOST")); v != "" {
		return v
	}

	u, err := url.Parse(apiHost)
	if err == nil {
		host := u.Hostname()
		if host == "localhost" || host == "127.0.0.1" {
			u.Host = "localhost:5173"
			u.Path = ""
			u.RawQuery = ""
			u.Fragment = ""
			return strings.TrimRight(u.String(), "/")
		}
	}

	return defaultWebHostValue
}

func Run(ctx context.Context, args []string) error {
	if len(args) == 0 {
		printUsage()
		return nil
	}

	switch args[0] {
	case "login":
		return runLogin(ctx, args[1:])
	case "projects":
		return runProjects(ctx, args[1:])
	case "pull":
		return runPull(ctx, args[1:])
	case "sync":
		return runSync(ctx, args[1:])
	case "preview":
		return runPreview(ctx, args[1:])
	case "publish":
		return runPublish(ctx, args[1:])
	case "version":
		fmt.Println(version)
		return nil
	case "help", "-h", "--help":
		printUsage()
		return nil
	default:
		return fmt.Errorf("unknown command: %s", args[0])
	}
}

func printUsage() {
	fmt.Println(`talizen cli

Usage:
  talizen login [--api=https://talizen.com] [--web=https://talizen.com]
  talizen projects
  talizen pull --site_id=<project_id>/<site_id> --dir=./mysite
  talizen sync --site_id=<project_id>/<site_id> --dir=./mysite
  talizen preview --site_id=<project_id>/<site_id>
  talizen publish [--commit=<commit>]
  talizen version`)
}

func clientFromConfig(apiHost string) (*talizen.Client, Config, error) {
	cfg, err := loadConfig()
	if err != nil {
		return nil, Config{}, err
	}
	if apiHost != "" {
		cfg.APIHost = apiHost
	}

	return talizen.NewClient(cfg.APIHost, cfg.Token), cfg, nil
}

func runLogin(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("login", flag.ContinueOnError)
	apiHost := fs.String("api", "", "Talizen API host")
	webHost := fs.String("web", "", "Talizen web host")
	err := fs.Parse(args)
	if err != nil {
		return err
	}

	client, cfg, err := clientFromConfig(*apiHost)
	if err != nil {
		return err
	}

	resolvedWebHost := strings.TrimSpace(*webHost)
	if resolvedWebHost == "" {
		resolvedWebHost = defaultWebHost(cfg.APIHost)
	}

	session, err := client.CreateCLIAuthSession(ctx, resolvedWebHost)
	if err != nil {
		return err
	}

	fmt.Printf("Open this URL to authorize Talizen CLI:\n%s\n", session.VerifyURL)
	_ = openBrowser(session.VerifyURL)

	deadline := time.Now().Add(time.Duration(session.ExpiresIn) * time.Second)
	for time.Now().Before(deadline) {
		time.Sleep(2 * time.Second)

		result, err := client.GetCLIAuthSession(ctx, session.Code)
		if err != nil {
			return err
		}
		if result.Status == "approved" {
			cfg.Token = result.Token
			err = saveConfig(cfg)
			if err != nil {
				return err
			}

			fmt.Println("Logged in.")
			return nil
		}
		if result.Status == "expired" {
			return fmt.Errorf("authorization expired")
		}
	}

	return fmt.Errorf("authorization timed out")
}

func runProjects(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("projects", flag.ContinueOnError)
	apiHost := fs.String("api", "", "Talizen API host")
	err := fs.Parse(args)
	if err != nil {
		return err
	}

	client, _, err := clientFromConfig(*apiHost)
	if err != nil {
		return err
	}

	projects, err := client.GetProjectList(ctx)
	if err != nil {
		return err
	}

	for _, project := range projects.List {
		fmt.Printf("%s\t%s\n", project.ID, project.Name)
		for _, site := range project.SiteList {
			fmt.Printf("  %s/%s\t%s\n", project.ID, site.ID, site.Name)
		}
	}

	return nil
}

func runPull(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("pull", flag.ContinueOnError)
	apiHost := fs.String("api", "", "Talizen API host")
	siteID := fs.String("site_id", "", "project_id/site_id")
	dir := fs.String("dir", ".", "local directory")
	err := fs.Parse(args)
	if err != nil {
		return err
	}

	projectID, realSiteID, err := parseSiteRef(*siteID)
	if err != nil {
		return err
	}

	client, _, err := clientFromConfig(*apiHost)
	if err != nil {
		return err
	}

	files, err := client.GetFileList(ctx, projectID, realSiteID)
	if err != nil {
		return err
	}

	err = writeRemoteFiles(*dir, files.List)
	if err != nil {
		return err
	}

	previewURL, _ := previewURL(ctx, client, realSiteID)
	fmt.Printf("Pulled %d files into %s\n", len(files.List), *dir)
	if previewURL != "" {
		fmt.Printf("Preview: %s\n", previewURL)
	}

	return nil
}

func runSync(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("sync", flag.ContinueOnError)
	apiHost := fs.String("api", "", "Talizen API host")
	siteID := fs.String("site_id", "", "project_id/site_id")
	dir := fs.String("dir", ".", "local directory")
	err := fs.Parse(args)
	if err != nil {
		return err
	}

	projectID, realSiteID, err := parseSiteRef(*siteID)
	if err != nil {
		return err
	}

	client, _, err := clientFromConfig(*apiHost)
	if err != nil {
		return err
	}

	syncer, err := NewSyncer(client, projectID, realSiteID, *dir)
	if err != nil {
		return err
	}

	previewURL, _ := previewURL(ctx, client, realSiteID)
	if previewURL != "" {
		fmt.Printf("Preview: %s\n", previewURL)
	}

	return syncer.Run(ctx)
}

func runPreview(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("preview", flag.ContinueOnError)
	apiHost := fs.String("api", "", "Talizen API host")
	siteID := fs.String("site_id", "", "project_id/site_id")
	err := fs.Parse(args)
	if err != nil {
		return err
	}

	_, realSiteID, err := parseSiteRef(*siteID)
	if err != nil {
		return err
	}

	client, _, err := clientFromConfig(*apiHost)
	if err != nil {
		return err
	}

	url, err := previewURL(ctx, client, realSiteID)
	if err != nil {
		return err
	}
	if url == "" {
		return fmt.Errorf("preview URL is unavailable")
	}

	fmt.Printf("Opening preview: %s\n", url)
	return openBrowser(url)
}

func runPublish(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("publish", flag.ContinueOnError)
	commit := fs.String("commit", "", "commit to tag")
	err := fs.Parse(args)
	if err != nil {
		return err
	}

	if fs.NArg() > 1 {
		return fmt.Errorf("publish accepts at most one commit argument")
	}
	if fs.NArg() == 1 {
		if strings.TrimSpace(*commit) != "" {
			return fmt.Errorf("commit specified twice")
		}
		*commit = fs.Arg(0)
	}

	tag, err := releaseTag(version)
	if err != nil {
		return err
	}

	target := strings.TrimSpace(*commit)
	if target == "" {
		target = "HEAD"
	}

	resolvedCommit, err := gitOutput(ctx, "rev-parse", "--verify", target+"^{commit}")
	if err != nil {
		return fmt.Errorf("resolve commit %q: %w", target, err)
	}

	err = gitRun(ctx, "rev-parse", "--is-inside-work-tree")
	if err != nil {
		return fmt.Errorf("publish must be run inside a git repository: %w", err)
	}

	err = gitRun(ctx, "rev-parse", "--verify", "refs/tags/"+tag)
	if err == nil {
		return fmt.Errorf("tag %s already exists locally", tag)
	}

	fmt.Printf("Publishing %s at %s\n", tag, resolvedCommit)
	if err := gitRun(ctx, "tag", tag, resolvedCommit); err != nil {
		return fmt.Errorf("create tag %s: %w", tag, err)
	}
	if err := gitRun(ctx, "push", "origin", tag); err != nil {
		return fmt.Errorf("push tag %s: %w", tag, err)
	}

	fmt.Printf("Published %s\n", tag)
	return nil
}

func releaseTag(rawVersion string) (string, error) {
	v := strings.TrimSpace(rawVersion)
	if v == "" || v == "dev" {
		return "", fmt.Errorf("cannot publish version %q", rawVersion)
	}
	if strings.HasPrefix(v, "v") {
		return v, nil
	}

	return "v" + v, nil
}

func gitRun(ctx context.Context, args ...string) error {
	_, err := gitOutput(ctx, args...)
	return err
}

func gitOutput(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg == "" {
			return "", err
		}
		return "", errors.New(msg)
	}

	return strings.TrimSpace(string(out)), nil
}

func parseSiteRef(ref string) (string, string, error) {
	parts := strings.Split(strings.TrimSpace(ref), "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("site_id must be <project_id>/<site_id>")
	}

	return parts[0], parts[1], nil
}

func previewURL(ctx context.Context, client *talizen.Client, siteID string) (string, error) {
	info, err := client.GetSystemInfo(ctx)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(info.SelfAPIHost) == "" {
		return "", nil
	}

	u, err := url.Parse(info.SelfAPIHost)
	if err != nil {
		return "", err
	}
	u.Host = siteID + ".preview." + u.Host
	u.Path = "/"
	u.RawQuery = ""
	u.Fragment = ""

	return u.String(), nil
}

func openBrowser(rawURL string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", rawURL)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", rawURL)
	default:
		cmd = exec.Command("xdg-open", rawURL)
	}

	return cmd.Start()
}
