import React, { useEffect } from 'react';
import ReactGA from 'react-ga4';
import { useLocation } from "react-router-dom";

import config, { isProduction } from '../config';

const useAnalytics = () => {
  const [isInitialized, setIsInitialized] = React.useState(false);
  const analyticsID = config.analyticsID;

  useEffect(() => {
    // initialize google analytics only in production environment
    if (analyticsID && isProduction()) {
      // eslint-disable-next-line no-console
      console.debug("initializing google analytics");
      ReactGA.initialize(analyticsID, {
        gaOptions: {
          siteSpeedSampleRate: 100,
        },
      });
    }
    setIsInitialized(true);
  }, [analyticsID]);

  return {
    isInitialized,
  };
};

const usePageTracking = () => {
  const { isInitialized } = useAnalytics();
  const location = useLocation();

  useEffect(() => {
    if (isInitialized) {
      ReactGA.send({ hitType: "pageview", page: location.pathname + location.search });

    }
  }, [isInitialized, location])

}

export default usePageTracking;
