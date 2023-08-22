import React from 'react';
import Alert from 'react-bootstrap/Alert';

const Alerts = ({ alerts, setAlerts }) => {
  const handleClose = (id) => {
    const filtered = alerts.filter((alert) => alert.id !== id);
    setAlerts(filtered);
  }

  const alertsList = alerts.map(({ id, variant, msg }) => (
    <Alert
      key={id}
      variant={variant}
      dismissible
      onClose={() => handleClose(id)}
    >
      {msg}
    </Alert>
  ));

  return (
    <div className="alerts">
      { alertsList }
    </div>
  );
}

export default Alerts;