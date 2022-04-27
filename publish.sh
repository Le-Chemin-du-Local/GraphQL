#!/bin/bash

docker build --build-arg PORT_NUMBER=3000 --build-arg ENV=production -t chemindulocal/chemindulocal-api:latest .
docker tag chemindulocal/chemindulocal-api:latest chemindulocal/chemindulocal-api:$1
docker push chemindulocal/chemindulocal-api:latest
docker push chemindulocal/chemindulocal-api:$1