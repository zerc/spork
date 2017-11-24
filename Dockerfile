FROM phusion/baseimage:0.9.22

ADD . /main
WORKDIR /main

CMD ["./main"]
