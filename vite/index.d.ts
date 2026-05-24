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
   * Extra or overriding import map entries.
   */
  importMap?: Record<string, string>
}

export declare function talizen(options?: TalizenVitePluginOptions): Plugin
export default talizen
