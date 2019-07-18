# Lil

Simple URL shortener based on Go.

## API

### Redirect

`GET /<id>`           - Redirects to target URL.

### Redirect page

`GET /p/<id>`         - HTML page which redirects to target url. Page additionally renders
                          OpenGraph tags specified while creating short url. Useful for sharing
                          previewable link in social media sites.

### Create a short url

`POST /api/new` - Create a random short url for given url, accepts JSON post body.

#### Params

- `url`  - Target url to redirect.
- `title` - Page title used in paged redirect.
- `og_tags` - List of Open graph tags to be used in paged redirect.
  - `property` - Open graph property tag
  - `content` - Open graph content tag

#### Response

Response returns short uri ID, redirect url and page redirect url.

```json
{
    "data": {
        "id": "<id>",
        "url": "http://localhost/<id>",
        "page": "http://localhost/p/<id>"
    }
}
```

### Get redirect links

`GET /api/<id>` - Get redirect links for given short uri.

#### Response

Response returns short uri ID, redirect url and page redirect url.

```json
{
    "data": {
        "id": "<id>",
        "url": "http://localhost/<id>",
        "page": "http://localhost/p/<id>"
    }
}
```

### Delete a short url

`DELETE /api/<id>` - Delete a give short url.

#### Response

```json
{
    "data": true
}
```

## Examples

### Create a short url

```
# Request
curl -X "POST" "http://localhost:8085/api/new" \
     -H 'Content-Type: text/plain; charset=utf-8' \
     -d $'{
    "url": "https://zerodha.com",
    "title": "Zerodha",
    "og_tags": [
        {"property": "og:image", "content": "https://zerodha.com/static/images/kite-dashboard.png"}
    ]
}'

# Response
HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8

{
    "data": {
        "id": "72p9abOM",
        "url": "http://localhost:8085/72p9abOM",
        "page": "http://localhost:8085/p/72p9abOM"
    }
}
```

### Direct redirect

```
# Request
curl http://localhost:8085/72p9abOM

# Response
HTTP/1.1 301 Moved Permanently
Content-Type: text/html; charset=utf-8
Location: https://zerodha.com

<a href="https://zerodha.com">Moved Permanently</a>.
```

### Page redirect

```
# Page redirect which renders additional open graph data.
curl http://localhost:8085/p/72p9abOM

# Response
HTTP/1.1 200 OK
Date: Thu, 18 Jul 2019 08:14:03 GMT

<html prefix="og: http://ogp.me/ns#">
<head>
    <title>Zerodha</title>
    <meta property="og:image" content="https://zerodha.com/static/images/kite-dashboard.png" />
    <script >
        window.location.replace("https:\/\/zerodha.com");
    </script>
</head>
</html>
```

### Get redirect links

```
# Request
curl "http://localhost:8085/api/72p9abOM"

# Response
HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8

{
    "data": {
        "id": "72p9abOM",
        "url": "http://localhost:8085/72p9abOM",
        "page": "http://localhost:8085/p/72p9abOM"
    }
}
```

# Delete a short url

```
# Request
curl -X "DELETE" "http://localhost:8085/api/72p9abOM"

# Response
HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8

{
    "data": true
}
```

## Backend store

Currently Redis is the only backend store available but new stores can be easily
added by implementing [store interface](store/store.go), for example here is the
[Redis store implementation](store/redis/store.go).

## TODO

- Basic auth for create and delete APIs. Currently this can be implemented
  behind reverse proxy like Nginx or API gateways like Kong, AWS API gateway.
- Custom path for short URLs instead of random generated paths.
- Redirect stats.
