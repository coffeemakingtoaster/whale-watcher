# This file is bogus and meant to violate policies
FROM debian:latest AS runtime

WORKDIR /app

RUN echo "hello" > test.txt

RUN apt-get update && apt-get install -y curl vim

RUN rm test.txt

ENTRYPOINT ["vim"]
