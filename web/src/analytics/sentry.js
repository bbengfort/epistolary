import * as Sentry from '@sentry/react';
import config from '../config';

const initSentry = () => {
  // Ensure revision and version of application is set as is sentry DSN
  if (!config.sentry.dsn) {
    // eslint-disable-next-line no-console
    console.warn("sentry is not configured")
    return
  }

  // Set the environment for sentry monitoring
  const environment = config.sentry.environment ? config.sentry.environment : config.environment;

  Sentry.init({
    dsn: config.sentry.dsn,
    integrations: [
      new Sentry.BrowserTracing({
        // Set 'tracePropagationTargets' to control for which URLs distributed tracing should be enabled
        tracePropagationTargets: ["localhost", config.apiBaseURL],
      }),
      new Sentry.Replay(),
    ],

    // Performance Monitoring
    tracesSampleRate: 0.25,         // Do not capture 100% of the transactions in production!

    environment,
    release: config.appInfo.version,
    ignoreErrors: ['Session expired'],

    // Session Replay
    replaysSessionSampleRate: 0.1, // This sets the sample rate at 10%. You may want to change it to 100% while in development and then sample at a lower rate in production.
    replaysOnErrorSampleRate: 1.0, // If you're not already sampling the entire session, change the sample rate to 100% when sampling sessions where errors occur.
  });

  // eslint-disable-next-line no-console
  console.debug("sentry configured and loaded")
}

export default initSentry;