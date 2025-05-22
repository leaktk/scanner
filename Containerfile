FROM docker.io/alpine/git
COPY leaktk /usr/local/bin/leaktk
RUN \
  mkdir -p /var/lib/leaktk && \
  adduser -D -h /var/lib/leaktk -u 1001 -G root leaktk && \
  chmod -R ug=rwX,o-rwx /var/lib/leaktk

USER 1001
ENTRYPOINT ["/usr/local/bin/leaktk"]
