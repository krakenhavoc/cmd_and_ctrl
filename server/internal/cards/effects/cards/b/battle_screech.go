package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:      "e73131cd-b454-405b-9539-9d777e232b9e",
		Name:          "Battle Screech",
		Purpose:       game.Purpose{Tokens: 2},
		Completeness:  effects.CompletenessFull,
		CastableZones: []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{
			effects.FlashbackTap(3, "three untapped white creatures you control", effects.OfColor("W"), effects.Creature()),
		},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.CreateToken{Controller: item.Controller, Template: effects.TokenCard("1/1 white Bird with flying"), N: 2}.Apply(ctx)
		},
	})
}
