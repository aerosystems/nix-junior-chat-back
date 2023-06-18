FROM alpine:latest
RUN mkdir /app
RUN mkdir /app/secrets

COPY ./chat-service.bin /app
COPY ./secrets/* /app/secrets/

# Run the server executable
CMD [ "/app/chat-service.bin" ]
EXPOSE 80