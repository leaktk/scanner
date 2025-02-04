FROM scratch
COPY scanner /usr/local/bin/scanner
ENTRYPOINT ["/usr/local/bin/scanner"]