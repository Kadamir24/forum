FROM golang:latest AS build-env
RUN mkdir /forum
WORKDIR /forum

# Get dependancies - will also be cached if we won't change mod/sum

# COPY the source code as the last step
COPY . /forum
RUN cd /forum && go mod download && go build -o forum
# Build the binary

FROM debian:buster
#Alpine is one of the lightest linux containers out there, only a few 4.15MB
# RUN apt-get update && apt-get add ca-certificates && rm -rf /var/cache/apk*
RUN mkdir /app
WORKDIR /app
COPY --from=build-env /forum /app
#Here we copy the binary from the first image (build-env) to the new alpine container
EXPOSE 8080
ENTRYPOINT [ "./forum" ]