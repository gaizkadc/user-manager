
.DEFAULT_GOAL := all

include scripts/Makefile.golang
include scripts/Makefile.docker
include scripts/Makefile.k8s
include scripts/Makefile.common
