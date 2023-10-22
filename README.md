# INDEX file binary layout

- for floats
{key}\x00{type byte = 'f'}{value}{n file indexes}{file indexes}{value}{n file indexes}{file indexes}\n

- for strings
{key}\x00{type byte = 's'}{value}\x00{n file indexes}{file indexes}{value}\x00{n file indexes}{file indexes}\n

- for nulls
{key}\x00{type byte = 'n'}{n file indexes}{file indexes}\n
