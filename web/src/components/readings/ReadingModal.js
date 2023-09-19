import React from 'react';
import { useQuery } from '@tanstack/react-query';

import Alert from 'react-bootstrap/Alert';
import Modal from 'react-bootstrap/Modal';
import Button from 'react-bootstrap/Button';
import Spinner from 'react-bootstrap/Spinner';
import ReadingForm from './ReadingForm';

import { fetchReading } from '../../api';

import readingIcon from '../../images/reading.png';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faSave } from '@fortawesome/free-solid-svg-icons'

const ReadingModal = ({ readingID, show, setShow, addAlert }) => {

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
      return "Loading ...";
    }

    if (isError) {
      return "Something went wrong";
    }

    if (!data) {
      return <></>;
    }

    return (
      <h5 className='mt-2'>
        <img src={data.favicon || readingIcon} width="20" height="20" alt="favicon" style={{marginTop: "-4px"}} />{' '}
        Reading Detail <span className="text-muted">({data.id})</span>
      </h5>
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
        <Modal.Title as="div" id="reading-detail-modal-title">
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
        ) : data && (
          <ReadingForm reading={data} addAlert={addAlert} />
        )}
      </Modal.Body>
      <Modal.Footer>
        {(!isLoading && !isFetching && !isError) &&
          <>
          <Button variant="primary" form="formReading" type="submit">
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