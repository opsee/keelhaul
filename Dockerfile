FROM quay.io/opsee/vinz:latest

ENV POSTGRES_CONN ""
ENV VAPE_KEYFILE ""
ENV NSQLOOKUPD_ADDRS ""
ENV NSQD_HOST ""
ENV NSQ_TOPIC ""
ENV ETCD_ADDR ""
ENV VAPE_ENDPOINT ""
ENV SLACK_ENDPOINT ""
ENV FIERI_ENDPOINT ""
ENV BARTNET_ENDPOINT ""
ENV BEAVIS_ENDPOINT ""
ENV SPANX_ENDPOINT ""
ENV KEELHAUL_ADDRESS ""
ENV BASTION_CONFIG_KEY ""
ENV BASTION_CF_TEMPLATE ""
ENV AWS_ACCESS_KEY_ID ""
ENV AWS_SECRET_ACCESS_KEY ""
ENV AWS_DEFAULT_REGION ""
ENV AWS_INSTANCE_ID ""
ENV AWS_SESSION_TOKEN ""
ENV APPENV ""

RUN apk add --update bash ca-certificates curl
RUN curl -Lo /opt/bin/migrate https://s3-us-west-2.amazonaws.com/opsee-releases/go/migrate/migrate-linux-amd64 && \
    chmod 755 /opt/bin/migrate
RUN curl -Lo /opt/bin/ec2-env https://s3-us-west-2.amazonaws.com/opsee-releases/go/ec2-env/ec2-env && \
    chmod 755 /opt/bin/ec2-env

COPY run.sh /
COPY target/linux/amd64/bin/* /
COPY migrations /migrations
COPY etc/bastion-cf.template /
COPY vape.test.key /

EXPOSE 9092
CMD ["/keelhaul"]
