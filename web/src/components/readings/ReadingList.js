import React from 'react';

import Table from 'react-bootstrap/Table';
import Button from 'react-bootstrap/Button';


import readingIcon from '../../images/reading.png';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faPenToSquare, faCircleCheck, faBookOpenReader, faHourglassStart } from '@fortawesome/free-solid-svg-icons'


const ReadingList = ({ readings, setReadingDetail }) => {

  const statusButton = status => {
    switch (status) {
      case "finished":
        return (
          <Button disabled variant="outline-danger" size="sm" className='me-2'>
            <FontAwesomeIcon icon={faCircleCheck} />
          </Button>
        );
      case "started":
        return (
          <Button disabled variant="outline-success" size="sm" className='me-2'>
            <FontAwesomeIcon icon={faBookOpenReader} />
          </Button>
        );
      case "queued":
        return (
          <Button disabled variant="outline-light" size="sm" className='me-2'>
            <FontAwesomeIcon icon={faHourglassStart} />
          </Button>
        );
      default:
        return null;
    }
  }

  const renderReadings = () => {
    if (readings) {
      return readings.map(reading => {
        return (
          <tr key={reading.id}>
            <td className='text-center'>
              <img src={reading.favicon || readingIcon} width="28" height="28" alt="favicon" />
            </td>
            <td className='align-middle fs-6'>
              <a className='text-decoration-none' href={reading.link} target="_blank" rel="noreferrer">{reading.title || "unknown title"}</a>
            </td>
            <td>
              {statusButton(reading.status)}
              <Button variant="outline-primary" size="sm" onClick={() => setReadingDetail(reading.id)}>
                <FontAwesomeIcon icon={faPenToSquare} />
              </Button>
            </td>
          </tr>
        );
      });
    }
  }

  return (
    <Table hover responsive>
      <tbody>
        { renderReadings() }
      </tbody>
    </Table>
  );
}

export default ReadingList;