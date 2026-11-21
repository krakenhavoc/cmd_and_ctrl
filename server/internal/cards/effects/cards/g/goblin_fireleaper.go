package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goblin Fireleaper — Creature — Goblin Warrior {1}{R}, 1/1 (The Hobbit
// Eternal, HOC):
//
//	"{1}{R}: This creature gets +1/+0 until end of turn.
//	 When this creature dies, it deals damage equal to its power to
//	 target creature an opponent controls."
//
// The pump is Akroma, Angel of Fury's firebreathing at a different price.
// The dies trigger is Infernal Phantom's: the target is chosen as the
// trigger goes on the stack (eerieDiesWithPower), and the damage is the
// power the Fireleaper had as it died (CR 603.10, last-known information,
// pumps and counters included), dealt by that last-known object rather than
// by the card now in the graveyard (CR 608.2h). A target that has gone
// by resolution is skipped (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c5fa1631-8685-49e1-baad-86ef72a9499c",
		Name:         "Goblin Fireleaper",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{R}: This creature gets +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{1}{R}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.BoostUntilEOT{Target: ctx.Source(), Power: 1, Label: "Goblin Fireleaper — +1/+0"}.Apply(ctx)
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.EerieDiesWithPower("Goblin Fireleaper — it deals damage equal to its power to target creature an opponent controls",
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls()),
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					ref, ok := ctx.SourceRef()
					if !ok {
						return nil
					}
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard {
							continue
						}
						return effects.DealDamage{SourceObject: &ref, Target: t.ID, Amount: item.Params.Amount}.Apply(ctx)
					}
					return nil
				}),
		},
	})
}
