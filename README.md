The purpose of this project is to play with Golang and its way of doing microservices and maybe, create something useful for a company's intranet.

# Development

Check `Makefile` for some insights.

# URL Shortener

A service to make a long URL given shorter.

### API

#### Submit URL

Request:

```
[POST] /api/urls/
```

Form data expected:

* `url` - URL to make shorter


Response:

```
http://your.domain/s/u32dsad321/
```
