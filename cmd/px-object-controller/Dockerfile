FROM alpine
LABEL maintainers="PX Control Plane team"
LABEL description="PX Object Controller"
COPY ./bin/px-object-controller /px-object-controller
RUN apk add curl 
ENTRYPOINT ["/px-object-controller"]