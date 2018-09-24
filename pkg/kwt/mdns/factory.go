package mdns

type Factory struct{}

func NewFactory() Factory { return Factory{} }

func (f Factory) Build(ipResolver IPResolver, logger Logger) *Server {
	customHandler := NewCustomHandler(ipResolver, logger)

	filter := NewLocalIfaceMsgFilter(logger)
	go filter.UpdateIfacesContiniously()

	// Don't use dns's mux since it will return SERVFAIL if there no questions in the request
	return NewServer(customHandler, filter, logger)
}
