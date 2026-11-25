package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sunlit Hoplite — Creature — Human Soldier {1}{W}, 2/1 (Theros Beyond Death,
// S82 OLD, #5514):
//
//	"During your turn, this creature has first strike.
//	 This creature gets +1/+0 as long as you control an Elspeth planeswalker."
//
// Spartan Veteran's first strike on its controller's own turn (a layer 6
// grant read live, so it switches on and off with the turn) and Vivien's
// Crocodile's pump while the controller controls a planeswalker with the
// subtype Elspeth (a word of the source's text).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e99c58bd-b405-4474-bd32-1a993c9f2979",
		Name:         "Sunlit Hoplite",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
				return effects.ThisPermanent(target, g, source) && effects.IsActivePlayer(g, source.Controller)
			}, "first strike"),
			{
				Label:     "This creature gets +1/+0 as long as you control an Elspeth planeswalker.",
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7C_Modify,
				AppliesTo: effects.SelfWhileYouControlA(effects.And(effects.Planeswalker(), effects.Subtype("Elspeth"))),
				Apply:     effects.PumpBy(1, 0),
			},
		},
	})
}
