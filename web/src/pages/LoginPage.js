import React, { useState } from 'react';
import Form from 'react-bootstrap/Form';
import Button from 'react-bootstrap/Button';
import Alerts from '../components/Alerts';
import { useNavigate } from "react-router-dom";

import jwt from 'jwt-decode';
import  { useForm }  from  "react-hook-form";
import useBodyClass from '../hooks/body';
import useAuth from '../hooks/auth';

import { login } from '../api';
import logo from '../images/logo.png';
import './LoginPage.css';

function LoginPage() {
  useBodyClass(["bg-body-tertiary"])
  const navigate = useNavigate();
  const { register, handleSubmit, formState:{errors} } = useForm();
  const [ alerts, setAlerts ] = useState([]);
  const [ , setAuth ] = useAuth();

  const onSubmit = async (data) => {
    if (errors.username || errors.password) {
      addAlert("no username or password");
      return
    }

    let response = await login(data.username, data.password);
    if (response.error) {
      addAlert(response.error);
    } else {
      // Parse the JWT token
      if (response.access_token) {
        const user = jwt(response.access_token);
        setAuth(user);
        navigate("/");
      } else {
        addAlert("could not parse login response")
      }
    }
  }

  const addAlert = (msg) => {
    const alert = {msg: msg, id: alerts.length+1, variant: 'danger'};
    setAlerts(alerts => {
      return [...alerts, alert];
    });
  }

  return (
    <div className='full-height align-items-center py-5 w-100 h-100'>
      <main className='form-signin w-100 m-auto mt-5'>
        <div className="text-center">
          <img src={logo} alt="Logo" className='mb-4' width="72" height="72" />
          <h1 className="h3 mb-3 fw-normal">Please sign in</h1>
        </div>
        <Alerts alerts={alerts} setAlerts={setAlerts} />
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
