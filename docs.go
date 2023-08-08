/*
Package mql provides a language that end users can use to query your
database models, without them having to learn SQL or exposing your
application to sql injection.

You define a model you wish to query as a Go struct and provide a mql query. The
package then uses the query along with a model to generate a parameterized SQL
where clause.

Fields in your model can be compared with the following operators:
=, !=, >=, <=, <, >, %

Double quotes `"` can be used to quote strings.

Comparison operators can have optional leading/trailing whitespace.

The % operator allows you to do partial string matching using LIKE and this
matching is case insensitive.

The = equality operator is case insensitive when used with string fields.

Comparisons can be combined using: and, or.

More complex queries can be created using parentheses.

Example: name=alice and age > 11 and (region % Boston or region="south shore")
*/
package mql
