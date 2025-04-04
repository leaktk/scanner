FROM alpine/git
COPY leaktk-scanner /usr/local/bin/leaktk-scanner
RUN addgroup -S scanner && adduser -S scanner -G scanner
USER scanner
ENTRYPOINT ["/usr/local/bin/leaktk-scanner"]
