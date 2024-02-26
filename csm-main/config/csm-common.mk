# Common base image for all CSM images. When base image is upgraded, update the following 3 lines with URL, Version, and DEFAULT_BASEIMAGE variable.
# URL: https://catalog.redhat.com/software/containers/ubi9/ubi-micro/615bdf943f6014fa45ae1b58?architecture=amd64&image=
# Version: ubi9/ubi-micro 
DEFAULT_BASEIMAGE="registry.access.redhat.com/ubi9/ubi-micro@sha256:"
DEFAULT_GOIMAGE="golang:1.21"