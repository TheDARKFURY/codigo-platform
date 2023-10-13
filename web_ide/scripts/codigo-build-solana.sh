#!/usr/bin/env bash
# Fork from https://github.com/solana-labs/solana/blob/master/scripts/cargo-install-all.sh

here="$(dirname "$0")"
readlink_cmd="readlink"
echo "OSTYPE IS: $OSTYPE"
cargo="$("${readlink_cmd}" -f "${here}/../cargo")"

set -ex

# Setup installation DIR
installDir=/usr/local/solana
installDir="$(
  mkdir -p "$installDir"
  cd "$installDir"
  pwd
)"

cd "$(dirname "$0")"/..

# Components to build
BINS=(
  solana
  solana-keygen
  solana-test-validator
  cargo-build-bpf
  cargo-build-sbf
  cargo-test-bpf
  cargo-test-sbf
)

# Prepare components that are going to be build
binArgs=()
for bin in "${BINS[@]}"; do
  binArgs+=(--bin "$bin")
done

# Build components in release mode
"$cargo" build --release "${binArgs[@]}"

# Move artifacts to installation dir
mkdir -p "$installDir/bin/deps"

for bin in "${BINS[@]}"; do
  cp -fv "target/release/$bin" "$installDir"/bin
done

# Build the SBF and move artifacts
mkdir -p "$installDir"/bin/sdk/sbf

"$cargo" build --manifest-path programs/bpf_loader/gen-syscall-list/Cargo.toml
"$cargo" run --bin gen-headers
cp -a sdk/sbf/* "$installDir"/bin/sdk/sbf

# Move dependencies artifacts to installation dir
shopt -s nullglob
for dep in target/release/deps/libsolana*program.*; do
  cp -fv "$dep" "$installDir/bin/deps"
done
