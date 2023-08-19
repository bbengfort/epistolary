import React, { useState } from 'react';
import Container from 'react-bootstrap/Container';
import Footer from '../components/Footer';

function HomePage() {
  return (
    <>
    <main className='flex-shrink-0'>
      <Container>
        <h1 className='mt-5'>Epistolary</h1>
      </Container>
    </main>
    <Footer />
    </>
  );
}

export default HomePage;
