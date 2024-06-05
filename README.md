# webtransport-go-pic

## Requirements

- go 1.21

## Usage
- clone this repo
```
git clone https://github.com/jaehooni/webtransport-go-pic.git
```
- Follow the installation instructions of [mkcert](https://github.com/FiloSottile/mkcert)
- create `localhost.pem` and `localhost-key.pem`
```
cd webtransport-go-video
mkcert -install
mkcert localhost
```

- (Google Chrome) Enable WebTransport Developer Mode
  - [chrome://flags/#webtransport-developer-mode](chrome://flags/#webtransport-developer-mode)

- run `go run main.go`

- open `index.html`
