/# Common base image for all CSM images. When base image is upgraded, do not update anything as it is automated# URL: https://catalog.redhat.com/software/containers/ubi9/ubi-micro/615bdf943f6014fa45ae1b58?architecture=amd64&image=65e0ac6949fc66cfe14185b4
# Version: ubi9/ubi-micro 9.3-15
DEFAULT_BASEIMAGE="registry.access.redhat.com/ubi9/ubi-micro@sha256:8e33df2832f039b4b1adc53efd783f9404449994b46ae321ee4a0bf4499d5c42"
DEFAULT_GOIMAGE="golang:1.22"