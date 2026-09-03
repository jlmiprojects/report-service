# Run Stage
FROM alpine:3.15.4
RUN apk add --no-cache tzdata
ENV TZ=Africa/Johannesburg

ARG version

# Set environment variable
ENV APP_NAME report-service

# Copy only required data into this image
COPY ./builds/$version/$APP_NAME .

VOLUME ["/reports","/scripts"]


# Start app
CMD ./$APP_NAME
