# Домашнее задание №6: Контроллер умного дома

Начиная с лекции 6 все домашние задания выполняются в рамках проекта: "Контроллер умного дома".
Каждая следующая работа является продолжением предыдущей либо будет использоваться в этом проекте.

Контроллер умного дома это сервис, который предоставляет интерфейс для мониторинга состояния систем дома.  
Он принимает информацию от датчиков, сохраняет ее, может отдать информацию по запросу.  
Мы умышленно упростили некоторую логику этого сервиса, например функционал контроллера ограничен только работой с датчиками двух типов:
- `ContactClosure` - сухие контакты (датчик протечки, переключатель на замыкание или размыкание).
  Обычно события приходят при изменении состоянии датчика. Например, при открытии двери или окна.
- `ADC` - аналоговый вход (термометры, датчики уровней). Обычно события приходят при изменении показаний датчика.
  Например, при изменении показаний датчика температуры или влажности.

## Задание
Детали задания указаны на edu

## Как работать в проекте

* Для каждого задания создайте отдельную ветку.
* После выполнения задания создайте Pull Request в ветку `main` проекта.
* После создания Pull Request отправьте ссылку на PR в EDU
* После того как задача будет принята примите Pull Request в ветку `main` проекта.
* Для работы над следующим заданием сделайте новую ветку от ветки `main` проекта.

## Как подтянуть изменения в форк

Обратите внимание, для того чтобы скачать спецификацию и тесты для следующих заданий, нужно подтянуть изменения из основного репозитория.
Для этого:
1. Замержите в свой main всю накопленную работу в своём форке и переключитесь на обновлённый main локально
2. Если не настроен upstream, то сделайте ```git remote add upstream git@github.com:central-university-dev/2025-go-course-lesson6-2025-spring-go-course-lesson6.git``` или ```git remote add upstream https://github.com/central-university-dev/2025-go-course-lesson6-2025-spring-go-course-lesson6.git```
3. Обновите upstream: ```git fetch upstream``` или ```git fetch --all```
4. Подтяните изменения из upstream и ребазируйтесь на них: ```git rebase upstream/main```

## Подготовка окружения

1. Установить docker ([windows](https://docs.docker.com/desktop/install/windows-install/), [Mac](https://docs.docker.com/desktop/install/mac-install/), [Linux](https://docs.docker.com/desktop/install/linux-install/))
    * Если установили не docker-desktop, а docker отдельно - необходимо установить [docker-compose](https://docs.docker.com/compose/install/)
2. Установить [migrate](https://github.com/golang-migrate/migrate/blob/master/cmd/migrate/README.md)
3. Базу данных можно развернуть с помощью docker-compose (файл в корне проекта). Для этого необходимо выполнить команду `docker-compose up -d`. После того, как она запустится, к ней можно подключаться - `postgres://postgres:postgres@127.0.0.1:5432/db`.
4. Для миграции нужно выполнить команду `migrate -path=./migrations -database postgres://postgres:postgres@127.0.0.1:5432/db?sslmode=disable up`. Также к проекту приложен Makefile, с помощью которого тоже можно выполнить миграцию - `make migrate-up`.

Если решили выполнить миграцию через Make (`make migrate-up`) на Windows - его нужно [установить](https://stackoverflow.com/questions/32127524/how-to-install-and-use-make-in-windows). В Mac и Linux установка не требуется.

## Запуск приложения

Для запуска приложения требуется [переменная окружения](https://gobyexample.com/environment-variables) `DATABASE_URL` - URL подключения к базе (`postgres://postgres:postgres@127.0.0.1:5432/db?sslmode=disable`).

## Запуск тестов

Тесты в процессе запуска используют docker. Убедитесь, что он у вас запущен.

1. зайти в терминале в каталог с домашним заданием
2. вызвать ```go test -v ./... -race```

## Запуск линтера

Для линтинга используется [golangci-lint](https://golangci-lint.run/).
Инструкцию по установке можно найти [тут](https://golangci-lint.run/usage/install/).

Для запуска линтера нужно выполнить команду `golangci-lint run` в корне проекта.
Большую часть ошибок линтера можно поправить с использованием флага `--fix`.

## Обратите внимание
От того как вы выполните это задание зависит и то, как ваш проект будет продвигаться в дальнейшем. Если вы в чем-то сомневаетесь, то не стесняйтесь задавать вопросы.


## Команда для генерации struct'ов

swagger generate model
