FROM golang:1.19.3-alpine
ENV ROOT=/go/src/app
WORKDIR ${ROOT}
# Set destination for COPY
ENV TZ /usr/share/zoneinfo/Asia/Tokyo
ENV GO111MODULE=on

RUN go install github.com/cosmtrek/air@v1.29.0
RUN apk update && apk add git

# Download Go modules
COPY *.go ./
COPY pca ./pca
COPY go.mod .
COPY go.sum .
RUN go mod tidy
# Copy the source code. Note the slash at the end, as explained in
# https://docs.docker.com/engine/reference/builder/#copy
# COPY *.go ./

# # Build
# RUN CGO_ENABLED=0 GOOS=linux go build -o /proxy

# Optional:
# To bind to a TCP port, runtime parameters must be supplied to the docker command.
# But we can document in the Dockerfile what ports
# the application is going to listen on by default.
# https://docs.docker.com/engine/reference/builder/#expose
EXPOSE 80

# Run
CMD go run .