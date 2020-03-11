# i18n_migrate

Simple commad to search for i18n properties in several folders and move them to a new one

```sh 
go get github.com/benaich/i18n-migrate
```

## Usage

```sh
i18n_migrate src dst key.txt
```

## Demo 

### Given

```yaml
# src/messages_en.properties
feeling.happiness=happiness
feeling.pain=pain
feeling.fear=fear

# src/messages_fr.properties
feeling.happiness=bonheur
feeling.pain=chagrin
feeling.fear=peur

# keys.properties
feeling.fear
```

### When
```sh
i18n_migrate src dst keys.properties
```

### Then
```yaml
# dst/messages_en.properties
feeling.fear=fear

# dst/messages_fr.properties
feeling.fear=peur
```
