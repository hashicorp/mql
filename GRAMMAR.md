# MQL Grammar

In a general form, a mql query defines a `condition` which must be satisfied and any resource not
satisfying this condition will be eliminated from output.  

A `condition` is any expression that evaluates to a result of type boolean. 

## keywords (case-insensitive)

* and
* or
  
## tokens

* eq: `=`
* gte: `>=`
* gt: `>`
* lte: `<=`
* lt: `<`
* ne: `!=`
* lparen: `(`
* rparen: `)`
* contains: `%`
* string: `example`
* quote: `"`

## productions

### condition

\<comparison expr> | \<logical expr>

### comparison expr

(lparen)? \<column> (\<whitespace>)? \<comparison operator> (\<whitespace>)?
\<value> (\<rparen)?

### logical expr

(lparen)? \<logical expr> | \<comparison expr> \<logical operator> \<logical expr> |
\<comparison expr> (\<rparen)?


### string

* \<string>

### quoted string

A string delimited by quotes. Note: you can use a quoted string whenever you
need to include logical operators or whitespace in a string value.

* \<quote> \<string> \<quote>

### value

A string (quoted or not) which is the value of a column used in a comparison
expr.  The string must be a valid value/type for the column which will be
enforced by the RDBMS when the query is executed.

* \<string>


### column identifier

An identifier string token that forms a column name and must match a name in the
Go struct used in conjunction with the query and of course it must be a valid
column name for the resource being queried in the RDBMS.

* \<string>

### comparison operator

* \<eq>
* \<gte>
* \<gt>
* \<lte>
* \<lt>
* \<ne>

### logical operator

* \<and>
* \<or>
