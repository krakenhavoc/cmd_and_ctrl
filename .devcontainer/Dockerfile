FROM mcr.microsoft.com/devcontainers/javascript-node:4-24-bookworm

# Pre-create persistent home-directory paths owned by the `node` user, so the
# named volumes mounted on top of them inherit this ownership via Docker's
# "populate a volume" behaviour. Empty directories aren't reliably populated,
# so each gets a `.keep` file to force the copy.
USER root
RUN mkdir -p /home/node/.vscode-server \
  /home/node/.claude \
  /home/node/.npm-global/bin \
  && touch /home/node/.vscode-server/.keep \
  /home/node/.claude/.keep \
  /home/node/.npm-global/.keep \
  && chown -R node:node /home/node/.vscode-server \
  /home/node/.claude \
  /home/node/.npm-global

# Keep npm at a known current stable release during image build.
RUN npm install -g npm@11.12.1
