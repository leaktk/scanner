FROM docker.io/alpine/git
COPY leaktk-scanner /usr/local/bin/leaktk-scanner
RUN \
  mkdir -p /var/lib/leaktk && \
  adduser -D -h /var/lib/leaktk/scanner -u 1001 -G root leaktk-scanner && \
  chmod -R ug=rwX,o-rwx /var/lib/leaktk/scanner

USER 1001
ENTRYPOINT ["/usr/local/bin/leaktk-scanner"]
