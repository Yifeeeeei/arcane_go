# ArcaneGo

requires go >= 1.22

## if you have makefile

Generate settings:

```bash
make gen
```

config mass_producer_params_xlsx.json

then

```bash
make produce
```

All output will be in output dir

to test your own card, go to file make_card_test.go

then

```bash
make test
```

## if not

Generate settings:

```bash
sh gen.sh
```

config mass_producer_params_xlsx.json

then

```bash
sh produce.sh
```

All output will be in output dir

to test your own card, go to file make_card_test.go

then

```bash
sh test.sh
```

## or directly use go

Generate settings:

```bash
go run github.com/Yifeeeeei/arcane_go gen
```

config mass_producer_params_xlsx.json

then

```bash
go run github.com/Yifeeeeei/arcane_go
```

All output will be in output dir

to test your own card, go to file make_card_test.go

then

```bash
go test
```

