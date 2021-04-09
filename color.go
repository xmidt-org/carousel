package carousel

//go:generate go-enum -f=$GOFILE --marshal --names  --noprefix

// Color is simple type that serves to group deployment groups
// In our case, only two groups should be needed
/*
ENUM(
blue
green
)
*/
type Color int

const Unknown = Color(-1)

// ValidColors is an Array of Valid Color Groups
var ValidColors = []Color{Blue, Green}

// Other returns the complement color between Green and Blue
func (c Color) Other() Color {
	switch c {
	case Green:
		return Blue
	case Blue:
		return Green
	}
	return Unknown
}
