package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nirkana Assassin — {2}{B} Creature — Vampire Assassin Ally 2/3 (Jumpstart 2022, #5948):
//
//	"Whenever you gain life, this creature gains deathtouch until end of turn."
//
// One trigger per life-gain event (YouGainedLife).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "efb64da9-1ca7-4686-a427-e0d67101030c",
		Name:         "Nirkana Assassin",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventChangeLife, effects.YouGainedLife, "Nirkana Assassin — gains deathtouch until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					if !effects.OnBattlefield(g, item.SourceCardID) || effects.SourceIsNewObject(g, item) {
						return nil
					}
					return effects.GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{"deathtouch"}, Label: "Nirkana Assassin — deathtouch"}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
