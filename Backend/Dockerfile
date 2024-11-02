FROM golang:1.20 AS builder
WORKDIR /Backend
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main .
FROM alpine:latest
WORKDIR /Backend
# Define environment variables
ENV MONGODB_URI=mongodb+srv://jenilparmar:dsfkjnksdfaa@cluster0.utm2zr0.mongodb.net/


# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/main .

# Command to run the executable
CMD ["./main"]

