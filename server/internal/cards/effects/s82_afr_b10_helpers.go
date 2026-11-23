package effects

// afrB10Count is the number a three-row d20 table counts for the result:
// 1—9 is one, 10—19 two and 20 three (Swarming Goblins' tokens, Hoarding
// Ogre's shape). A result below 1 is no row (CR 706.3a).
func AfrB10Count(result int) int {
	switch {
	case result >= 20:
		return 3
	case result >= 10:
		return 2
	case result >= 1:
		return 1
	}
	return 0
}
