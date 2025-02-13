FROM scratch
COPY leaktk-scanner /usr/local/bin/leaktk-scanner
ENTRYPOINT ["/usr/local/bin/leaktk-scanner"]