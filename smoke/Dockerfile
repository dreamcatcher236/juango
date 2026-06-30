# Runner image for the containerized smoke test (Approach B).
# Bakes in Go + Node + headless Chromium so the whole scaffold->build->serve->worker
# path runs identically on a laptop and in CI, with no host toolchain dependencies.
FROM golang:1.26.3-bookworm

# Node 22 (for the scaffolded frontend build) and Chromium (for the browser tests).
RUN apt-get update \
    && apt-get install -y --no-install-recommends curl ca-certificates gnupg \
    && curl -fsSL https://deb.nodesource.com/setup_22.x | bash - \
    && apt-get install -y --no-install-recommends nodejs chromium \
    && rm -rf /var/lib/apt/lists/*

# chromedp / the browser tests look for these names on PATH.
ENV CHROME_BIN=/usr/bin/chromium

WORKDIR /workspace

# The repo is mounted at /workspace by docker-compose; the module/build caches are
# named volumes so repeat runs are fast.
ENV GOFLAGS=-buildvcs=false

CMD ["go", "test", "./integration/...", "-v", "-timeout", "15m"]
