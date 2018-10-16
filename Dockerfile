FROM debian:9

SHELL ["/bin/bash", "-exo", "pipefail", "-c"]

RUN apt-get update ;\
	DEBIAN_FRONTEND=noninteractive apt-get install --no-install-{recommends,suggests} -y \
		easy-rsa ;\
	apt-get clean ;\
	rm -vrf /var/lib/apt/lists/*

RUN cp -r /usr/share/easy-rsa /pki-master ;\
	cd /pki-master ;\
	ln -vs openssl{-1.0.0,}.cnf ;\
	. vars ;\
	./clean-all ;\
	./pkitool --initca ;\
	./pkitool --server 172.17.0.1

RUN cp -r /usr/share/easy-rsa /pki-agent ;\
	cd /pki-agent ;\
	ln -vs openssl{-1.0.0,}.cnf ;\
	. vars ;\
	./clean-all ;\
	./pkitool --initca ;\
	for i in {1..10}; do ./pkitool agent$i; done
