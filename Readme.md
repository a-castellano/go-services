# go-services

[![pipeline status](https://git.windmaker.net/a-castellano/go-services/badges/master/pipeline.svg)](https://git.windmaker.net/a-castellano/go-services/pipelines)[![coverage report](https://git.windmaker.net/a-castellano/go-services/badges/master/coverage.svg)](https://a-castellano.gitpages.windmaker.net/go-services/coverage.html)[![Quality Gate Status](https://sonarqube.windmaker.net/api/project_badges/measure?project=a-castellano_go-services_7930712b-1aab-4ea2-a917-853d91ec9cc6&metric=alert_status&token=sqb_a42785fa06f27139e2134dd8221c060aa2324877)](https://sonarqube.windmaker.net/dashboard?id=a-castellano_go-services_7930712b-1aab-4ea2-a917-853d91ec9cc6)

This repo stores services used by many of my projects.

The aim is this repo is save time and repeated code unifying them in one only source.

# Available Services

- [MemoryDatabase](/memorydatabase): MemoryDatabase service that can use [redis](https://git.windmaker.net/a-castellano/go-types/-/tree/master/redis) type from go-types library.
- [MessageBroker](/messagebroker): MessageBroker service that can use [rabbitmq](https://git.windmaker.net/a-castellano/go-types/-/tree/master/rabbitmq) type from go-types library.
