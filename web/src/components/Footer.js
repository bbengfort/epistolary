import React from 'react';
import Container from 'react-bootstrap/Container';

const Footer = () => {
  return (
    <footer className="footer mt-auto py-3 bg-body-tertiary fixed-bottom">
      <Container className='text-center'>
        <span className="text-body-secondary">
          Made with &spades; by <a href="https://github.com/bbengfort">@bbengfort</a>
        </span>
      </Container>
    </footer>
  );
}

export default Footer;