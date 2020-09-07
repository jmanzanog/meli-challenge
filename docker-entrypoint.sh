#!/usr/bin/env bash
set -Eeo pipefail
file_env() {
	local var="$1"
	local fileVar="${var}_FILE"
	local def="${2:-}"
	if [ "${!var:-}" ] && [ "${!fileVar:-}" ]; then
		echo >&2 "error: both $var and $fileVar are set (but are exclusive)"
		exit 1
	fi
	local val="$def"
	if [ "${!var:-}" ]; then
		val="${!var}"
	elif [ "${!fileVar:-}" ]; then
		val="$(< "${!fileVar}")"
	fi
	export "$var"="$val"
	unset "$fileVar"
}
_is_sourced() {
	# https://unix.stackexchange.com/a/215279
	[ "${#FUNCNAME[@]}" -ge 2 ] \
		&& [ "${FUNCNAME[0]}" = '_is_sourced' ] \
		&& [ "${FUNCNAME[1]}" = 'source' ]
}
docker_create_db_directories() {
	local user; user="$(id -u)"

	mkdir -p "$PGDATA"
	chmod 700 "$PGDATA"
	mkdir -p /var/run/postgresql || :
	chmod 775 /var/run/postgresql || :
	if [ -n "$POSTGRES_INITDB_WALDIR" ]; then
		mkdir -p "$POSTGRES_INITDB_WALDIR"
		if [ "$user" = '0' ]; then
			find "$POSTGRES_INITDB_WALDIR" \! -user postgres -exec chown postgres '{}' +
		fi
		chmod 700 "$POSTGRES_INITDB_WALDIR"
	fi

	if [ "$user" = '0' ]; then
		find "$PGDATA" \! -user postgres -exec chown postgres '{}' +
		find /var/run/postgresql \! -user postgres -exec chown postgres '{}' +
	fi
}

docker_init_database_dir() {
	if ! getent passwd "$(id -u)" &> /dev/null && [ -e /usr/lib/libnss_wrapper.so ]; then
		export LD_PRELOAD='/usr/lib/libnss_wrapper.so'
		export NSS_WRAPPER_PASSWD="$(mktemp)"
		export NSS_WRAPPER_GROUP="$(mktemp)"
		echo "postgres:x:$(id -u):$(id -g):PostgreSQL:$PGDATA:/bin/false" > "$NSS_WRAPPER_PASSWD"
		echo "postgres:x:$(id -g):" > "$NSS_WRAPPER_GROUP"
	fi

	if [ -n "$POSTGRES_INITDB_WALDIR" ]; then
		set -- --waldir "$POSTGRES_INITDB_WALDIR" "$@"
	fi

	eval 'initdb --username="$POSTGRES_USER" --pwfile=<(echo "$POSTGRES_PASSWORD") '"$POSTGRES_INITDB_ARGS"' "$@"'

	# unset/cleanup "nss_wrapper" bits
	if [ "${LD_PRELOAD:-}" = '/usr/lib/libnss_wrapper.so' ]; then
		rm -f "$NSS_WRAPPER_PASSWD" "$NSS_WRAPPER_GROUP"
		unset LD_PRELOAD NSS_WRAPPER_PASSWD NSS_WRAPPER_GROUP
	fi
}

docker_verify_minimum_env() {

	if [ "${#POSTGRES_PASSWORD}" -ge 100 ]; then
		cat >&2 <<-'EOWARN'
			WARNING: The supplied POSTGRES_PASSWORD is 100+ characters.
			  This will not work if used via PGPASSWORD with "psql".
			  https://www.postgresql.org/message-id/flat/E1Rqxp2-0004Qt-PL%40wrigleys.postgresql.org (BUG #6412)
			  https://github.com/docker-library/postgres/issues/507
		EOWARN
	fi
	if [ -z "$POSTGRES_PASSWORD" ] && [ 'trust' != "$POSTGRES_HOST_AUTH_METHOD" ]; then
		cat >&2 <<-'EOE'
			Error: Database is uninitialized and superuser password is not specified.
			       You must specify POSTGRES_PASSWORD to a non-empty value for the
			       superuser. For example, "-e POSTGRES_PASSWORD=password" on "docker run".
			       You may also use "POSTGRES_HOST_AUTH_METHOD=trust" to allow all
			       connections without a password. This is *not* recommended.
			       See PostgreSQL documentation about "trust":
			       https://www.postgresql.org/docs/current/auth-trust.html
		EOE
		exit 1
	fi
	if [ 'trust' = "$POSTGRES_HOST_AUTH_METHOD" ]; then
		cat >&2 <<-'EOWARN'
			********************************************************************************
			WARNING: POSTGRES_HOST_AUTH_METHOD has been set to "trust". This will allow
			         anyone with access to the Postgres port to access your database without
			         a password, even if POSTGRES_PASSWORD is set. See PostgreSQL
			         documentation about "trust":
			         https://www.postgresql.org/docs/current/auth-trust.html
			         In Docker's default configuration, this is effectively any other
			         container on the same system.
			         It is not recommended to use POSTGRES_HOST_AUTH_METHOD=trust. Replace
			         it with "-e POSTGRES_PASSWORD=password" instead to set a password in
			         "docker run".
			********************************************************************************
		EOWARN
	fi
}

