import React from 'react';
import Button from 'react-bootstrap/Button';


const Pager = ({ pagination, setPage, isPreviousData, isFetching }) => {
  return (
    <div className="d-grid gap-2 d-flex justify-content-around mt-5">
      <Button
        type="button"
        className='btn btn-primary'
        disabled={(!isPreviousData && !pagination.prevPageToken) || isFetching}
        onClick={() => setPage(pagination.prevPageToken)}
      >
        &laquo; Prev
      </Button>
      <Button
        type="button"
        className='btn btn-primary'
        disabled={!pagination.nextPageToken || isFetching}
        onClick={() => setPage(pagination.nextPageToken)}
      >
        Next &raquo;
      </Button>
    </div>
  );
}

export default Pager;