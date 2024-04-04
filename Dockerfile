# Derived from: https://mitchellh.com/writing/nix-with-dockerfiles
# Nix builder
FROM nixos/nix:latest AS builder

# Copy our source and setup our working dir.
COPY . /tmp/build
WORKDIR /tmp/build

# Build the todo-server Go binary with our Nix environment
RUN nix \
    --extra-experimental-features "nix-command flakes" \
    --option filter-syscalls false \
    build .#todo-server

# Copy the Nix store closure into a directory. The Nix store closure is the
# entire set of Nix store values that we need for our build.
RUN mkdir /tmp/nix-store-closure
RUN cp -R $(nix-store -qR result/) /tmp/nix-store-closure

# Move the built binary to a temporary directory
RUN mv result /tmp/todo-server-result

# Run the assets build
RUN nix \
    --extra-experimental-features "nix-command flakes" \
    --option filter-syscalls false \
    build .#todo-assets

# Final image is based on scratch. We copy a bunch of Nix dependencies
# but they're fully self-contained so we don't need Nix anymore.
FROM scratch

WORKDIR /app

# Copy /nix/store
COPY --from=builder /tmp/nix-store-closure /nix/store
COPY --from=builder /tmp/todo-server-result /app
COPY --from=builder /tmp/build/result /app
CMD ["/app/bin/todo-server"]

# 18.2 MB

