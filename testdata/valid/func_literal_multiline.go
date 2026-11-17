package valid

type server struct{}

type holder struct {
	serve func() error
}

func serveStdio(server *server) error {
	return nil
}

func run(server *server) error {
	return serveWithDisplays(server, (func() error {
		return serveStdio(server)
	}))
}

func runEmpty(server *server) error {
	return serveWithDisplays(server, func() error {})
}

func runAssigned(server *server) error {
	serve := (func() error {
		return serveStdio(server)
	})
	return serve()
}

func runReturned(server *server) func() error {
	return func() error { return serveStdio(server) }
}

func runDeferred(server *server) {
	defer func() { _ = serveStdio(server) }()
}

func runField(server *server) {
	_ = holder{serve: func() error { return serveStdio(server) }}
}

func serveWithDisplays(server *server, serve func() error) error {
	return serve()
}
