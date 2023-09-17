import React, { useState } from 'react';
import { useQuery } from '@tanstack/react-query';

import Alert from 'react-bootstrap/Alert';
import Container from 'react-bootstrap/Container';

import Pager from '../components/Pager';
import Toasts from '../components/Toasts';
import Footer from '../components/Footer';
import Spinner from 'react-bootstrap/Spinner';
import EpistolaryNavbar from '../components/EpistolaryNavbar';

import { listReadings } from '../api';
import readingIcon from '../images/reading.png';
import CreateReadingForm from '../components/readings/CreateReadingForm';


const HomePage = () => {
  const [ page, setPage ] = useState("");
  const [ pagination, setPagination ] = useState({prevPageToken: "", nextPageToken: ""});
  const [ alerts, setAlerts ] = useState([]);

  const fetchReadings = async () => {
    try {
      const data = await listReadings(page);
      setPagination({
        prevPageToken: data.prev_page_token,
        nextPageToken: data.next_page_token,
      });
      return data.readings;
    } catch (err) {
      addAlert(err.message);
      throw err;
    }
  };

  const {
    isLoading,
    isError,
    error,
    data,
    isFetching,
    isPreviousData,
  } = useQuery({
    queryKey: ['readings', page],
    keepPreviousData: true,
    queryFn: fetchReadings,
  })

  const addAlert = (msg, bg='danger', header='') => {
    setAlerts(alerts => {
      const alert = {msg: msg, id: alerts.length+1, bg: bg, header: header};
      return [...alerts, alert];
    });
  }

  const renderReadings = (data) => {
    if (data) {
      return data.map(reading => {
        return (
          <li key={reading.id}>
            <img src={reading.favicon || readingIcon} width="16" height="16" alt="favicon" />{' '}
            <a className="mx-2" href={reading.link} target="_blank" rel="noreferrer">{reading.title || "unknown title"}</a>
          </li>
        );
      });
    }
  }

  return (
    <>
    <EpistolaryNavbar />
    <main className='flex-shrink-0' style={{paddingBottom: "96px"}}>
      <Toasts alerts={alerts} setAlerts={setAlerts} />
      <Container className="mt-4">
        <CreateReadingForm addAlert={addAlert} />
      </Container>
      <Container className="py-4">
        {isLoading ? (
          <div className="text-center">
            <Spinner animation="border" variant="primary" role="status">
              <span className="visually-hidden">Loading...</span>
            </Spinner>
          </div>
        ) : isError ? (
          <Alert variant="danger">
            <Alert.Heading>Something went wrong</Alert.Heading>
            <p>Could not load readings for this page. Received error <span className="text-muted">{error.message}</span></p>
          </Alert>
        ) : (
          <>
          <ul className='list-unstyled'>
            { renderReadings(data) }
          </ul>
          <Pager pagination={pagination} setPage={setPage} isPreviousData={isPreviousData} isFetching={isFetching} />
          </>
        )}
      </Container>
    </main>
    <Footer />
    </>
  );
}

export default HomePage;
