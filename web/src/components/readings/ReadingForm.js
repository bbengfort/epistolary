import React from 'react';
import { useForm }  from  "react-hook-form";
import { useMutation, useQueryClient } from '@tanstack/react-query';

import dayjs  from 'dayjs';
import relativeTime from 'dayjs/plugin/relativeTime';
import timezone from 'dayjs/plugin/timezone';

import Col from 'react-bootstrap/Col';
import Row from 'react-bootstrap/Row';
import Form from 'react-bootstrap/Form';

import { updateReading } from '../../api';


// Extend dayjs with relative time.
dayjs.extend(relativeTime);
dayjs.extend(timezone);
dayjs.tz.setDefault(dayjs.tz.guess());


const formatReading = reading => {
  if (reading.started) {
    reading.started = dayjs(reading.started).format("YYYY-MM-DDTHH:mm");
  }

  if (reading.finished) {
    reading.finished = dayjs(reading.finished).format("YYYY-MM-DDTHH:mm");
  }

  return reading;
}


const ReadingForm = ({ reading, addAlert }) => {
  // Form State
  const { register, handleSubmit } = useForm({
    defaultValues: formatReading(reading),
  });

  // Query mutation
  const queryClient = useQueryClient();
  const mutation = useMutation({
    mutationFn: updateReading,
    onError: error => addAlert("Could not update reading: " + error.message),
    onSuccess: updated => {
      queryClient.setQueryData(['reading', updated.id], updated);
      addAlert(updated.title + " was successfully updated", "success", "Reading Updated");
    }
  })


  const onSubmit = async (data) => {
    if (data.started) {
      data.started = dayjs(data.started).format();
    }

    if (data.finished) {
      data.finished = dayjs(data.finished).format();
    }

    mutation.mutate(data);
  }

  return (
    <Form id="formReading" onSubmit={handleSubmit(onSubmit)}>
      <Form.Group className="mb-3" controlId='formReadingTitle'>
        <Form.Label>Title</Form.Label>
        <Form.Control
          type='text'
          placeholder='Title'
          {...register("title", { required: true })}
        />
        <Form.Text>
          Status: {reading.status}
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