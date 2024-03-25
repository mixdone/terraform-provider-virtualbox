# Используем официальный образ Golang для сборки и запуска тестов
FROM golang:1.21

# Устанавливаем необходимые зависимости
RUN go get -u github.com/gruntwork-io/terratest/modules/terraform
RUN go get -u github.com/stretchr/testify/assert

# Установка Terraform
RUN wget https://releases.hashicorp.com/terraform/0.12.29/terraform_0.12.29_linux_amd64.zip
RUN unzip terraform_0.12.29_linux_amd64.zip
RUN mv terraform /usr/local/bin/
RUN rm terraform_0.12.29_linux_amd64.zip

# Установка VirtualBox
RUN apt-get update && apt-get install -y virtualbox

# Копируем код тестов и Terraform модулей в рабочую директорию образа
COPY . /go/src/my-terraform-tests

# Указываем рабочую директорию
WORKDIR /go/src/my-terraform-tests

