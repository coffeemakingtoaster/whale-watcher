# This file is bogus and meant to violate policies
FROM debian:latest AS runtime

RUN apt-get update && apt-get install -y curl vim

ENTRYPOINT ["vim"]
