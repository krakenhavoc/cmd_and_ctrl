#!/usr/bin/env bash
#
# Pre-flight checks for deploy/Caddyfile, run in CI (see
# .github/workflows/ci-cd.yml) so a Caddyfile that cannot be deployed
# fails the pull request instead of the production deploy.
#
# `caddy validate` alone is demonstrably insufficient: a Caddyfile with
# `admin off` validates cleanly and then makes `systemctl reload caddy`
# fail, because ExecReload runs `caddy reload`, which POSTs the adapted
# config to the running instance's admin API. That is exactly how
# Actions run 34507228563 went red, leaving the host serving a stale
# in-memory config. So this script checks the reload CONTRACT, not just
# the syntax.
#
# Usage: deploy/check-caddyfile.sh [path/to/Caddyfile]
# No toolchain required; uses `caddy` for the extra checks when present.

set -euo pipefail

CADDYFILE="${1:-$(dirname "$0")/Caddyfile}"
status=0

fail() {
	echo "::error file=${CADDYFILE}::$*"
	status=1
}

if [ ! -f "${CADDYFILE}" ]; then
	echo "::error::${CADDYFILE} not found"
	exit 1
fi

# 1. The admin endpoint must not be disabled. `caddy reload` has no way
#    to apply a config to a process with no admin API, so a Caddyfile
#    carrying `admin off` can never be deployed by a reload -- only by a
#    restart, which drops connections. Caddy's default is
#    localhost:2019, loopback only, which is what we want.
if grep -Eq '^[[:space:]]*admin[[:space:]]+off[[:space:]]*$' "${CADDYFILE}"; then
	fail "\`admin off\` disables the admin API, which \`caddy reload\` (systemd ExecReload) requires. Use \`admin localhost:2019\` -- it binds loopback only."
fi

# 2. Guard the reason this file lives in the repo: the @api matcher
#    mirrors the server's route table, and a path that goes missing
#    fails silently by serving index.html. These four are the routes
#    that were absent for months (see the comments in the Caddyfile).
for route in '/auth/*' '/avatars/*' '/bugreport' '/logout'; do
	if ! grep -qF -- "${route}" "${CADDYFILE}"; then
		fail "the @api matcher is missing ${route}; it would fall through to file_server and serve index.html"
	fi
done

# 3. Syntax, when a caddy binary is available. Skipped rather than
#    failed when it is not, so the checks above still run everywhere.
if command -v caddy >/dev/null 2>&1; then
	caddy validate --config "${CADDYFILE}" --adapter caddyfile

	# The structural form of check 1: adapt to JSON and confirm the
	# admin endpoint did not come out disabled, whatever spelling was
	# used to disable it.
	if caddy adapt --config "${CADDYFILE}" --adapter caddyfile 2>/dev/null |
		tr -d ' \n\t' | grep -q '"admin":{"disabled":true'; then
		fail "the adapted config disables the admin endpoint; \`caddy reload\` could not apply it"
	fi
else
	echo "note: caddy not on PATH; skipped \`caddy validate\` (the deploy job validates on the host)"
fi

if [ "${status}" -eq 0 ]; then
	echo "${CADDYFILE}: OK"
fi
exit "${status}"
