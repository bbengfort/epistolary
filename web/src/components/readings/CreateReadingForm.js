import React from 'react';
import { useForm }  from  "react-hook-form";
import { useMutation, useQueryClient } from '@tanstack/react-query';

import Form from 'react-bootstrap/Form';
import Button from 'react-bootstrap/Button';
import Spinner from 'react-bootstrap/Spinner';
import InputGroup from 'react-bootstrap/InputGroup';

import { createReading } from '../../api';


const CreateReadingForm = ({ addAlert }) => {
  // Form state
  const { register, handleSubmit, reset, formState:{errors} } = useForm();

  // Query mutation
  const queryClient = useQueryClient();
  const mutation = useMutation({
    mutationFn: createReading,
    onError: error => addAlert("could not create reading: " + error.message),
    onSuccess: result => {
      queryClient.invalidateQueries({ queryKey: ['readings'] });
      addAlert(result.title + " was successfully created", "success", "Reading Created");
    },
    onSettled: () => reset(),
  });

  const onSubmit = async (data) => {
    if (errors.link) {
      addAlert(errors.link);
      return
    }
    mutation.mutate(data.link);
  };

  return (
    <Form onSubmit={handleSubmit(onSubmit)}>
      <InputGroup>
        <Form.Control
          type="url"
          placeholder="Insert URL to add to readings ..."
          autoComplete="link"
          {...register("link", { required: true })}
        />
        <Button type="submit" variant="outline-dark" disabled={mutation.isLoading}>
          {mutation.isLoading && (
            <>
            <Spinner
              as="span"
              animation="border"
              size="sm"
              role="status"
              aria-hidden="true"
              variant="dark"
            />
            <span className="visually-hidden">Loading...</span>
            </>
          )}
          {' '}Add
        </Button>
      </InputGroup>
    </Form>
  );
};

export default CreateReadingForm;