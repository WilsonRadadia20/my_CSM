# Common base image for all CSM images. When base image is upgraded, do not update anything as it is automated.
# URL: https://catalog.redhat.com/software/containers/ubi9/ubi-micro/615bdf943f6014fa45ae1b58?architecture=amd64&image=664f4a3aed0ebe9d680a65fe
# Version: ubi9/ubi-micro 9.4-6.1716471860
DEFAULT_BASEIMAGE="registry.access.redhat.com/ubi9/ubi-micro@sha256:1c8483e0fda0e990175eb9855a5f15e0910d2038dd397d9e2b357630f0321e6d"
DEFAULT_GOIMAGE="golang:1.22"