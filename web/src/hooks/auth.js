import { useEffect } from "react";
import useSessionStorage from "./session";

const AUTH_SESSION_KEY = "epistolaryAuthUser";
const ANONYMOUS = {username: "anonymous", exp: 0}

export const isAnonymous = user => {
  if (user) {
    return user.username === ANONYMOUS && user.exp <= 0;
  }
  return true;
}

export const isExpired = user => {
  if (user) {
    if (Date.now() < (user.exp * 1000)) {
      return false;
    }
  }
  return true;
}

export const useAuth = () => {
  const [ authUser, setAuthUserSession ] = useSessionStorage(AUTH_SESSION_KEY, ANONYMOUS);

  // Do not set a user whose timestamp is expired
  const setAuthUser = user => {
    if (!isAnonymous(user) && !isExpired(user)) {
      setAuthUserSession(user);
    } else {
      setAuthUserSession(ANONYMOUS);
    }
  }

  useEffect(() => {
    // If the current value is not anonymous but is expired, then set the auth user to anonymous
    if (!isAnonymous(authUser) && isExpired(authUser)) {
      setAuthUserSession(ANONYMOUS);
    }
  }, [authUser, setAuthUserSession]);


  return [ authUser, setAuthUser ];
}

export default useAuth;