docker_process_init_files() {
	psql=( docker_process_sql )

	echo
	local f
	for f; do
		case "$f" in
			*.sh)

				if [ -x "$f" ]; then
					echo "$0: running $f"
					"$f"
				else
					echo "$0: sourcing $f"
					. "$f"
				fi
				;;
			*.sql)    echo "$0: running $f"; docker_process_sql -f "$f"; echo ;;
			*.sql.gz) echo "$0: running $f"; gunzip -c "$f" | docker_process_sql; echo ;;
			*.sql.xz) echo "$0: running $f"; xzcat "$f" | docker_process_sql; echo ;;
			*)        echo "$0: ignoring $f" ;;
		esac
		echo
	done
}

docker_process_sql() {
	local query_runner=( psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --no-password )
	if [ -n "$POSTGRES_DB" ]; then
		query_runner+=( --dbname "$POSTGRES_DB" )
	fi

	"${query_runner[@]}" "$@"
}

docker_setup_db() {
	if [ "$POSTGRES_DB" != 'postgres' ]; then
		POSTGRES_DB= docker_process_sql --dbname postgres --set db="$POSTGRES_DB" <<-'EOSQL'
			CREATE DATABASE :"db" ;
		EOSQL
		echo
	fi
}

docker_setup_env() {
	file_env 'POSTGRES_PASSWORD'

	file_env 'POSTGRES_USER' 'postgres'
	file_env 'POSTGRES_DB' "$POSTGRES_USER"
	file_env 'POSTGRES_INITDB_ARGS'

	: "${POSTGRES_HOST_AUTH_METHOD:=md5}"

	declare -g DATABASE_ALREADY_EXISTS

	if [ -s "$PGDATA/PG_VERSION" ]; then
		DATABASE_ALREADY_EXISTS='true'
	fi
}

pg_setup_hba_conf() {
	{
		echo
		if [ 'trust' = "$POSTGRES_HOST_AUTH_METHOD" ]; then
			echo '# warning trust is enabled for all connections'
			echo '# see https://www.postgresql.org/docs/12/auth-trust.html'
		fi
		echo "host all all all $POSTGRES_HOST_AUTH_METHOD"
	} >> "$PGDATA/pg_hba.conf"
}

docker_temp_server_start() {
	if [ "$1" = 'postgres' ]; then
		shift
	fi

	set -- "$@" -c listen_addresses='' -p "${PGPORT:-5432}"

	PGUSER="${PGUSER:-$POSTGRES_USER}" \
	pg_ctl -D "$PGDATA" \
		-o "$(printf '%q ' "$@")" \
		-w start
}

docker_temp_server_stop() {
	PGUSER="${PGUSER:-postgres}" \
	pg_ctl -D "$PGDATA" -m fast -w stop
}

_pg_want_help() {
	local arg
	for arg; do
		case "$arg" in
			-'?'|--help|--describe-config|-V|--version)
				return 0
				;;
		esac
	done
	return 1
}

_main() {
	if [ "${1:0:1}" = '-' ]; then
		set -- postgres "$@"
	fi

	if [ "$1" = 'postgres' ] && ! _pg_want_help "$@"; then
		docker_setup_env
		# setup data directories and permissions (when run as root)
		docker_create_db_directories
		if [ "$(id -u)" = '0' ]; then
			# then restart script as postgres user
			exec su-exec postgres "$BASH_SOURCE" "$@"
		fi

		if [ -z "$DATABASE_ALREADY_EXISTS" ]; then
			docker_verify_minimum_env

			ls /docker-entrypoint-initdb.d/ > /dev/null

			docker_init_database_dir
			pg_setup_hba_conf

			export PGPASSWORD="${PGPASSWORD:-$POSTGRES_PASSWORD}"
			docker_temp_server_start "$@"

			docker_setup_db
			docker_process_init_files /docker-entrypoint-initdb.d/*

			docker_temp_server_stop
			unset PGPASSWORD

			echo
			echo 'PostgreSQL init process complete; ready for start up.'
			echo
		else
			echo
			echo 'PostgreSQL Database directory appears to contain a database; Skipping initialization'
			echo
		fi
	fi

	exec "$@"
}

if ! _is_sourced; then
	_main "$@"
fi