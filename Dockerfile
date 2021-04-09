FROM centos:7 as builder

ARG VERSION
ARG GITCOMMIT
ARG BUILDTIME

ARG TERRAFORM_VER=0.13.4

RUN yum install -y golang make git epel-release rpm-build rpmdevtools && rpmdev-setuptree
RUN rpm --import https://mirror.go-repo.io/centos/RPM-GPG-KEY-GO-REPO
RUN curl -s https://mirror.go-repo.io/centos/go-repo.repo | tee /etc/yum.repos.d/go-repo.repo
RUN yum install -y gcc openssl-devel bzip2-devel wget unzip
RUN yum update -y

RUN wget https://releases.hashicorp.com/terraform/${TERRAFORM_VER}/terraform_${TERRAFORM_VER}_linux_amd64.zip
RUN unzip terraform_${TERRAFORM_VER}_linux_amd64.zip
RUN mv terraform /usr/local/bin/

RUN yum remove git* -y && yum -y install https://packages.endpoint.com/rhel/7/os/x86_64/endpoint-repo-1.7-1.x86_64.rpm && yum install git -y

RUN yum install -y golang
RUN yum clean all
RUN mkdir -p /go/bin
ENV GOPATH /go
ENV GOBIN /go/bin
ENV PATH $PATH:$GOBIN


WORKDIR /src

COPY . .

RUN make generate build
RUN go get github.com/geofffranks/spruce/cmd/spruce && chmod +x /go/bin/spruce

FROM centos:7

COPY --from=builder /usr/local/bin/terraform /usr/local/bin/terraform
COPY --from=builder /src/carousel /go/bin/spruce /usr/local/bin/
COPY --from=builder /src/carousel.yaml /src/deploy/packaging/entrypoint.sh /src/Dockerfile /src/NOTICE /src/LICENSE /src/CHANGELOG.md /
COPY --from=builder /src/deploy/packaging/carousel_spruce.yaml /tmp/carousel_spruce.yaml

RUN mkdir /etc/carousel/ && touch /etc/carousel/carousel.yaml && chmod 666 /etc/carousel/carousel.yaml

ENTRYPOINT ["/entrypoint.sh"]


CMD ["/carousel"]