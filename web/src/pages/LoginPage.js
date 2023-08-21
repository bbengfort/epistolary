import React, { useState } from 'react';
import Form from 'react-bootstrap/Form';
import Button from 'react-bootstrap/Button';
import Alert from 'react-bootstrap/Alert';
import { useNavigate } from "react-router-dom";

import  { useForm }  from  "react-hook-form";
import { useBodyClass } from '../hooks';

import { login } from '../api';
import logo from '../images/logo.png';
import './LoginPage.css';

function LoginPage() {
  useBodyClass(["bg-body-tertiary"])
  const navigate = useNavigate();
  const { register, handleSubmit, formState:{errors} } = useForm();
  const [ alerts, setAlerts ] = useState([]);

  const onSubmit = async (data) => {
    if (errors.username || errors.password) {
      alerts.push("no username or password");
      setAlerts(alerts);
      return
    }

    let response = await login(data.username, data.password);
    console.log(response);
    if (response.error) {
      alerts.push(response.error);
      setAlerts(alerts);
    } else {
      navigate("/");
    }
  }

  const removeAlert = (i) => {
    alerts.splice(i, 1);
    setAlerts(alerts);
  }

  const renderAlerts = () => {
    return alerts.map((msg, i) => {
      return <Alert key={i} variant="danger" onClose={() => removeAlert(i)} dismissible><p className="p-0 m-0">{msg}</p></Alert>
    })
  }

  return (
    <div className='full-height align-items-center py-5 w-100 h-100'>
      <main className='form-signin w-100 m-auto mt-5'>
        <div className="text-center">
          <img src={logo} alt="Logo" className='mb-4' width="72" height="72" />
          <h1 className="h3 mb-3 fw-normal">Please sign in</h1>
        </div>
        <div className="alerts">{ renderAlerts() }</div>
        <Form onSubmit={handleSubmit(onSubmit)}>
            <div className="form-floating">
              <Form.Control
                type="text"
                placeholder="username"
                autoComplete="username"
                {...register("username", { required: true })}
              />
              <Form.Label>Username</Form.Label>
            </div>
            <div className="form-floating">
              <Form.Control
                type="password"
                placeholder="password"
                autoComplete="current-password"
                {...register("password", { required: true })}
              />
              <Form.Label>Password</Form.Label>
            </div>
            <div className="form-check text-start my-3">
              <Form.Check type="checkbox" />
              <Form.Label>Remember Me</Form.Label>
            </div>
            <Button type="submit" className="btn btn-primary w-100 py-2">
              Submit
            </Button>
        </Form>
      </main>
    </div>
  );
}

export default LoginPage;
