# inline-cafebazaar

This project used glide for package manager.

### How to run

0. `git clone https://github.com/Vajehyab/inline-cafebazaar`
1. Modify your `Vajehyab` token in `vajehyab.go`
2. Install golang and `go build -i -o inline`.
3. `go get github.com/gorilla/handlers github.com/gorilla/mux github.com/oxtoacart/bpool` or use glide update to retrieve vendor tree.
4. Write your `Caddyfile` or use this one :

```
your_site {
        proxy / :<PORT_NUMBER> {
                except /static
        }
        gzip
        root <$HOME>/inline-cafebazaar/public
        tls {
                dns cloudflare
        }
}
```

5. Start caddy server using `caddy` command.
6. Start `./inline` program.
