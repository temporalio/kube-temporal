package annotation

// Manager describes a thing that can set and retrieve Annotations.
type Manager interface {
	// Annotations returns the set of Annotations.
	Annotations() Annotations
	// ReplaceAnnotations replaces the resource's set of Annotations with
	// the supplied Annotations.
	ReplaceAnnotations(Annotations)
}
