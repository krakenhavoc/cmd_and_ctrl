package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mirran Safehouse — Artifact {3}:
//
//	"As long as this artifact is on the battlefield, it has all
//	 activated abilities of all land cards in all graveyards."
//
// A layer-6 borrow over every graveyard (ADR 0155). A basic land card in
// a graveyard lends its CR 305.6 mana ability ("{T}: Add {G}" for a
// Forest), read from its printed land types, which is what the card is for.
//
// No simplification.
func init() {
	const clause = "As long as this artifact is on the battlefield, it has all activated abilities of all land cards in all graveyards."
	Register(Spec{
		OracleID:     "08826de5-a2d7-4732-a5f7-f580507a7844",
		Name:         "Mirran Safehouse",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{HasActivatedAbilitiesOf(game.AbilityBorrow{
			Label:   clause,
			Lenders: GraveyardLenders(AllGraveyards, ByController(Land())),
		})},
	})
}
