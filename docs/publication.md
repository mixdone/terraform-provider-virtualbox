# Публикация провайдера

### 1. Имя репозитория GitHub должно соответствовать terraform-provider-{lowercase}

### 2. Документация
 > Документация технически необязательна и не мешает публикации

Для создания документации предназначен инструмент [tfplugindocs](https://github.com/hashicorp/terraform-plugin-docs)

Как использовать:

+ Добавить `/tools/tools.go` в проект provider
```go
    //go:build tools

    package tools

    import (
	    // Documentation generation
        _ "github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs"
    )
```
+ Запустить следующие команды для проверки установки и создания документации
``` 
    export GOBIN=$PWD/bin
    export PATH=$GOBIN:$PATH
    go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs
    which tfplugindocs
    tfplugindocs
```

+ Запуск `tfplugindocs` сгенерирует дерево на основе написанного кода. Стоит обратить внимание на поля вида [(schema.Schema).MarkdownDescription](https://github.com/hashicorp/terraform-plugin-framework/blob/v1.4.0/datasource/schema/schema.go#L47)

```
    .
    ├── docs
    │  ├── index.md
    │  ├── resources
    │  │   └── index.md
    │  └── data-sources
    │      └── collection.md
```

### 3. Манифест

+ Добавить `terraform-registry-manifest.json` файл в корневой каталог проекта
  > Возможно верисия должна быть `5.0`

```json
    {
      "version": 1,
      "metadata": {
          "protocol_versions": ["6.0"]
      }
    }
```
### 4. Авторизация

+ Войти в [Terraform Registry](https://registry.terraform.io/) и авторизоваться через GitHub.  

+ Получить ключ GPG с помощью следующей команды и добавить его к [Signing Keys](https://registry.terraform.io/settings/gpg-keys)

```
    gpg --armor --export "Key_ID_or_email"

    # -----BEGIN PGP PUBLIC KEY BLOCK-----
    # abcdedfg...
    # -----END PGP PUBLIC KEY BLOCK-----
```
На этом этапе Terraform Registry должен иметь возможность автоматически определять провайдер на основе имени и иметь доступ к его релизам на GitHub.

### 5. Добавить `.goreleaser.yml`
 > `goreleaser` создаёт новый релиз на GitHub с различными артефактами
 
+ Скопировать этот файл с [terraform-provider-scaffolding-framework](https://github.com/hashicorp/terraform-provider-scaffolding-framework/blob/c7f8b736aec6b14daac8533176931af51a0df22a/.goreleaser.yml)
+ Выполнить следующую последовательность команд
```
    git tag v0.1.1
    git push origin v0.1.1
    GITHUB_TOKEN=$(gh auth token) goreleaser release --clean
```
> Если возникает ошибка с `GPG_FINGERPRINT`, то может помочь [следующее](https://developer.hashicorp.com/terraform/registry/providers/publishing#preparing-and-adding-a-signing-key)

