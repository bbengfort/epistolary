import React, { useEffect, useState } from 'react';
import Container from 'react-bootstrap/Container';
import Footer from '../components/Footer';
import Form from 'react-bootstrap/Form';
import InputGroup from 'react-bootstrap/InputGroup';
import Button from 'react-bootstrap/Button';
import Toast from 'react-bootstrap/Toast';
import ToastContainer from 'react-bootstrap/ToastContainer';

import  { useForm }  from  "react-hook-form";
import { listReadings, createReading } from '../api';
import { ToastHeader } from 'react-bootstrap';

function HomePage() {
  const { register, handleSubmit, formState:{errors} } = useForm();
  const [ readings, setReadings ] = useState([]);
  const [ alerts, setAlerts ] = useState([]);

  useEffect(() => {
    const fetchReadings = async () => {
      const response = await listReadings();
      if (response.error) {
        setAlerts(alerts => {
          return [...alerts, response.error];
        });
      } else {
        setReadings(response.readings);
      }
    }
    fetchReadings();
  }, []);

  const onSubmit = async (data) => {
    if (errors.link) {
      setAlerts(alerts => {
        return [...alerts, errors.link];
      });
      return
    }

    let response = await createReading(data.link);
    if (response.error) {
      setAlerts(alerts => {
        return [...alerts, response.error];
      });
    } else {
      setReadings(readings => {
        return [...readings, response];
      })
    }
  };

  const removeAlert = (idx) => {
    setAlerts(alerts => {
      return alerts.splice(idx, 1);
    });
  }

  const renderAlerts = () => {
    return alerts.map((msg, i) => {
      return (
        <Toast autohide delay={3000} bg={'danger'} key={i} onClose={() => removeAlert(i)}>
          <ToastHeader>
            <strong className="me-auto">An Error Occurred</strong>
          </ToastHeader>
          <Toast.Body>
            { msg }
          </Toast.Body>
        </Toast>
      );
    });
  }

  return (
    <>
    <main className='flex-shrink-0'>
      <ToastContainer position='top-end' className="mt-2 mx-2">
        { renderAlerts() }
      </ToastContainer>
      <Container className="my-4">
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
    </main>
    <Footer />
    </>
  );
}

export default HomePage;
