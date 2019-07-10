# Lil

Simple URL shortener based on Go.

## API

`POST /new`         - Create a random short url for a given url. `url` is the only required param.

`GET /<path>`       - Redirect to short url

`DELETE /<path>`    - Delete a short URL

## Example

```sh
# Create a new short url.
curl -X POST "http://localhost:8085/new" -d "url=https://zerodha.com"

# Full URL is returned as response.
{
    "data": "http://localhost:8085/27Fo2rI2"
}

# Use short URL.
curl http://localhost:8085/27Fo2rI2

# Response shows its a permanent redirect our full URL.
<a href="https://zerodha.com">Moved Permanently</a>.

# Delete a existing short URL.
curl -X DELETE http://localhost:8085/27Fo2rI2

# Response if URL exisits.
{
    "data": true
}

# Try accessing delete URL again.
curl http://localhost:8085/27Fo2rI2

# Response
{
    "error":"Not found"
}
## Backend store
```

Currently Redis is the only backend store available but new stores can be easily
added by implementing [store interface](store/store.go), for example here is the
[Redis store implementation](store/redis/store.go).

## TODO

- Basic auth for create and delete APIs. Currently this can be implemented
  behind reverse proxy like Nginx or API gateways like Kong, AWS API gateway.
- Custom path for short URLs instead of random generated paths.
- Redirect stats.
