#!/bin/bash
set -euo pipefail

mkdir -p hls/dev

ffmpeg -y -i assets/sample.mp4 \
  -codec:v libx264 \
  -codec:a aac \
  -preset veryfast \
  -g 48 \
  -sc_threshold 0 \
  -hls_time 4 \
  -hls_playlist_type vod \
  -hls_segment_filename "hls/dev/segment_%03d.ts" \
  hls/dev/playlist.m3u8

echo "HLS generation complete: ./hls/dev/playlist.m3u8"