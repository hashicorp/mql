# Coverage

This `coverage` directory contains the bits required to generate the code
coverage report and badge which are published for this repo.  

After making changes to source, please run `make coverage` in the root directory
of this repo and check-in any changes.

- **cov-diff.sh** - generates a new coverage report and checks the previous
  entry in `coverage.log` for differences.  It's used by the github action to
  ensure that the published coverage report and badge are up to date.
- **coverage.sh** - generates `coverage.log`, `coverage.svg`, and
  `coverage.html`. 
- **coverage.log** - A log of coverage report runs.
- **coverage.html** - The published coverage report.
- **coverage.svg** - The published coverage badge.
