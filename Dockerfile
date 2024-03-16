# FROM kneerunjun/gogingonic:latest
FROM golang:1.21.8-alpine3.19
# from the vanilla image of go gin with mgo driver
# mapping for log files
ARG SRC
ARG LOG
ARG RUN
ARG ETC 
ARG BIN
ARG APPNAME
RUN apk add git
RUN mkdir -p ${SRC} && mkdir -p ${LOG} && mkdir -p ${RUN} && mkdir -p ${ETC} && mkdir -p ${BIN}
WORKDIR ${SRC}
# # getting  all the shells to an executable location
# COPY ./shells/ ${BIN} 
RUN chmod -R +x ${BIN}
# RUN touch mycron
# RUN crontab -l > mycron
# # cron that runs every minute t call the trigger
# RUN echo "30 11 * * * /usr/bin/debit-adjust.sh >> /var/log/psa/cron.log" >> mycron
# # RUN echo "46 11 * * * /usr/bin/debit-adjust.sh" >> mycron
# RUN echo "0 11 26 * * /usr/bin/send-poll.sh >> /var/log/psa/cron.log" >> mycron
# #install new cron file
# RUN crontab mycron
# RUN rm mycron
# https://stackoverflow.com/questions/30215830/dockerfile-copy-keep-subdirectory-structure
# since we want the entire directory structure recursively to be copied onto the container
COPY go.mod .
COPY go.sum .
RUN go mod download 

COPY . .
# ENTRYPOINT ["/bin/sh", "-c","go run ."]
# # TODO: when moving to production remove above statement and enable below statement 
# # TODO: change entrypoint in docker-compose.yml
RUN go build -o ${BIN}/${APPNAME} .