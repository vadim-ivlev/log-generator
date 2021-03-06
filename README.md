log-generator
================

Докер контейнер для генерации тестового лога.
Часть проекта log-monitor.


Генерирует записи лога вида
```json
{"counter":11,"flight_number":5,"level":"error","msg":"Рейс 5 задерживается на 4.956s","status":"DELAYED","time":"2020-04-25T04:33:02.971Z","wait":4956}

```


docker-compose.yml
```yaml
    # генерирует тестовый лог для проверки работоспособности системы
    log-generator:
        image: vadimivlev/log-generator:latest
        container_name: log-generator
        restart: unless-stopped
        environment: 
            # максимальная задержка добавления записей в лог
            - MAX_DELAY=5000
            # максимальное количество добавленных записей лога перед ротацией
            - MAX_RECORDS=10
            # имя файла лога внутри директории назначенной в параметре volumes:
            - LOG_FILE=logrus.log
            # адрес elasticsearch для прямой записи логов
            - ELASTIC_URL=http://elk:9200
            # хост elasticsearch для прямой записи логов
            - ELASTIC_HOST=elk
        volumes: 
            - ./logs:/app/logs
```


Порядок разработки
=============

1. Модифицируем код.

2. Запуск программы

        go run main.go 

    Чтобы программа сработала директория logs/ где генерируются логи должна уже существовать.

3. Построение контейнера

        sh/build-container.sh

4. Запуск программы в докере

        docker-compose up -d && docker-compose logs -f 

5. Пуш если все в порядке

        sh/push.sh