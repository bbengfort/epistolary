import React from 'react';
import { Button, Nav } from "react-bootstrap";
import Container from 'react-bootstrap/Container';
import Navbar from 'react-bootstrap/Navbar';

import logo from '../images/logo.png'
import { logout } from '../api';
import { useNavigate } from 'react-router-dom';
import { useAuth, ANONYMOUS } from '../hooks/auth';


const EpistolaryNavbar = () => {
  const navigate = useNavigate();
  const [ , setAuthUser ] = useAuth();

  const logoutHandler = async () => {
    await logout();
    setAuthUser(ANONYMOUS);
    navigate('/login');
  }

  return (
    <Navbar bg="dark" expand="lg" className="navbar-dark">
      <Container>
        <Navbar.Brand>
          <img
            alt=""
            src={logo}
            width="30"
            height="30"
            className='d-inline-block align-top'
          />{' '}
          Epistolary
        </Navbar.Brand>
        <Navbar.Toggle aria-controls="navbar-nav" />
        <Navbar.Collapse id="navbar--nav">
          <Nav className="ms-auto">
            <Nav.Link>
              <Button className="btn-warning" onClick={logoutHandler}>Logout</Button>
            </Nav.Link>
          </Nav>
        </Navbar.Collapse>
      </Container>
    </Navbar>
  );
}

export default EpistolaryNavbar;