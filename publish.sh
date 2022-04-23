#!/bin/bash

docker build --build-arg PORT_NUMBER=3000 --build-arg ENV=production -t chemindulocal/api:latest .