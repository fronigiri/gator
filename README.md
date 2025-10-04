A simple program that aggregates blog posts through RSS

# Requirements

- Postgres
- Go

# Install
after cloning the repo, you can build the binary with the command ```go install .```


# Usage
you run commands by using ```gator``` along with the given commands:

```feeds```: Displays all the feeds in the database

```agg```: Collects all of the feeds form the RSS at a given interval (i.e 1m, 10s, 1h, etc.)


