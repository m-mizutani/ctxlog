package ctxlog

// Option represents configuration options for logger creation
type Option interface {
	apply(cfg *config)
}

// config holds configuration for logger creation
type config struct {
	scope     *Scope
	sampling  *float64
	condition func() bool
	fastRand  bool
}

// samplingOption implements Option interface for sampling
type samplingOption struct {
	rate float64
}

func (s samplingOption) apply(c *config) {
	c.sampling = &s.rate
}

// WithSampling creates an option to enable probabilistic logging
func WithSampling(rate float64) Option {
	return samplingOption{rate: rate}
}

// conditionOption implements Option interface for conditional logging
type conditionOption struct {
	condition func() bool
}

func (co conditionOption) apply(c *config) {
	c.condition = co.condition
}

// WithCond creates an option to enable conditional logging
func WithCond(condition func() bool) Option {
	return conditionOption{condition: condition}
}

// fastRandOption implements Option interface for fast random number generation
type fastRandOption struct{}

func (fro fastRandOption) apply(c *config) {
	c.fastRand = true
}

// WithFastRand creates an option to use fast pseudo-random numbers for sampling
// instead of cryptographically secure random numbers for better performance
func WithFastRand() Option {
	return fastRandOption{}
}
