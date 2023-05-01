FROM golang:1.20.3-alpine
ENV ROOT=/go/src/app
WORKDIR ${ROOT}
# Set destination for COPY
ENV TZ /usr/share/zoneinfo/Asia/Tokyo
ENV GO111MODULE=on

RUN go install github.com/cosmtrek/air@v1.29.0

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

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
# EXPOSE 8000

# Run
CMD ["air"]