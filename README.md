# paus-frontend
[![Build Status](https://travis-ci.org/dtan4/paus-frontend.svg?branch=master)](https://travis-ci.org/dtan4/paus-frontend)
[![Docker Repository on Quay](https://quay.io/repository/dtan4/paus-frontend/status "Docker Repository on Quay")](https://quay.io/repository/dtan4/paus-frontend)

Web frontend of [Paus](https://github.com/dtan4/paus)

![paus-frontend](images/paus-frontend.png)

## Usage

``` bash
$ PAUS_BASE_DOMAIN=pausapp.com \
  PAUS_GITHUB_CLIENT_ID=a058xxxxxxxxxxxxxxxx \
  PAUS_GITHUB_CLIENT_SECRET=3d68xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx \
  paus-frontend
```

## Prerequisites

You need to register new GitHub Developer application from [here](https://github.com/settings/applications/new).
Authorization callback URL must be `http(s)://<base domain>/oauth/callback`.

## Environment variables

| Key                         | Required | Description                                         | Default                 | Example                                    |
|-----------------------------|----------|-----------------------------------------------------|-------------------------|--------------------------------------------|
| `PAUS_BASE_DOMAIN`          | Required | Base domain of application URL                      |                         | `pausapp.com`                              |
| `PAUS_GITHUB_CLIENT_ID`     | Required | GitHub OAuth Client ID                              |                         | `a058xxxxxxxxxxxxxxxx`                     |
| `PAUS_GITHUB_CLIENT_SECRET` | Required | GitHub OAuth Client Secret                          |                         | `3d68xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx` |
| `PAUS_SECRET_KEY_BASE`      |          | Secret key for session store                        | (Random string)         | `130d1be37c13a04674a01c4c9c881b7eb1d270be` |
| `PAUS_URI_SCHEME`           |          | URI scheme of application URL (`http`&#124;`https`) | `http`                  | `http`                                     |

## Development

Go 1.6 or higher is required.

``` bash
$ make deps
$ make build
$ bin/paus-frontend
```
