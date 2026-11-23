package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

const mirkwoodSpiderLabel = "Mirkwood Spider — target legendary creature you control gains deathtouch until end of turn"

// Mirkwood Spider — Creature — Spider {G}, 1/1 (The Lord of the Rings:
// Tales of Middle-earth Commander):
//
//	"Deathtouch
//	 Whenever this creature attacks, target legendary creature you control
//	 gains deathtouch until end of turn."
//
// Deathtouch is a printed keyword. The attack trigger is Aphelia's
// self-attack watch with a legendary-creature target; the Spider may
// target itself only if it is legendary. A target that left in response
// gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "07fb2903-b239-43d0-864b-c4437b7a5230",
		Name:            "Mirkwood Spider",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventAttack, effects.Self, mirkwoodSpiderLabel,
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"deathtouch"}, Label: mirkwoodSpiderLabel}.Apply(ctx)
				}),
				effects.TargetCreature("target legendary creature you control", effects.YouControl(), effects.Legendary())),
		},
	})
}
