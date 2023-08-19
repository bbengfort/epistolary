import './index.css';

import React from 'react';
import ReactDOM from 'react-dom/client';
import { BrowserRouter } from 'react-router-dom';

import App from './App';
import config from './config';
import initSentry from './analytics/sentry';

console.info("initializing epistolary interface", config.environment, config.appInfo.version, config.appInfo.revision);
initSentry();

const root = ReactDOM.createRoot(document.getElementById('root'));

root.render(
  <React.StrictMode>
    <BrowserRouter basename='/'>
      <App />
    </BrowserRouter>
  </React.StrictMode>
);