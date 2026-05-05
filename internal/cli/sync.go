package cli

import (
	"bysir/talizen-cli/internal/talizen"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

type Syncer struct {
	client    *talizen.Client
	projectID string
	siteID    string
	dir       string
	clientID  string

	mu           sync.Mutex
	remoteByPath map[string]talizen.File
}

func NewSyncer(client *talizen.Client, projectID string, siteID string, dir string) (*Syncer, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("resolve sync dir: %w", err)
	}

	return &Syncer{
		client:       client,
		projectID:    projectID,
		siteID:       siteID,
		dir:          absDir,
		clientID:     newClientID(),
		remoteByPath: map[string]talizen.File{},
	}, nil
}

func newClientID() string {
	var b [16]byte
	_, err := rand.Read(b[:])
	if err != nil {
		return fmt.Sprintf("talizen-cli-%d", time.Now().UnixNano())
	}

	return "talizen-cli-" + hex.EncodeToString(b[:])
}

func (s *Syncer) Run(ctx context.Context) error {
	if err := s.Push(ctx); err != nil {
		return err
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("create watcher: %w", err)
	}
	defer watcher.Close()

	err = s.watchDirs(watcher)
	if err != nil {
		return err
	}

	fmt.Printf("Syncing %s -> %s/%s\n", s.dir, s.projectID, s.siteID)

	debounce := map[string]*time.Timer{}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-watcher.Errors:
			if err != nil {
				fmt.Printf("watch error: %v\n", err)
			}
		case event := <-watcher.Events:
			if shouldSkipLocalPath(s.dir, event.Name) {
				continue
			}
			if event.Op&fsnotify.Create != 0 {
				info, statErr := os.Stat(event.Name)
				if statErr == nil && info.IsDir() {
					_ = filepath.WalkDir(event.Name, func(path string, d os.DirEntry, err error) error {
						if err != nil || !d.IsDir() {
							return nil
						}
						if shouldSkipLocalPath(s.dir, path) {
							return filepath.SkipDir
						}
						return watcher.Add(path)
					})
				}
			}

			key := event.Name
			if timer, ok := debounce[key]; ok {
				timer.Stop()
			}
			debounce[key] = time.AfterFunc(400*time.Millisecond, func() {
				delete(debounce, key)
				if err := s.handleEvent(context.Background(), event); err != nil {
					fmt.Printf("sync %s: %v\n", event.Name, err)
				}
			})
		}
	}
}

func (s *Syncer) Push(ctx context.Context) error {
	err := os.MkdirAll(s.dir, 0o755)
	if err != nil {
		return fmt.Errorf("create local dir: %w", err)
	}

	err = s.refreshRemote(ctx)
	if err != nil {
		return err
	}

	return s.syncLocalSnapshot(ctx)
}

func (s *Syncer) refreshRemote(ctx context.Context) error {
	files, err := s.client.GetFileList(ctx, s.projectID, s.siteID)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.remoteByPath = make(map[string]talizen.File, len(files.List))
	for _, file := range files.List {
		if file.IsDir {
			continue
		}
		s.remoteByPath[file.Path] = file
	}

	return nil
}

func (s *Syncer) syncLocalSnapshot(ctx context.Context) error {
	return filepath.WalkDir(s.dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if shouldSkipLocalPath(s.dir, path) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			return nil
		}

		return s.upsertLocalFile(ctx, path)
	})
}

func (s *Syncer) watchDirs(watcher *fsnotify.Watcher) error {
	return filepath.WalkDir(s.dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if shouldSkipLocalPath(s.dir, path) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if !d.IsDir() {
			return nil
		}

		return watcher.Add(path)
	})
}

func (s *Syncer) handleEvent(ctx context.Context, event fsnotify.Event) error {
	if event.Op&(fsnotify.Remove|fsnotify.Rename) != 0 {
		remotePath, err := localPathToRemote(s.dir, event.Name)
		if err != nil {
			return err
		}

		return s.deleteRemotePath(ctx, remotePath)
	}
	if event.Op&(fsnotify.Write|fsnotify.Create) != 0 {
		info, err := os.Stat(event.Name)
		if err != nil || info.IsDir() {
			return nil
		}

		return s.upsertLocalFile(ctx, event.Name)
	}

	return nil
}

func (s *Syncer) upsertLocalFile(ctx context.Context, localPath string) error {
	remotePath, err := localPathToRemote(s.dir, localPath)
	if err != nil {
		return err
	}

	bodyBytes, err := os.ReadFile(localPath)
	if err != nil {
		return fmt.Errorf("read %s: %w", remotePath, err)
	}
	if !isUTF8FileBody(bodyBytes) {
		return nil
	}
	body := string(bodyBytes)

	s.mu.Lock()
	remote, exist := s.remoteByPath[remotePath]
	s.mu.Unlock()

	if exist && remote.Readonly {
		return nil
	}

	action := talizen.SiteActionChange{
		Action: "file_create",
		File: talizen.SiteActionFileSpec{
			Path: talizen.StringPtr(remotePath),
			Body: talizen.StringPtr(body),
		},
	}
	if exist {
		action.Action = "file_update"
		action.File = talizen.SiteActionFileSpec{
			ID:   remote.ID,
			Body: talizen.StringPtr(body),
		}
	}

	_, err = s.client.DoSiteAction(ctx, s.projectID, s.siteID, s.clientID, []talizen.SiteActionChange{action})
	if err != nil {
		return err
	}

	err = s.refreshRemote(ctx)
	if err != nil {
		return err
	}

	fmt.Printf("synced %s\n", remotePath)
	return nil
}

func (s *Syncer) deleteRemotePath(ctx context.Context, remotePath string) error {
	s.mu.Lock()
	_, exist := s.remoteByPath[remotePath]
	s.mu.Unlock()
	if !exist {
		return nil
	}

	_, err := s.client.DoSiteAction(ctx, s.projectID, s.siteID, s.clientID, []talizen.SiteActionChange{
		{
			Action: "file_delete",
			File: talizen.SiteActionFileSpec{
				Path: talizen.StringPtr(remotePath),
			},
		},
	})
	if err != nil {
		return err
	}

	err = s.refreshRemote(ctx)
	if err != nil {
		return err
	}

	fmt.Printf("deleted %s\n", remotePath)
	return nil
}
