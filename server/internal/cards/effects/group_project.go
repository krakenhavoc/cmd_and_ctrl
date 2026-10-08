package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Group Project — Sorcery {1}{W}:
//
//	"Create a 2/2 red and white Spirit creature token.
//	 Flashback—Tap three untapped creatures you control. (You may cast
//	 this card from your graveyard for its flashback cost. Then exile
//	 it.)"
//
// The flashback (CR 702.34a) is ADR 0135 §1's tap alternative cost
// (#2030): no mana, three untapped creatures you control of any colour,
// tapped with the spell on the stack; a creature that arrived this turn
// pays (CR 302.6).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:      "e1ce210e-2d1d-4ab4-bfb7-8e36884797fc",
		Name:          "Group Project",
		Completeness:  CompletenessFull,
		CastableZones: []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{
			FlashbackTap(3, "three untapped creatures you control", Creature()),
		},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return CreateToken{Controller: item.Controller, Template: TokenCard("2/2 red and white Spirit"), N: 1}.Apply(ctx)
		},
	})
}
