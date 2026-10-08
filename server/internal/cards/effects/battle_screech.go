package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Battle Screech — Sorcery {2}{W}{W}:
//
//	"Create two 1/1 white Bird creature tokens with flying.
//	 Flashback—Tap three untapped white creatures you control. (You may
//	 cast this card from your graveyard for its flashback cost. Then
//	 exile it.)"
//
// The flashback (CR 702.34a) is ADR 0135 §1's tap alternative cost
// (#2030): no mana, three untapped white creatures you control, tapped
// with the spell on the stack. Tapping a creature to pay a spell's cost
// is not the {T} symbol (CR 302.6), so the two Birds the hand cast just
// made can pay for the flashback, as the card's rulings say.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:      "e73131cd-b454-405b-9539-9d777e232b9e",
		Name:          "Battle Screech",
		Completeness:  CompletenessFull,
		CastableZones: []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{
			FlashbackTap(3, "three untapped white creatures you control", OfColor("W"), Creature()),
		},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return CreateToken{Controller: item.Controller, Template: TokenCard("1/1 white Bird with flying"), N: 2}.Apply(ctx)
		},
	})
}
