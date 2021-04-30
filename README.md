# Тестовое задание для Golang разработчика

Есть сервис, который может обрабатывать объекты (назовем их `[]Item`) батчами. 

Но при этом есть ограничение на кол-во обрабатываемых `[]Item` в заданный интервал. При превышении лимита запросы будут блокироваться.

Например если лимит установлен в 2000 `Item` в 2 минуты, то можно сделать 4 запроса в сервис по 500 `Item` или 1 запрос с 2000 `Item` или 2 запроса по 1000 `Item`.

Необходимо реализовать механизм позволяющий делать запросы к серису без блокировок.

При реализации можно оперировать следующими типами:

```golang
type Item struct{}
type Service interface {
  Process(ctx context.Context, items []Item) error
}
var ErrBlocked = errors.New("blocked")
```

## Требования

- Решение нужно оформить в отдельном репозитории на github/gitlab/bitbucket.
- Желательно написать тесты к своему решению.
- Если при разработке использовались статические анализаторы, то нужно конфиг файл (.golangci.yml) положить в репозиторий.
