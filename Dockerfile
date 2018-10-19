FROM debian:9

SHELL ["/bin/bash", "-exo", "pipefail", "-c"]

RUN apt-get update ;\
	DEBIAN_FRONTEND=noninteractive apt-get install --no-install-{recommends,suggests} -y \
		git ca-certificates openssl ;\
	pushd /opt ;\
	git clone https://github.com/OpenVPN/easy-rsa.git ;\
	pushd easy-rsa ;\
	git checkout b38f65927c377c8bb55510229f0cbb8208c756a9 ;\
	popd ;\
	popd ;\
	PATH="${PATH}:/opt/easy-rsa/easyrsa3" ;\
	mkdir /pki-master ;\
	pushd /pki-master ;\
	easyrsa --batch init-pki ;\
	easyrsa --batch build-ca nopass ;\
	easyrsa --batch --req-cn=172.17.0.1 gen-req 172.17.0.1 nopass ;\
	easyrsa --batch --subject-alt-name=IP:172.17.0.1 sign-req server 172.17.0.1 ;\
	popd ;\
	mkdir /pki-agent ;\
	pushd /pki-agent ;\
	easyrsa --batch init-pki ;\
	easyrsa --batch build-ca nopass ;\
	for i in {1..10}; do \
	easyrsa --batch --req-cn=agent$i gen-req agent$i nopass ;\
	easyrsa --batch sign-req client agent$i ;\
	done ;\
	popd ;\
	rm -rf /opt/easy-rsa ;\
	DEBIAN_FRONTEND=noninteractive apt-get purge -y git ca-certificates openssl ;\
	DEBIAN_FRONTEND=noninteractive apt-get autoremove --purge -y ;\
	apt-get clean ;\
	rm -vrf /var/lib/apt/lists/*
