class APIError extends Error {
  constructor(success, message, statusCode) {
    super(message);
    this.name = "APIError";
    this.success = success;
    this.statusCode = statusCode;
  }
}

export default APIError;