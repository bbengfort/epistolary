import React, { useEffect, useState } from 'react';
import Container from 'react-bootstrap/Container';
import Footer from '../components/Footer';
import Form from 'react-bootstrap/Form';
import InputGroup from 'react-bootstrap/InputGroup';
import Button from 'react-bootstrap/Button';

import  { useForm }  from  "react-hook-form";
import { listReadings, createReading } from '../api';

function HomePage() {
  const { register, handleSubmit, formState:{errors} } = useForm();
  const [ readings, setReadings ] = useState([]);

  useEffect(() => {
    const fetchReadings = async () => {
      const response = await listReadings();
      console.log(response);
    }
    fetchReadings();
  }, []);

  const onSubmit = async (data) => {
    let response = await createReading(data.link);
    console.log(response);
  };

  return (
    <>
    <main className='flex-shrink-0'>
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
