FROM photon:4.0
COPY before_reindex /usr/local/bin
WORKDIR /workdir
ENTRYPOINT ["/usr/local/bin/before_reindex"]