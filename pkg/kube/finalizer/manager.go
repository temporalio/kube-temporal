package finalizer

// HasFinalizers describes a thing that has Finalizers.
type HasFinalizers interface {
	// Finalizers returns the set of Finalizers.
	Finalizers() []string
}

// Manager describes a thing that can set and retrieve Finalizers.
type Manager interface {
	HasFinalizers
	// ReplaceFinalizers replaces the resource's set of Finalizers with
	// the supplied Finalizers.
	ReplaceFinalizers([]string)
	// SetFinalizer adds or replaces the supplied finalizer on the resource.
	SetFinalizer(string)
}
