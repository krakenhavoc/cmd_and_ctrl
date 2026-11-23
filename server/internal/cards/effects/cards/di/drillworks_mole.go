package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Drillworks Mole — Artifact Creature — Mole {1}, 1/1 (Commander Legends:
// Battle for Baldur's Gate, #5948):
//
//	"{2}, {T}: Put a +1/+1 counter on this creature and a +1/+1 counter on up
//	 to one target commander creature you control."
//
// The Mole always gets its counter (if it is still on the battlefield); the
// target is optional, chosen as the ability is put on the stack, and must be
// a creature you control that is a commander (CR 903.3). A commander that
// became an illegal target is skipped (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9fa3286b-ee10-4f68-bfc3-974cdc39d986",
		Name:         "Drillworks Mole",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}: Put a +1/+1 counter on this creature and a +1/+1 counter on up to one target commander creature you control.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
			Targets: effects.UpToOneTargetCreature("up to one target commander creature you control", effects.YouControl(), effects.ACommander()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if effects.OnBattlefield(g, item.SourceCardID) {
					if err := (effects.AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
						return err
					}
				}
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetCard {
						if err := (effects.AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
							return err
						}
					}
				}
				return nil
			},
		}},
	})
}
