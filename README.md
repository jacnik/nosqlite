# INDEX file binary layout

- for floats
{key}\x00{type byte = 'f'}{n of values}{value}{n file indexes}{file indexes}{value}{n file indexes}{file indexes}

- for strings
{key}\x00{type byte = 's'}{n of values}{value}\x00{n file indexes}{file indexes}{value}\x00{n file indexes}{file indexes}

- for nulls
{key}\x00{type byte = 'n'}{n file indexes}{file indexes}
