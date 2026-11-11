#!/usr/bin/env bash
# Pulls master, rebuilds the API, applies migrations, and waits until it is healthy.
# GitHub Actions runs it over SSH; it can also be run manually on the server.
set -euo pipefail

cd "$(dirname "$0")/.."

compose() {
	docker compose -f deploy/docker-compose.yml "$@"
}

git pull --ff-only

# Restart with the freshly pulled copy so changes to this script apply to the current deploy.
if [ -z "${DEPLOY_REEXEC:-}" ]; then
	DEPLOY_REEXEC=1 exec "$0" "$@"
fi

# Baked into the binary and shown by GET /liveness.
APP_VERSION="$(git rev-parse --short HEAD)"
export APP_VERSION

docker network inspect web >/dev/null 2>&1 || docker network create web

# Build before migrating so a broken build never leaves the schema ahead of the running API.
compose build api
compose --profile tools run --rm migrate up
compose up -d --remove-orphans

# Use the container's own health check, so this works whatever the proxy is routing to.
container="$(compose ps -q api)"
for _ in $(seq 1 30); do
	status="$(docker inspect --format '{{.State.Health.Status}}' "$container")"
	if [ "$status" = "healthy" ]; then
		echo "Deployed $(git rev-parse --short HEAD), API is healthy"
		exit 0
	fi
	sleep 2
done

echo "API did not become healthy (status: $status)" >&2
compose logs --tail 50 api >&2
exit 1
