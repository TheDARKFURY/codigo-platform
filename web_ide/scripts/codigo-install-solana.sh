#!/usr/bin/env bash

set -Eeuo pipefail

apt-get update
apt-get install -y --no-install-recommends ca-certificates wget bzip2 libssl-dev libudev-dev pkg-config zlib1g-dev llvm clang cmake make libprotobuf-dev protobuf-compiler

wget "https://github.com/solana-labs/solana/archive/refs/tags/v$SOLANA_VERSION.tar.gz" -P /
tar -xzf /v$SOLANA_VERSION.tar.gz
mv /scripts/codigo-build-solana.sh /solana-$SOLANA_VERSION/scripts
chmod +x /solana-$SOLANA_VERSION/scripts/codigo-build-solana.sh

./solana-$SOLANA_VERSION/scripts/codigo-build-solana.sh

# Everything will be install under ~/.cache/solana after SBF completes
# installing, move the content to /home/codigo/.cache/solana
chmod +x /usr/local/solana/bin/sdk/sbf/env.sh
chmod +x /usr/local/solana/bin/sdk/sbf/scripts/install.sh

cd /usr/local/solana/bin/sdk/sbf
./env.sh
cd /

apt-get remove -y --auto-remove ca-certificates wget libssl-dev libudev-dev pkg-config zlib1g-dev llvm clang cmake make libprotobuf-dev protobuf-compiler
rm -rf /var/lib/apt/lists/*
rm /v$SOLANA_VERSION.tar.gz
rm -rf /solana-$SOLANA_VERSION
