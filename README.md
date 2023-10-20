# LIFE BLOG CRAWLER

## Docker

- Build

```sh
docker build -t maochindada/life-blog-crawler .
docker system prune --volumes -f
```

- Run

```sh
auth="Bearer xxxxxxx"
zone="zzzzzzz"
docker run --rm \
    --name lbcrawler \
    -e CLOUDFLARE_AUTH=$auth \
    -e CLOUDFLARE_ZONE=$zone \
    maochindada/life-blog-crawler:latest
```

- Stop

```sh
docker stop lbcrawler
docker system prune --volumes -f
```

```sh
docker rmi -f $(docker images -a -q)
```
