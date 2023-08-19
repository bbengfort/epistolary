import React from 'react';
import Form from 'react-bootstrap/Form';
import { useBodyClass } from '../hooks';

function LoginPage() {
  useBodyClass(["d-flex", "align-items-center", "py-4", "bg-body-tertiary"])

  return (
    <main className='form-signin w-100 m-auto'>
      <Form>
        <h1 className="h3 mb-3 fw-normal">Please sign in</h1>
      </Form>
    </main>
  );
}

export default LoginPage;
