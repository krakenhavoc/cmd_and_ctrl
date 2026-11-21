package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cover of Darkness — Enchantment {1}{B} (Onslaught, #5948):
//
//	"As this enchantment enters, choose a creature type.
//	 Creatures of the chosen type have fear. (They can't be blocked except by
//	 artifact creatures and/or black creatures.)"
//
// The type is the stored "as this enters" choice (Card.NamedTribe, Door of
// Destinies' shape); the grant is a layer 6 fear on every creature of that type,
// whoever controls it, read through the layers (a changeling counts). Before a type
// is chosen it gives nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d78654fc-7d90-40ce-9c36-a211424d1a88",
		Name:         "Cover of Darkness",
		Completeness: CompletenessFull,
		AsEnters:     ChooseCreatureTypeAsEnters("Cover of Darkness — choose a creature type"),
		Static: []game.StaticAbility{
			TribalKeywordGrant(TribeFilter{Chosen: true}, "fear"),
		},
	})
}
