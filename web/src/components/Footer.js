import { React, useEffect, useState } from 'react';
import Container from 'react-bootstrap/Container';
import API from '../api';
import config from '../config';

const Footer = () => {
  const [ apiStatus, setAPIStatus ] = useState({'status': '', 'uptime': '', 'version': ''});

  useEffect(() => {
    const fetchStatus = async () => {
      try {
        const response = await API.get('status');
        setAPIStatus(response.data)
      } catch (error) {
        if (error.response) {
          // Handle maintenance mode
          if (error.response.status === 503) {
            setAPIStatus(error.response.data);
            return
          }
          console.error("received api error", error.response.status, error.response.data);
        } else {
          console.error("could not connect to status endpoint", error.message);
        }
        setAPIStatus({'status': 'offline', 'uptime': '', 'version': ''});
      }
    }
    fetchStatus();
  }, []);

  const renderVersionInfo = () => {
    const version = config.appInfo.version;
    const revision = config.appInfo.revision;

    if (version && revision) {
      return `${version} (${revision})`;
    }

    if (version) {
      return version;
    }

    if (revision) {
      return 'revision ' + revision;
    }

    else {
      return 'version unknown';
    }
  }

  const renderAPIStatus = () => {
    switch (apiStatus.status) {
      case 'ok':
       return <span className="text-success">◉</span>;
      case 'offline':
        return <span className="text-danger">◉</span>;
      case 'stopping':
        return <span className="text-info">◉</span>;
      case 'maintenance':
        return <span className="text-warning">◉</span>;
      default:
        return <span className="text-muted">◉</span>;
    }
  }

  return (
    <footer className="footer mt-auto py-4 bg-body-tertiary fixed-bottom">
      <Container className='text-center'>
        <p className="text-body-secondary mb-0 pb-0">
          Made with &spades; by <a href="https://github.com/bbengfort">@bbengfort</a>
        </p>
        <ul className='list-unstyled list-inline text-muted pb-0 mb-0'>
          <li className='list-inline-item'>
            <small>Epistolary UI { renderVersionInfo() }</small>
          </li>
          <li className='list-inline-item'>
            <small>{ renderAPIStatus() }</small>
          </li>
          <li className='list-inline-item'>
            <small>API {  apiStatus.version ? `v${apiStatus.version}` : 'version unknown' }</small>
          </li>
        </ul>
      </Container>
    </footer>
  );
}

export default Footer;