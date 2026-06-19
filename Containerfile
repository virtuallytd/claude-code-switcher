FROM node:22-bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    curl \
    git \
    python3 \
    apt-transport-https \
    ca-certificates \
    gnupg \
  && curl https://packages.cloud.google.com/apt/doc/apt-key.gpg | gpg --dearmor -o /usr/share/keyrings/cloud.google.gpg \
  && echo "deb [signed-by=/usr/share/keyrings/cloud.google.gpg] https://packages.cloud.google.com/apt cloud-sdk main" \
     > /etc/apt/sources.list.d/google-cloud-sdk.list \
  && apt-get update && apt-get install -y --no-install-recommends google-cloud-cli \
  && apt-get clean && rm -rf /var/lib/apt/lists/*

RUN npm install -g @anthropic-ai/claude-code

ARG USER_UID=501
ARG USER_GID=20

RUN groupadd -g ${USER_GID} claude 2>/dev/null || true \
  && useradd -m -u ${USER_UID} -g ${USER_GID} -s /bin/bash claude

COPY --chmod=755 entrypoint.sh /usr/local/bin/entrypoint.sh

USER claude
RUN mkdir -p /home/claude/.claude

ENTRYPOINT ["entrypoint.sh"]
