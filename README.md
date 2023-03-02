# Certbot manual DNS challenge

## Common

go 1.18.9

```bash
go mod tidy
CGO_ENABLED=0 GOOS=linux go build -o certbot-dns-yandexcloud main.go
```

## Quickstart

1. Create file **hook.sh** with content:

    ```bash
    #!/bin/bash

    export CLOUD_ID=
    export SERVICE_ACCOUNT_ID=
    export KEY_ID=
    export DNS_ZONE_ID=
    export DNS_RECORD_NAME=
    export DNS_RECORD_TYPE=TXT
    export DNS_RECORD_TTL=60
    export KEY_FILE=

    /root/certbot-dns-yandexcloud/certbot-dns-yandexcloud
    ```

    - **CLOUD_ID** - ID yandex cloud where you have DNS zones
    - **SERVICE_ACCOUNT_ID** - ID service account that has right for change DNS record
    - **KEY_ID** - ID authorized key for service account
    - **DNS_ZONE_ID** - ID DNS zone where you want create DNS record
    - **DNS_RECORD_TYPE** - if not set, default TXT
    - **DNS_RECORD_TTL** - if not set, default 60
    - **KEY_FILE** - absolute path to file containing private key
    - **/root/certbot-dns-yandexcloud/certbot-dns-yandexcloud** - path to binary file this application

2. Run:

    ```bash
    chmod 0700 path/to/hook.sh
    chmod 0700 path/to/certbot-dns-yandexcloud
    chmod 0600 path/to/private_key
    ```

3. Add file **/etc/cron.d/certbot** with content:

    ```bash
    5 1 * * * root /usr/bin/certbot certonly -n --manual-auth-hook /root/certbot-dns-yandexcloud/hook.sh --manual -d '*.DOMAIN.EXAMPLE'
    ```

    where:  
    ***.DOMAIN.EXAMPLE** - your wildcard domain.
