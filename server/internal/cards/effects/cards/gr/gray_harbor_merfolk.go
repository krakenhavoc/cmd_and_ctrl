package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gray Harbor Merfolk — Creature — Merfolk Rogue {1}{U}, 0/3 (Commander
// Legends: Battle for Baldur's Gate, #5948):
//
//	"This creature can't be blocked.
//	 This creature gets +2/+0 as long as you control a commander that's a
//	 creature or planeswalker."
//
// Mist-Cloaked Herald's evasion and Ballynock Cohort's layer 7c bonus, read
// continuously. "A commander" is any commander its controller controls, one
// they own or one they stole, as long as it is a creature or planeswalker now.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f0f79389-95f1-4110-9d3d-91779bc71604",
		Name:         "Gray Harbor Merfolk",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.RestrictSelf(game.CantBeBlocked),
			{
				Label:    "This creature gets +2/+0 as long as you control a commander that's a creature or planeswalker.",
				Layer:    game.Layer7PT,
				SubLayer: game.SubLayer7C_Modify,
				AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
					if !effects.SelfCreature(target, g, source) {
						return false
					}
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.IsCommander && c.Controller == source.Controller && (c.IsCreature() || c.IsPlaneswalker()) {
							return true
						}
					}
					return false
				},
				Apply: effects.PumpBy(2, 0),
			},
		},
	})
}
