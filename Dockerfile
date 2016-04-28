FROM alpine:3.3

RUN apk add --update bash ca-certificates curl
RUN mkdir -p /opt/bin && \
		curl -Lo /opt/bin/s3kms https://s3-us-west-2.amazonaws.com/opsee-releases/go/vinz-clortho/s3kms-linux-amd64 && \
    chmod 755 /opt/bin/s3kms && \
    curl -Lo /opt/bin/migrate https://s3-us-west-2.amazonaws.com/opsee-releases/go/migrate/migrate-linux-amd64 && \
    chmod 755 /opt/bin/migrate

ENV KEELHAUL_POSTGRES_CONN ""
ENV KEELHAUL_VAPE_KEYFILE ""
ENV KEELHAUL_NSQLOOKUPD_ADDRS ""
ENV KEELHAUL_NSQD_HOST ""
ENV KEELHAUL_NSQ_TOPIC ""
ENV KEELHAUL_ETCD_ADDR ""
ENV KEELHAUL_VAPE_USERINFO_ENDPOINT ""
ENV KEELHAUL_VAPE_EMAIL_ENDPOINT ""
ENV KEELHAUL_LAUNCHES_SLACK_ENDPOINT ""
ENV KEELHAUL_LAUNCHES_ERROR_SLACK_ENDPOINT ""
ENV KEELHAUL_TRACKER_SLACK_ENDPOINT ""
ENV KEELHAUL_FIERI_ENDPOINT ""
ENV KEELHAUL_BARTNET_ENDPOINT ""
ENV KEELHAUL_BEAVIS_ENDPOINT ""
ENV KEELHAUL_HUGS_ENDPOINT ""
ENV KEELHAUL_SPANX_ENDPOINT ""
ENV KEELHAUL_ADDRESS ""
ENV KEELHAUL_BASTION_CONFIG_KEY ""
ENV KEELHAUL_BASTION_CF_TEMPLATE ""
ENV KEELHAUL_CERT="cert.pem"
ENV KEELHAUL_CERT_KEY="key.pem"
ENV APPENV ""

COPY run.sh /
COPY target/linux/amd64/bin/* /
COPY migrations /migrations
COPY etc/bastion-cf.template /
COPY etc/bastion-beta-cf.template /
COPY vape.test.key /
COPY cert.pem /
COPY key.pem /

EXPOSE 9092
CMD ["/keelhaul"]
