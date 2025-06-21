# This file is bogus and meant to violate policies
FROM debian:lastest AS runtime

RUN apt-get update && apt-get install curl vim

ENTRYPOINT ["vim"]
