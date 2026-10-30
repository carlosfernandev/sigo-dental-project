interface ImportMetaEnv {
  readonly DEV: boolean;
  readonly VITE_API_BASE_URL: string;
  readonly VITE_LOCAL_LOGIN_EMAIL?: string;
  readonly VITE_LOCAL_LOGIN_PASSWORD?: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
