package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hunted Nightmare — Creature — Nightmare {1}{B}{B}, 4/5 (Ikoria: Lair of Behemoths, #5512):
//
//	"Menace
//	 When this creature enters, target opponent puts a deathtouch counter on a
//	 creature they control."
//
// The targeted opponent chooses the creature as the trigger resolves, from
// the creatures they control then (not targeted); with one it goes there
// without a question and with none nothing happens. The opponent may put it
// on a creature that is not the Nightmare's best target, which is the point.
// An illegal target leaves the trigger without effect (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b5aae2f9-ced6-41d3-9121-418831ef5b8f",
		Name:            "Hunted Nightmare",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Hunted Nightmare — target opponent puts a deathtouch counter on a creature they control",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetPlayer {
							continue
						}
						var mine []uuid.UUID
						for _, c := range g.BattlefieldCardsForEffect() {
							if c.Controller == t.ID && c.IsCreature() {
								mine = append(mine, c.InstanceID)
							}
						}
						return effects.S82OldB114CounterPick(g, t.ID, ctx.Source(), "Hunted Nightmare", mine, game.CounterDeathtouch)
					}
					return nil
				}),
				effects.TargetPlayer("target opponent", effects.Opponent())),
		},
	})
}
