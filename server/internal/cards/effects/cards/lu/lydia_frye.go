package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lydia Frye — Legendary Creature — Human Assassin {2}{U/B}, 3/2 (Assassin's
// Creed, #5948):
//
//	"Lydia Frye can't be blocked by creatures with power 3 or greater.
//	 At the beginning of your end step, surveil X, where X is the number of
//	 tapped Assassins you control."
//
// The evasion is April O'Neil's block rule (the blocker's power read live as
// blockers are declared). X is counted as the trigger resolves: tapped
// Assassins you control, Lydia herself included when she is tapped, each
// once. With none, X is 0 and nothing is surveilled (CR 701.25: surveil 0
// does nothing).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0315f5e7-2867-428e-8e43-66dda0db57a0",
		Name:         "Lydia Frye",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedBy(effects.OnSelf(), effects.PowerGE(3), "creatures with power 3 or greater"),
		},
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStep("Lydia Frye — surveil X, where X is the number of tapped Assassins you control",
				func(g *game.Game, item *game.StackItem) error {
					assassin := effects.HasSubtype("Assassin")
					x := 0
					for _, c := range g.Battlefield.Cards {
						if c.Controller == item.Controller && c.Tapped && assassin(g, item.Controller, c) {
							x++
						}
					}
					if x == 0 {
						return nil
					}
					return effects.Surveil{Player: item.Controller, N: x}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
