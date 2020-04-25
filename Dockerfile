FROM alpine:3.7

WORKDIR /app
RUN mkdir logs 
COPY log-generator .
CMD sh -c "cd /app && ./log-generator"




