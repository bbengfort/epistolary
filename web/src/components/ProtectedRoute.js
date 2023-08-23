import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';

import { useAuth, isAnonymous, isExpired } from '../hooks/auth';

const ProtectedRoute = (props) => {
  const navigate = useNavigate();
  const [ isLoggedIn, setIsLoggedIn ] = useState(false);
  const [ authUser ] = useAuth();

  useEffect(() => {
    if (isAnonymous(authUser) || isExpired(authUser)) {
      setIsLoggedIn(false);
      return navigate('/login');
    }
    setIsLoggedIn(true);
  }, [isLoggedIn, authUser, navigate]);

  return (
    <>{ isLoggedIn ? props.children : null }</>
  )
}

export default ProtectedRoute;