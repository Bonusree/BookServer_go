FROM golang:latest AS builder
# create base working directory inside container
WORKDIR /app
# Step 1: Only copy go.mod and go.sum
COPY go.* ./
# Install all the dependencies
RUN go mod tidy
# Now copy all
COPY . .
# build the go application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bookserver_go .

FROM alpine:latest
WORKDIR /app
# COPY everything from last image working directory
COPY --from=builder /app/bookserver_go .
EXPOSE 8080
ENTRYPOINT ["./bookserver_go"]
# Set default command when a container will be started
# In my case, my start command starts the server in 8081 port by default
CMD ["start"]