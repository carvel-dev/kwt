package workspace

type ContainerEnv struct{}

func (e ContainerEnv) WorkingDir() string { return "/tmp/build/ary23" }
