# INDEX file binary layout

- for floats
{key}\x00{type byte = 'f'}{value}{n file indexes}{file indexes}{value}{n file indexes}{file indexes}\n
// TODO possible failure point when distinguising between \n and lower byte of {value}, alternative would be
{key}\x00{type byte = 'f'}{n of values}{value}{n file indexes}{file indexes}{value}{n file indexes}{file indexes}

- for strings
{key}\x00{type byte = 's'}{value}\x00{n file indexes}{file indexes}{value}\x00{n file indexes}{file indexes}\n

- for nulls
{key}\x00{type byte = 'n'}{n file indexes}{file indexes}\n
