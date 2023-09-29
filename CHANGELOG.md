# mql CHANGELOG

Canonical reference for changes, improvements, and bugfixes for mql.

## Next

* chore: add govulncheck to github actions ([PR](https://github.com/hashicorp/mql/pull/30))
* update go matrix in CI: remove 1.18 and add 1.21 ([PR](https://github.com/hashicorp/mql/pull/30))

## 0.1.2 (2023/09/18)

* fix: remove "like" from sql keywords checked in fuzzing ([PR](https://github.com/hashicorp/mql/pull/26))
* feat: add support for backtick and single-quote string delimiters ([PR](https://github.com/hashicorp/mql/pull/25))
* feat: require string tokens used as comparison values to be delimited ([PR](https://github.com/hashicorp/mql/pull/23))
* chore: automate some coverage reporting bits ([PR](https://github.com/hashicorp/mql/pull/12))
* tests: add fuzz test for mql.Parse(...)([PR](https://github.com/hashicorp/mql/pull/11))

## 0.1.1 (2023/08/16)

It was a fast-follower patch release, but was needed to support developers that
use the [database/sql](https://pkg.go.dev/database/sql) package.

* tests: add postgres integration tests ([PR](https://github.com/hashicorp/mql/pull/8)).
* feat: add WithPgPlaceholder() option
  ([PR](https://github.com/hashicorp/mql/pull/7)). This PR was critical to
  support folks who use the
  [database/sql](https://pkg.go.dev/database/sql) package.

## 0.1.0 (2023/08/15)

v0.1.0 is the first release.  As a result there are no changes, improvements, or bugfixes from past versions.
