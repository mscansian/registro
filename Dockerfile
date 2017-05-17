FROM scratch
WORKDIR /registro
COPY ./registro /registro
ENTRYPOINT ["/registro/registro"]