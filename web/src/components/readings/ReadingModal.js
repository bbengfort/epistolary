import React from 'react';
import { useQuery } from '@tanstack/react-query';

import Alert from 'react-bootstrap/Alert';
import Modal from 'react-bootstrap/Modal';
import Button from 'react-bootstrap/Button';
import Spinner from 'react-bootstrap/Spinner';

import { fetchReading } from '../../api';

import readingIcon from '../../images/reading.png';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faTrashCan, faSave } from '@fortawesome/free-solid-svg-icons'


const ReadingModal = ({ readingID, show, setShow }) => {

  const fetchReadingDetail = async() => {
      if (readingID) {
        return await fetchReading(readingID);
      }
      return null;
  }

  const {
    isLoading,
    isError,
    error,
    data,
    isFetching,
  } = useQuery({
    queryKey: ['reading', readingID],
    keepPreviousData: true,
    queryFn: fetchReadingDetail,
    retry: false,
  })

  const modalTitle = () => {
    if (isLoading || isFetching) {
      return "Loading ..."
    }

    if (isError) {
      return "Something went wrong"
    }

    return (
      <>
      <img src={data.favicon || readingIcon} width="24" height="24" alt="favicon" className='my-0 py-0' />{' '}
      Reading Detail ({data.id})
      </>
    )
  }

  return (
    <Modal
      show={show}
      onHide={() => setShow(false)}
      backdrop="static"
      size="lg"
      aria-labelledby="reading-detail-modal-title"
      centered
    >
      <Modal.Header closeButton>
        <Modal.Title id="reading-detail-modal-title">
          {modalTitle()}
        </Modal.Title>
      </Modal.Header>
      <Modal.Body>
        {(isLoading || isFetching) ? (
          <div className="text-center">
            <Spinner animation="border" variant="primary" role="status">
              <span className="visually-hidden">Loading...</span>
            </Spinner>
          </div>
        ) : isError ? (
          <Alert variant='danger'>
            Could not fetch reading detail: { error.message }
          </Alert>
        ) : (
          <>
          <h6>{data.title}</h6>
          <p>{data.description}</p>
          <p>Status: {data.status}</p>
          <p>Started: {data.started} Finished: {data.finished}</p>
          <a href={data.link} target="_blank" rel="noreferrer">link</a>
          </>
        )}
      </Modal.Body>
      <Modal.Footer>
        {(!isLoading && !isFetching && !isError) &&
          <>
          <Button variant="danger" disabled>
            <FontAwesomeIcon icon={faTrashCan} />{' '}
            Delete
          </Button>
          <Button variant="primary" disabled>
            <FontAwesomeIcon icon={faSave} />{' '}
            Save
          </Button>
          </>
        }
        <Button variant="secondary" onClick={() => setShow(false)}>
          Close
        </Button>
      </Modal.Footer>
    </Modal>
  );
}

export default ReadingModal;