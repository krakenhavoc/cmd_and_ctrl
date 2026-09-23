package effects

// Exploration — Enchantment {G} (EDHREC rank 301):
//
//	"You may play an additional land on each of your turns."
//
// The permanent half of #500's land-drop work: `Spec.AdditionalLandPlays`
// was added with this card named as its reason for existing (see
// explore.go's doc comment), so the whole card is one field.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:            "0c2841bb-038c-4fbf-8360-bc0a1522b58d",
		Name:                "Exploration",
		Completeness:        CompletenessFull,
		AdditionalLandPlays: 1,
	})
}
