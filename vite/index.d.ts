import type { Plugin } from 'vite'

export type TalizenVitePluginOptions = {
  /**
   * Local Talizen project root. Defaults to Vite's root.
   */
  root?: string
  /**
   * Talizen API host used by talizen/cms and talizen/form at runtime.
   */
  apiHost?: string
  /**
   * Project id for runtime CMS/form requests.
   */
  projectId?: string
  /**
   * Optional CLI auth token. When set, local /api proxy sends it as Bearer.
   */
  token?: string
  /**
   * Platform import map entries. `talizen dev` fills this from server system info.
   * Manually configured entries override the built-in fallback when no server map is provided.
   */
  importMap?: Record<string, string>
}

export declare function talizen(options?: TalizenVitePluginOptions): Plugin
export default talizen
