import React, { useEffect, useState } from 'react';
import Container from 'react-bootstrap/Container';
import Footer from '../components/Footer';
import Form from 'react-bootstrap/Form';
import InputGroup from 'react-bootstrap/InputGroup';
import Button from 'react-bootstrap/Button';
import Toasts from '../components/Toasts';
import EpistolaryNavbar from '../components/EpistolaryNavbar';

import  { useForm }  from  "react-hook-form";
import { listReadings, createReading } from '../api';
import readingIcon from '../images/reading.png';

function HomePage() {
  const { register, handleSubmit, formState:{errors} } = useForm();
  const [ readings, setReadings ] = useState([]);
  const [ alerts, setAlerts ] = useState([]);

  useEffect(() => {
    const fetchReadings = async () => {
      const response = await listReadings();
      if (response.error) {
        addAlert(response.error);
      } else {
        setReadings(response.readings);
      }
    }
    fetchReadings();
  }, []);

  const onSubmit = async (data) => {
    if (errors.link) {
      addAlert(errors.link);
      return
    }

    let response = await createReading(data.link);
    if (response.error) {
      addAlert(response.error);
    } else {
      setReadings(readings => {
        return [...readings, response];
      })
    }
  };

  const addAlert = (msg) => {
    setAlerts(alerts => {
      const alert = {msg: msg, id: alerts.length+1, bg: 'danger'};
      return [...alerts, alert];
    });
  }

  const renderReadings = () => {
    return readings.map(reading => {
      return (
        <li key={reading.id}>
          <img src={reading.favicon || readingIcon} width="16" height="16" alt="favicon" />
          <a className="mx-2" href={reading.link} target="_blank" rel="noreferrer">{reading.title || "unknown title"}</a>
        </li>
      );
    });
  }

  return (
    <>
    <EpistolaryNavbar />
    <main className='flex-shrink-0' style={{paddingBottom: "96px"}}>
      <Toasts alerts={alerts} setAlerts={setAlerts} />
      <Container className="mt-4">
        <Form onSubmit={handleSubmit(onSubmit)}>
          <InputGroup>
            <Form.Control
              type="url"
              placeholder="Insert URL to add to readings ..."
              autoComplete="link"
              {...register("link", { required: true })}
            />
            <Button type="submit" variant="outline-secondary">
              Add
            </Button>
          </InputGroup>
        </Form>
      </Container>
      <Container className="py-4">
        <ul className='list-unstyled'>
          { renderReadings() }
        </ul>
      </Container>
    </main>
    <Footer />
    </>
  );
}

export default HomePage;
