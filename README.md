# take-positions

## Development

Install Go [with Brew](https://medium.com/@jimkang/install-go-on-mac-with-homebrew-5fa421fc55f5)

```sh
cp sample.env .env                              # prepare Env, add correct values to the .env file

go run main.go -race                            # run code
```

## CLI

```sh
cp sample.env .env                              # prepare Env, add correct values to the .env file

docker build -t tp .                            # build image

docker run -d -p 80:8080 --name=tp  --rm -it tp # run image

docker logs -f tp                               # check logs

docker stop tp                                  # stop running container to run updated image
```
