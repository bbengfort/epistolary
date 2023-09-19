import React from 'react';

import Table from 'react-bootstrap/Table';
import Button from 'react-bootstrap/Button';


import readingIcon from '../../images/reading.png';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faPenToSquare } from '@fortawesome/free-solid-svg-icons'


const ReadingList = ({ readings, setReadingDetail }) => {

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