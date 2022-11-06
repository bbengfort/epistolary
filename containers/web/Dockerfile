# Dynamic Builds
ARG BUILDER_IMAGE=node:18
ARG FINAL_IMAGE=nginx:stable

# Build stage
FROM ${BUILDER_IMAGE} as builder

WORKDIR /usr/src/app

# Dependencies
COPY web/package.json ./
COPY web/yarn.lock ./
RUN yarn install --frozen-lockfile --check-files

# Set production environment variable for build context
ARG REACT_APP_API_BASE_URL="https://api.epistolary.app/v1/"
ARG REACT_APP_ANALYTICS_ID
ARG REACT_APP_VERSION_NUMBER
ARG REACT_APP_GIT_REVISION
ARG REACT_APP_SENTRY_DSN
ARG REACT_APP_SENTRY_ENVIRONMENT
ARG REACT_APP_USE_DASH_LOCALE

# Build app with react-scripts
COPY web/ ./
RUN yarn build

# Final Stage
FROM ${FINAL_IMAGE} AS final

LABEL maintainer="Benjamin Bengfort <benjamin@bengfort.com>"
LABEL description="Epistolary User Interface"

COPY --from=builder /usr/src/app/build/ /usr/share/nginx/html
COPY containers/web/nginx.conf /etc/nginx/conf.d/default.conf

ENV NGINX_ENTRYPOINT_QUIET_LOGS=1
CMD [ "nginx", "-g", "daemon off;" ]