# MQL

Model Query Language (MQL) is a query language for your database models.

The `mql` Go package provides a language that end users can use to query your
database models, without them having to learn SQL or exposing your
application to SQL injection.

## Examples

### github.com/go-gorm/gorm

```Go
w, err := mql.Parse("name=alice or name=bob)",User{})
if err != nil {
    return nil, err
}
err = db.Where(w.Condition, w.Args...).Find(&users).Error
```

### database/sql

```Go
w, err := mql.Parse("name=alice or name=bob)",User{})
if err != nil {
    return nil, err
}
q := fmt.Sprintf("select * from users where %s", w.Condition)
rows, err := db.Query(q, w.Args...)
```

### github.com/hashicorp/go-dbw

```Go
w, err := mql.Parse("name=alice or name=bob)",User{})
if err != nil {
    return nil, err
}
err := rw.SearchWhere(ctx, &users, w.Condition, w.Args)
```

## Some bits about usage

First, you define a model you wish to query as a Go `struct` and then provide a `mql`
query. The package then uses the query along with a model to generate a
parameterized SQL where clause.

Fields in your model can be compared with the following operators: `=`, `!=`,
`>=`, `<=`, `<`, `>`, `%` .

Double quotes `"` can be used to quote strings.

Comparison operators can have optional leading/trailing whitespace.

The `%` operator allows you to do partial string matching using LIKE "%value%". This
matching is case insensitive.

The `=` equality operator is case insensitive when used with string fields.

Comparisons can be combined using: `and`, `or`.

More complex queries can be created using parentheses.

Example query:
`name=alice and age > 11 and (region % Boston or region="south shore")`

### Date/Time fields

If your model contains a time.Time field, then we'll append `::date` to the
column name when generating a where clause and the comparison value must be in
an `ISO-8601` format. Currently, this is the only supported way to compare
dates, if you need something different then you'll need to provide your own
custom validator/converter via `WithConverter(...)` when calling
`mql.Parse(...)`.

We provide default validation+conversion of fields in a model when parsing
and generating a `WhereClause`. You can provide optional validation+conversion
functions for fields in your model via `WithConverter(...)`.

### Mapping column names

You can also provide an optional map from query column identifiers to model
field names via `WithColumnMap(...)` if needed.

**Please note**: We take security and our users' trust very seriously. If you
believe you have found a security issue, please *[responsibly
disclose](https://www.hashicorp.com/security#vulnerability-reporting)* by
contacting us at  security@hashicorp.com.

### Ignoring fields

If your model (Go struct) has fields you don't want users searching then you can
optionally provide a list of columns to be ignored via `WithIgnoreFields(...)`

### Custom converters/validators

Sometimes the default out-of-the-box bits doesn't fit your needs.  If you need to
override how expressions (column name, operator and value) is converted and
validated during the generation of a WhereClause, then you can optionally
provide your own validator/convertor via `WithConverter(...)`

### Grammar

See: [GRAMMAR.md](./GRAMMER.md)

## Contributing

Thank you for your interest in contributing! Please refer to
[CONTRIBUTING.md](https://github.com/hashicorp/mql/blob/main/CONTRIBUTING.md)
for guidance.
