package gce

// MigLister is a mig listing interface
type MigLister interface {
	// GetMigs returns the list of migs
	GetMigs() []Mig
	// HandleMigIssue handles an issue with a given mig
	HandleMigIssue(migRef GceRef)
}

type migLister struct {
	cache *GceCache
}

// NewMigLister returns an instance of migLister
func NewMigLister(cache *GceCache) *migLister {
	return &migLister{
		cache: cache,
	}
}

// GetMigs returns the list of migs
func (l *migLister) GetMigs() []Mig {
	return l.cache.GetMigs()
}

// HandleMigIssue handles an issue with a given mig
func (l *migLister) HandleMigIssue(_ GceRef) {
}
