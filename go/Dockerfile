FROM gocv/opencv:4.10.0 AS gocv-stage

# FROM gocv-stage AS ffmpeg-stage
# USER root

# RUN apt update && apt install ffmpeg
FROM gocv-stage AS ffmpeg-stage
USER root
# RUN apk add --no-cache ffmpeg

RUN <<EOF
apt-get update
apt-get install -y ffmpeg
EOF

FROM ffmpeg-stage AS app-stage

WORKDIR /spiritio
COPY . .


# Add gcc for cgo, ffmpeg
# RUN apk update && apk add bash && apk add build-base && apk add git && apk add yasm && make install-ffmpeg


RUN CGO_ENABLED=1 go build -o spiritio ./cmd

# For documentation only
EXPOSE 8080 

ENTRYPOINT ["./spiritio"]

# ENTRYPOINT ["/gocv-version"]