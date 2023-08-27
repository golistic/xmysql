xmysql - MySQL Utilities for Go
===============================

Copyright (c) 2022, 2023, Geert JM Vanderkelen

The Go xmysql package offers functionality for MySQL.  
For example, to create and drop a schema you could use respectively
`CreateSchema()` and `DropSchema()`.

This package is SQL driver agnostic as long as MySQL is used. Both [Go MySQL Driver][1],
using the conventional protocol, and [Go MySQL Driver using X Protocol][2] will work.

License
-------

Distributed under the MIT license. See `LICENSE.md` for more information.

[1]: https://github.com/go-sql-driver/mysql

[2]: https://github.com/golistic/pxmysql