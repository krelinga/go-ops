package ops

type Opt interface {
	Update(Env)
}

type OptFunc func(Env)

func (f OptFunc) Update(env Env) {
	f(env)
}

type Opts []Opt

func (opts Opts) Update(env Env) {
	for _, opt := range opts {
		opt.Update(env)
	}
}