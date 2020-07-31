# take-positions

## CLI

```sh
cp sample.env .env

docker build -t tp .

docker run -d -p 80:8080 --name=tp  --rm -it tp

docker logs -f tp

docker stop tp
```
