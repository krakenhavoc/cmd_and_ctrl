package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Murasa Pyromancer — Creature — Human Shaman Ally {4}{R}{R} 3/2 (Zendikar,
// S82 OLD batch 492, #5583):
//
//	"Whenever this creature or another Ally you control enters, you may have
//	 this creature deal damage to target creature equal to the number of Allies
//	 you control."
//
// One trigger for the creature's own entry and for any other Ally you control
// (WwkThisOrAnotherAllyEntered, the Worldwake Ally row). The target is chosen
// as it triggers; "you may" is asked as it resolves; the damage is the number
// of Allies you control then, counted through the layers with the word read
// through the creature's text (CR 612.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1d6925f3-d132-4fa1-bde0-82d0128c319f",
		Name:         "Murasa Pyromancer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(effects.On(game.EventETB, effects.WwkThisOrAnotherAllyEntered,
				"Murasa Pyromancer — it deals damage to target creature equal to the number of Allies you control",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					n := 0
					ally := ctx.TextSubtype("Ally")
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller == ctx.Controller() && c.HasSubtype(ally) {
							n++
						}
					}
					for _, t := range ctx.LegalTargets() {
						if n > 0 {
							return effects.DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: n}.Apply(ctx)
						}
					}
					return nil
				}), effects.TargetCreature("target creature")),
				"Murasa Pyromancer — have it deal damage to the target creature?"),
		},
	})
}
