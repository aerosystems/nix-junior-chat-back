FROM alpine:latest
RUN mkdir /app

COPY ./chat-service.bin /app

# Run the server executable
CMD [ "/app/chat-service.bin" ]