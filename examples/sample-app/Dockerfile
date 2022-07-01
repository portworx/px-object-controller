FROM alpine
LABEL maintainers="PX Control Plane team"
LABEL description="PX Object Controller"
COPY ./bin/sample-app /sample-app
RUN apk add curl 
ENTRYPOINT ["/sample-app"]