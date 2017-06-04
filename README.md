# localtunnel

[![Travis](https://api.travis-ci.org/jweslley/localtunnel.png)](http://travis-ci.org/jweslley/localtunnel)

[localtunnel](http://localtunnel.me/) is a localtunnel client written in Golang.

localtunnel allows you to expose your localhost to the world for easy testing and sharing!


## Installing

[Download][] and put the binary somewhere in your path.

### Building from source

    git clone http://github.com/jweslley/localtunnel
    make build


## Usage

### Exposing a local port

Assuming your local server is running on port 8000, just use the `lt` command to start the tunnel.

    lt -p 8000

Thats it! A tunnel will be created and the command output will be something like:

    your url is: https://dlaaazhqwd.localtunnel.me

This URL can be used and shared to access your local service from anywhere in the world!

You can restart your local server all you want, `lt` is smart enough to detect this and reconnect once it is back.


### Exposing a local port with a custom subdomain

You also can access your service with a custom subdomain. To this, you need the `-s` option:

    lt -p 8000 -s ltdemo

Output:

    your url is: https://ltdemo.localtunnel.me


### Finishing the tunnel

To finish the tunnel just interrupt the program (`Ctrl-C`).


## API - [GoDoc][]

The localtunnel client is also usable through an API (for test integration, automation, etc).


### Creating a tunnel for a local port

```go
import "github.com/jweslley/localtunnel"

...

var port := 8000
var tunnel := localtunnel.NewLocalTunnel(port)
var err := tunnel.Open()
if (err != nil) {
	fmt.Printf("your url is: %s\n", tunnel.URL())
}

...

tunnel.Close()
```

### Creating a tunnel for a local port with a custom subdomain

```go
import "github.com/jweslley/localtunnel"

...

var port := 8000
var subdomain := "ltdemo"
var tunnel := localtunnel.NewLocalTunnel(port)
var err := tunnel.OpenAs(ltdemo)
if (err != nil) {
	fmt.Printf("your url is: %s\n", tunnel.URL())
}

...

tunnel.Close()
```

For more information, check out the [documentation][GoDoc].


## Bugs and Feedback

If you discover any bugs or have some idea, feel free to create an issue on GitHub:

    http://github.com/jweslley/localtunnel/issues


## License

MIT license. Copyright (c) 2016 Jonhnny Weslley <http://jonhnnyweslley.net>

See the LICENSE file provided with the source distribution for full details.


[download]: https://github.com/jweslley/localtunnel/releases
[GoDoc]: https://godoc.org/github.com/jweslley/localtunnel
