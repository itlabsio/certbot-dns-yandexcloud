# Certbot manual DNS challenge

## Table of content

- [Common](#common)
- [How to build](#how-to-build)
- [Quickstart](#quickstart)

## Common

This application automates the process of completing a dns-01 challenge (DNS01) by creating, and subsequently modification, TXT records using the Yandex.Cloud API.

When you run certbot with key **--manual**, he places validation TXT record (for DNS) in environment variable CERTBOT_VALIDATION. Then certbot causes shell-script **hook.sh** which configures certbot-dns-yandexcloud for work with Yandex.Cloud DNS API. Certbot-dns-yandexcloud goes to Yandex.Cloud DNS API and creates (or modify) TXT record for successfully complete the challenge. Next the execution flow is transferred to the certbot, which successfully receives the SSL certificate.

## How to build

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
