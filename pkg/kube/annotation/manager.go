package annotation

// HasAnnotations describes a thing that has Annotations.
type HasAnnotations interface {
	// Annotations returns the set of Annotations.
	Annotations() Annotations
}

// Manager describes a thing that can set and retrieve Annotations.
type Manager interface {
	HasAnnotations
	// ReplaceAnnotations replaces the resource's set of Annotations with
	// the supplied Annotations.
	ReplaceAnnotations(Annotations)
	// SetAnnotation adds or replaces the supplied annotation key and value.
	SetAnnotation(string, string)
}
