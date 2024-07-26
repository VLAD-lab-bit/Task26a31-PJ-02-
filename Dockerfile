# Этап 1: Сборка
FROM golang:latest AS compiling_stage

# Создаем директорию для проекта
RUN mkdir -p /go/src/Task26a31PJ-02
WORKDIR /go/src/Task26a31PJ-02

# Копируем файлы проекта
ADD . .

# Устанавливаем зависимости и компилируем приложение
RUN go mod tidy
RUN ls -la /go/src/Task26a31PJ-02
RUN go install .
RUN ls -la /go/bin

# Этап 2: Запуск
FROM alpine:latest

# Метаданные изображения
LABEL version="1.0.0"
LABEL maintainer="Test Vlad Veselovskiy<test@test.ru>"

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /root/

# Копируем скомпилированное приложение из первого этапа
COPY --from=compiling_stage /go/bin/Task26a31PJ-02 .

# Проверяем содержимое рабочей директории
RUN ls -la /root/

# Устанавливаем команду по умолчанию для запуска приложения
ENTRYPOINT ["./Task26a31PJ-02"]
