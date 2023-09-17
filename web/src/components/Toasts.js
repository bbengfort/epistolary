import React from 'react';

import Toast from 'react-bootstrap/Toast';
import ToastContainer from 'react-bootstrap/ToastContainer';

const Toasts = ({ alerts, setAlerts }) => {
   const handleClose = (id) => {
    const filtered = alerts.filter((alert) => alert.id !== id);
    setAlerts(filtered);
  }

  const alertsList = alerts.map((alert) => (
    <Toast
      autohide
      animation
      delay={5000}
      bg={alert.bg || 'danger'}
      key={alert.id}
      onClose={() => handleClose(alert.id)}
    >
      <Toast.Header>
        <strong className="me-auto">{ alert.header || "An Error Occurred" }</strong>
      </Toast.Header>
      <Toast.Body>
        { alert.msg }
      </Toast.Body>
    </Toast>
  ));

  return (
    <ToastContainer position='top-end' className="mt-2 mx-2">
      { alertsList }
    </ToastContainer>
  );
}

export default Toasts;