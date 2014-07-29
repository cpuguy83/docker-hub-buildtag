FROM cpuguy83/debian:jessie
RUN apt-get update && apt-get install -y ca-certificates
RUN mkdir /fetcher
ADD . /fetcher/
ENTRYPOINT ["/fetcher/fetcher", "-assets", "/fetcher"]
