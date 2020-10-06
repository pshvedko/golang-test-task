# Тестовое задание для Golang разработчика

Написать сервис для обработки и загрузки картинок на S3 хранилище.

## API

Схема работы с сервисом такая, что кто угодно (пользователи, сервисы, партнерские сервисы) грузят картинку, получают его `id`, и подтверждают необходимость сохранения картинки вызовом метода `claim`.

Подтверждение необходимо для того, чтобы не держать в хранилище неиспользуемые данные. То есть, сервис через некоторое конфигурируемое время удаляет неподтвержденные картинки из хранилища.

API должен быть реализован через REST. Но реализация логики не должна зависеть от протокола - чтобы с минимальными усилиями можно было добавить graphql или grpc

### Базовые методы

* `POST /api/upload` - метод для загрузки картинки, умеет работать как с файлом в теле запроса так и ссылкой на картинку. [Обрабатывает](#обработка) и грузит картинку на сконфигурированное S3 хранилище и возвращает ее ссылку и `id`

* `POST /api/claim` - метод для подтверждения картинок. Может принимать один или несколько `id`. Этот метод могут вызывать только [авторизированные](#авторизация) пользователи/сервисы используя API ключ.

### Усложнение (опционально)

Жизнь такова, что мы не можем надеятся только на S3 хранилище только от одного провайдера, и поэтому мы хотим чтобы наши картинки дублировались на один или несколько S3 хранилищ.

Есть главное хранилище, есть второстепенные. В первую очередь, картинки грузятся на главное, после подтверждения (`claim`), асинхронно грузятся на второстепенные.

* `PUT /api/storages` - обновить настройки storage, передаваемое тело запроса может быть, например таким:

```json
{
    "storages": [
        {
            "name": "mail_cloud_solutionns",
            "s3_region": "ru-msk",
            "s3_endpoint": "https://hb.bizmgr.com",
            "s3_access_key": "some_key",
            "s3_access_secret": "secret",
            "s3_bucket": "bucket-name",
            "type": "primary"
        },
        {
            "name": "amazon_s3",
            "s3_region": "eu-c3",
            "s3_endpoint": "https://s3.amazon.com",
            "s3_access_key": "some_key",
            "s3_access_secret": "secret",
            "s3_bucket": "bucket-name",
            "type": "secondary"
        }
    ]
}

```

После добавлении нового `storage`, старые картинки нужно туда асинхронно скопировать.

* `GET /api/storages` - возвращает те же настройки, только без `access_key` и `access_secret`

## Требования к проекту

* использовать PostgreSQL 10+

* использовать строчный тип для `id` картинок

* для обработки картинок можно использовать сторонние библиотеки

* при необходимости можно использовать redis, kafka, rabbitmq

* настроенный `Dockerfile` и `docker-compose.yml` для запуска проекта со всеми зависимостями

* настройка приложения через `.env` файл и переменные окружения

### Бонусные требования

* liveness и readiness healthchecks

* graceful terminatation

* код покрыт тестами на +80%

* настроен CI на прогон тестов и линтера на [drone.io](https://drone.io)/[github_actions](https://github.com/features/actions)/[gitlab_ci](https://docs.gitlab.com/ee/ci/)/other

### Обработка

Загруженные картинки нужно сжимать и нарезать на нужные размеры.

Всего нужно сделать 4 трансформации для одной картинки:

1. 690x920, качество 80% от оригинала - `product_690x920.jpg`

2. 300x400, качество 50% от оригинала - `product_300x400.jpg`

3. 60x80, качество 70% от оригинала - `product_60x80.jpg`

4. оригинальное размер и качество - `original.jpg`

`original` нужно сохранять как есть. Для всех остальных трансформаций, нужно:

* сконвертировать картинку в jpeg (если пришла в png например)

* нарезать нужный размер

* зашакалить как надо

* сохранить как [progressive jpeg](https://www.liquidweb.com/kb/what-is-a-progressive-jpeg/)

Например для картинки с `id = "m42bosu4x1xavjx"` трансформации доступны по ссылкам:

```txt
https://bucket.my-storage.com/optional-prefix/m42bosu4x1xavjx/original.jpg
https://bucket.my-storage.com/optional-prefix/m42bosu4x1xavjx/product_690x920.jpg
https://bucket.my-storage.com/optional-prefix/m42bosu4x1xavjx/product_300x400.jpg
https://bucket.my-storage.com/optional-prefix/m42bosu4x1xavjx/product_60x80.jpg
```

#### Бонус

Предусмотреть вомзожность добавления и управления другими трансформациями. Например, какой-то момент мы захотели добавить новую трансформацию, тогда для **всех** старых картинок перегенерируем новую трансформацию и сохраняем ее во всех хранилищах.

### Авторизация

Для метода `upload` можно не делать авторизацию.

Для метода `claim` можно сделать доступ по secret ключам. Будем считать что у каждого приложениям свой ключ и будем запоминать приложение, которое делает `claim` для картинок.

### Асинхронность

Прогресс всех асинхронных задач (удаление неподтвержденных картинок, копирование на второстепенные хранилища) фиксирутся в базе данных и если вдруг что-то пошло не так (пропала связь с БД, отвалился API у хранилища, выключился сервер) то задачи перезапускаются.
