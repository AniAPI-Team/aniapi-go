# AniAPI Golang
AniAPI Golang is a REST API meant to expose a collection of **anime**'s streaming videos available across the web.

It uses [MyAnimeList](https://myanimelist.net) as resource validator, to ensure a minimum quality level.

## Requirements
* Golang 1.14
* MongoDB instance

## Dependencies
- [x] github.com/PuerkitoBio/goquery v1.5.1
- [x] github.com/antchfx/htmlquery v1.2.3
- [x] github.com/antchfx/xmlquery v1.2.4
- [x] github.com/darenliang/jikan-go v1.1.0
- [x] github.com/gobwas/glob v0.2.3
- [x] github.com/gocolly/colly v1.2.0
- [x] github.com/kennygrant/sanitize v1.2.4
- [x] github.com/lithammer/fuzzysearch v1.1.0
- [x] github.com/saintfish/chardet v000-20120816061221-3af4cd4741ca
- [x] github.com/temoto/robotstxt v1.1.1
- [x] go.mongodb.org/mongo-driver v1.3.3
- [x] golang.org/x/net v0.0.0-20200513185701-a91f0712d120
- [x] google.golang.org/appengine v1.6.6

## Environment variables
| Name | Description | Required |
|------|-------------|----------|
| `MONGODB_URL` | MongoDB location url | Yes |
| `MONGODB_DB` | MongoDB database name | Yes |
| `PROXY_HOST` | Proxies' hostname | No |
| `PROXY_USER` | Proxies' auth username | No |
| `PROXY_PASSWORD` | Proxies' auth password | No |
| `PROXY_PORT` | Proxies' generic port | No |
| `PROXY_COUNT` | Number of proxies capacity (zero if none) | Yes |
| `PORT` | HttpListener port to bind | Yes |