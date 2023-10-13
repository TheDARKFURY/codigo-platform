#!/usr/bin/env bash
# Fork from https://github.com/rust-lang/docker-rust/blob/master/1.68.2/bullseye/slim/Dockerfile

set -eux

apt-get update
apt-get install -y --no-install-recommends \
  ca-certificates \
  gcc \
  libc6-dev \
  wget

dpkgArch="$(dpkg --print-architecture)"
case "${dpkgArch##*-}" in
amd64)
  rustArch='x86_64-unknown-linux-gnu'
  rustupSha256='bb31eaf643926b2ee9f4d8d6fc0e2835e03c0a60f34d324048aa194f0b29a71c'
  ;;
armhf)
  rustArch='armv7-unknown-linux-gnueabihf'
  rustupSha256='6626b90205d7fe7058754c8e993b7efd91dedc6833a11a225b296b7c2941194f'
  ;;
arm64)
  rustArch='aarch64-unknown-linux-gnu'
  rustupSha256='4ccaa7de6b8be1569f6b764acc28e84f5eca342f5162cd5c810891bff7ed7f74'
  ;;
i386)
  rustArch='i686-unknown-linux-gnu'
  rustupSha256='34392b53a25c56435b411d3e575b63aab962034dd1409ba405e708610c829607'
  ;;
*)
  echo >&2 "unsupported architecture: ${dpkgArch}"
  exit 1
  ;;
esac
url="https://static.rust-lang.org/rustup/archive/1.25.2/${rustArch}/rustup-init"
wget "$url"
echo "${rustupSha256} *rustup-init" | sha256sum -c -
chmod +x rustup-init
./rustup-init -y --no-modify-path --profile minimal --default-toolchain $RUST_VERSION --default-host ${rustArch}
chmod -R a+w $RUSTUP_HOME $CARGO_HOME
rustup component add rustfmt

apt-get remove -y --auto-remove wget
rm rustup-init
rm -rf /var/lib/apt/lists/*;