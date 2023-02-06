#  Before reindex

## Container command
Compile and create container
```
make -f Makefile.mk all
```
Usage:
   Fix duplicate name when the repository name unique constraint is broken when the index is failed
   The issue often happen when the upgrade from 2.2.x to 2.3.x or 2.4.x

```
docker run --rm firstfloor/before_reindex:latest -h
docker run --rm firstfloor/before_reindex:latest --dburl=postgres://postgres:root123@localhost:5432/registry --command=list
docker run --rm firstfloor/before_reindex:latest --dburl=postgres://postgres:root123@localhost:5432/registry --command=fix
```

