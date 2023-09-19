import React from 'react';
import { useForm }  from  "react-hook-form";
import dayjs  from 'dayjs';
import relativeTime from 'dayjs/plugin/relativeTime';

import Col from 'react-bootstrap/Col';
import Row from 'react-bootstrap/Row';
import Form from 'react-bootstrap/Form';

// Extend dayjs with relative time.
dayjs.extend(relativeTime);


const ReadingForm = ({ reading }) => {
  // Form State
  const { register, handleSubmit, formState:{errors} } = useForm({
    defaultValues: reading,
  });

  const onSubmit = async (data) => {
    if (errors) {
      console.error(errors);
      return
    }
    console.info(data);
  }

  return (
    <Form onSubmit={handleSubmit(onSubmit)}>
      <Form.Group className="mb-3" controlId='formReadingTitle'>
        <Form.Label>Title</Form.Label>
        <Form.Control
          type='text'
          placeholder='Title'
          {...register("title", { required: true })}
        />
        <Form.Text>
          Status: {reading.status}.
        </Form.Text>
      </Form.Group>

      <Form.Group className="mb-3" controlId='formReadingDescription'>
        <Form.Label>Description</Form.Label>
        <Form.Control
          as="textarea"
          rows={4}
          {...register("description")}
        />
      </Form.Group>

      <Form.Group className="mb-3" controlId='formReadingLink'>
        <Form.Label>Link</Form.Label>
        <Form.Control
          type='url'
          readOnly
          disabled
          {...register("link", { required: true })}
        />
      </Form.Group>

      <Row className='mb-3'>
        <Form.Group as={Col} controlId="formReadingStarted">
          <Form.Label>Started</Form.Label>
          <Form.Control
            type="datetime-local"
            placeholder="Started"
            {...register("started")}
          />
        </Form.Group>
        <Form.Group as={Col} controlId="formReadingFinished">
          <Form.Label>Finished</Form.Label>
          <Form.Control
            type="datetime-local"
            placeholder="Finished"
            {...register("finished")}
          />
        </Form.Group>
      </Row>
      <Form.Text>
        Created {dayjs(reading.created).fromNow()}, last modified {dayjs(reading.modified).fromNow()}.
      </Form.Text>
    </Form>
  );
}

export default ReadingForm;