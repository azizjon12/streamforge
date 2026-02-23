# StreamForge

StreamForge is a minimal live streaming pipeline built in Go.

It demonstrates a simplified version of a modern streaming architecture:

Video Input → Transcoding → HLS Packaging → API Server → Browser Playback

This project will be built to demonstrate knowledge of:

- HTTP Live Streaming (HLS)
- FFmpeg transcoding and packaging
- Adaptive bitrate concepts
- Go-based HTTP services
- Streaming file delivery
- Cloud-ready architecture

---

## Architecture

MP4 Input  
   ↓  
FFmpeg (H.264 + AAC)  
   ↓  
HLS Segments (.ts) + Manifest (.m3u8)  
   ↓  
Go HTTP Server  
   ↓  
Browser Player (HLS.js)

---

## Tech Stack

- Go 1.22+
- FFmpeg
- HLS.js
- HTML5 Video
- net/http

---

## Prerequisites

Install:

- Go
- FFmpeg

Verify FFmpeg:

```bash
ffmpeg -version
```

---
