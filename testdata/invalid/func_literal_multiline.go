package invalid

type server struct{}

func serveStdio(server *server) error {
	return nil
}

func run(server *server) error {
	return serveWithDisplays(server, ((func() error { return serveStdio(server) })))
}

func runEach(server *server) error {
	return each(func() error { return serveStdio(server) })
}

func configure(server *server) {
	serve := (func() error { return serveStdio(server) })
	_ = serve
}

var defaultServe = (func() error { return nil })

func serveWithDisplays(server *server, serve func() error) error {
	return serve()
}

func each(serve func() error) error {
	return serve()
}
