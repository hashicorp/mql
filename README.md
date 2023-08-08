# MQL

Model Query Language (MQL) is a query language for your database models.

The `mql` Go package that provides a language that end users can use to query your
database models, without them having to learn SQL or exposing your
application to sql injection.

First, you define a model you wish to query as a Go `struct` and then provide a `mql`
query. The package then uses the query along with a model to generate a
parameterized SQL where clause.

Fields in your model can be compared with the following operators: `=`, `!=`,
`>=`, `<=`, `<`, `>`, `%` . 

Double quotes `"` can be used to quote strings.

Comparison operators can have optional leading/trailing whitespace.

The `%` operator allows you to do partial string matching using LIKE and and this
matching is case insensitive.

The `=` equality operator is case insensitive when used with string fields.

Comparisons can be combined using: `and`, `or`.

More complex queries can be created using parentheses.

Example query: 
`name=alice and age > 11 and (region % Boston or region="south shore")` 

**Please note**: We take security and our users' trust very seriously. If you
believe you have found a security issue, please *[responsibly
disclose](https://www.hashicorp.com/security#vulnerability-reporting)* by
contacting us at  security@hashicorp.com.

## Contributing

Thank you for your interest in contributing! Please refer to
[CONTRIBUTING.md](https://github.com/hashicorp/mql/blob/main/CONTRIBUTING.md)
for guidance.
