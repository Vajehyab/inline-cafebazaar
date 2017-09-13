# inline-cafebazaar

### How to run

0. `git clone https://github.com/Vajehyab/inline-cafebazaar`
1. Modify your Vajehyab token in `vajehyab.go`
1. Install golang and `go build -i -o inline`.
2. Write your Caddyfile or use this one :

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

3. Start caddy server using `caddy` command.
4. Start `./inline` program.
