# gobore

A Go client for the [bore](https://github.com/ekzhang/bore) tunneling protocol — expose your local ports to the internet through `bore.pub`, with no dependencies beyond the Go standard library.

## What is this?

`gobore` is a pure Go implementation of the bore client. It connects to a bore-compatible server (such as the public `bore.pub`) and tunnels traffic from a remote public port to a local service on your machine — similar to how `ngrok` or `cloudflared` work, but based on the simple and open bore protocol.

## How it works

1. The client connects to `bore.pub:7835` (the control port)
2. It sends a `Hello` message requesting a remote port
3. The server assigns a public port and responds
4. Whenever someone connects to `bore.pub:<assigned-port>`, the server notifies the client via a `Connection` message
5. The client opens a new connection, sends an `Accept` message, and proxies traffic bidirectionally to `localhost:<localPort>`

The protocol uses null-byte delimited JSON frames over TCP, as defined by the original bore project.

## Usage

Edit the constants at the top of `main.go` to match your setup:

```go
const (
    boreServer = "bore.pub"
    localPort  = 8080       // your local service port
    remotePort = 0          // 0 = let the server pick a random port
)
```

Then run:

```bash
go run main.go
```

Output:
```
🔌 connected to bore.pub:7835
✅ tunnel open!
   remote → bore.pub:12345
   local  → localhost:8080
```

Your local service is now reachable at `bore.pub:12345`.

## Requirements

- Go 1.18 or later
- No external dependencies

## Credits

This project is a Go reimplementation of the bore client, originally written in Rust by **Eric Zhang**.

- Original project: [https://github.com/ekzhang/bore](https://github.com/ekzhang/bore)
- Original author: [Eric Zhang](https://github.com/ekzhang)

All credit for the protocol design and the public `bore.pub` server goes to the original author. This project simply reimplements the client side in Go for environments where a Go-native solution is preferred.

## License

MIT

---

# gobore (ქართულად)

Go-ზე დაწერილი კლიენტი [bore](https://github.com/ekzhang/bore) tunneling პროტოკოლისთვის — გამოაქვეყნე შენი local პორტები ინტერნეტში `bore.pub`-ის გავლით, სტანდარტული ბიბლიოთეკის გარდა სხვა dependency-ს გარეშე.

## რა არის ეს?

`gobore` არის bore კლიენტის სუფთა Go იმპლემენტაცია. ის უკავშირდება bore-თავსებად სერვერს (მაგ. საჯარო `bore.pub`) და პუბლიკური პორტიდან ტრაფიკს გადამისამართებს შენს local სერვისზე — მსგავსად `ngrok`-ისა და `cloudflared`-ისა, მაგრამ bore-ის მარტივ და ღია პროტოკოლზე დაყრდნობით.

## როგორ მუშაობს

1. კლიენტი უკავშირდება `bore.pub:7835`-ს (control პორტი)
2. აგზავნის `Hello` შეტყობინებას და ითხოვს remote პორტს
3. სერვერი ანიჭებს პუბლიკურ პორტს და პასუხობს
4. როდესაც ვინმე უკავშირდება `bore.pub:<პორტი>`-ს, სერვერი კლიენტს `Connection` შეტყობინებით ატყობინებს
5. კლიენტი ახალ კავშირს ხსნის, `Accept` შეტყობინებას უგზავნის, და ტრაფიკს ორივე მიმართულებით გადასცემს `localhost:<localPort>`-ს

პროტოკოლი იყენებს null byte-ით გამოყოფილ JSON ფრეიმებს TCP-ზე, ისევე როგორც განსაზღვრულია ორიგინალ bore პროექტში.

## გამოყენება

შეცვალე კონსტანტები `main.go`-ს დასაწყისში:

```go
const (
    boreServer = "bore.pub"
    localPort  = 8080  // შენი local სერვისის პორტი
    remotePort = 0     // 0 = სერვერი თავად ირჩევს პორტს
)
```

გაუშვი:

```bash
go run main.go
```

შედეგი:
```
🔌 connected to bore.pub:7835
✅ tunnel open!
   remote → bore.pub:12345
   local  → localhost:8080
```

შენი local სერვისი ხელმისაწვდომია `bore.pub:12345`-ზე.

## მოთხოვნები

- Go 1.18 ან უფრო ახალი
- გარე dependency არ სჭირდება

## მადლობა

ეს პროექტი bore კლიენტის Go-ზე გადაწერილი ვერსიაა, რომლის ორიგინალი Rust-ზეა დაწერილი **Eric Zhang**-ის მიერ.

- ორიგინალი პროექტი: [https://github.com/ekzhang/bore](https://github.com/ekzhang/bore)
- ავტორი: [Eric Zhang](https://github.com/ekzhang)

პროტოკოლის დიზაინის და საჯარო `bore.pub` სერვერის ყველა დამსახურება ეკუთვნის ორიგინალ ავტორს. ეს პროექტი მხოლოდ კლიენტის მხარეს ახორციელებს Go-ზე.

## ლიცენზია

MIT
