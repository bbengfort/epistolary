import axios from 'axios';
import config from '../config';
import APIError from './error';

const API =  axios.create({
  baseURL: config.apiBaseURL,
  headers: {
    'Content-Type': 'application/json'
  },
  withCredentials: true,
})

export const login = async (username, password) => {
  try {
    const req = {username, password};
    const response = await API.post('login', req);
    return response.data
  } catch (error) {
    if (error.response) {
      let data = error.response.data;
      data.statusCode = error.response.status;
      console.error("received api error", error.response.status, error.response.data);
      return data;
    }
    console.error("could not connect to the login endpoint", error.message);
    return {success: false, error: error.message, statusCode: null};
  }
}

export const logout = async () => {
  try {
    const response = await API.post('logout');
    return response.status === 204;
  } catch (error) {
    if (error.response) {
      console.error("received api error", error.response.status, error.response.data);
    } else {
      console.error("could not connect to logout endpoint", error.message);
    }
    return false;
  }
}

export const listReadings = async (pageToken) => {
    let params = {}
    if (pageToken) {
      params.page_token = pageToken;
    }

    try {
      const response = await API.get('reading', { params });
      return response.data;
    } catch (error) {
      if (error.response) {
        let data = error.response.data;
        throw new APIError(data.success, data.error, error.response.status);
      } else {
        throw new APIError(false, error.message, null);
      }
    }
}

export const createReading = async(link) => {
  try {
    const req = {link};
    const response = await API.post('reading', req);
    return response.data;
  } catch (error) {
    if (error.response) {
      let data = error.response.data;
      data.statusCode = error.response.status;

      console.error("received api error", error.response.status, error.response.data);
      return data;
    }

    console.error("could not connect to create readings endpoint", error.message);
    return {success: false, error: error.message, statusCode: null};
  }
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