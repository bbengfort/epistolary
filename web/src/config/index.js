
const config = {
  environment: process.env.NODE_ENV,
  analyticsID: process.env.REACT_APP_ANALYTICS_ID,
  apiBaseURL: process.env.REACT_APP_API_BASE_URL,
  appInfo: {
    version: process.env.REACT_APP_VERSION_NUMBER,
    revision: process.env.REACT_APP_GIT_REVISION,
  },
  sentry: {
    dsn: process.env.REACT_APP_SENTRY_DSN,
    environment: process.env.REACT_APP_SENTRY_ENVIRONMENT,
  },
  i18n: {
    useDashLocale: process.env.REACT_APP_USE_DASH_LOCALE,
  },
}

export const isProduction = () => {
  return config.environment === 'production';
}

export default config;