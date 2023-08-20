import axios from 'axios';
import config from '../config';

const API =  axios.create({
  baseURL: config.apiBaseURL,
  headers: {
    'Content-Type': 'application/json'
  },
})

export const login = async (username, password) => {
  const req = {username, password};
  let response = await API.post('login', req);
  return response.data
}

export const status = async () => {
  try {
    const response = await API.get('status');
    return response.data;
  } catch (error) {
    if (error.response) {
      // Handle maintenance mode
      if (error.response.status === 503) {
        return error.response.data;
      }
      console.error("received api error", error.response.status, error.response.data)
    } else {
      console.error("could not connect to status endpoint", error.message)
    }
    return {'status': 'offline', 'uptime': '', 'version': ''};
  }
}

export default API;