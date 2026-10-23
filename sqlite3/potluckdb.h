/*
** PotluckDB additions to the SQLite C API.
**
** The library that exports these functions is a drop-in replacement for the
** standard SQLite library: every sqlite3_* function behaves as in SQLite
** itself. Everything PotluckDB adds is declared here.
*/
#ifndef POTLUCKDB_H
#define POTLUCKDB_H

#ifdef __cplusplus
extern "C" {
#endif

#ifndef POTLUCKDB_API
# define POTLUCKDB_API
#endif

/* The PotluckDB build this library comes from, for example "0.8.0-dev.1".
** SQL can read the same value with SELECT potluckdb_version(). */
POTLUCKDB_API const char *potluckdb_version(void);

#ifdef __cplusplus
}
#endif

#endif /* POTLUCKDB_H */
