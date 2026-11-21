package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "07fb2903-b239-43d0-864b-c4437b7a5230",
		Name:            "Mirkwood Spider",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			Targeting(On(game.EventAttack, Self, mirkwoodSpiderLabel,
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					id, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return GrantKeywordUntilEOT{Target: id, Keywords: []string{"deathtouch"}, Label: mirkwoodSpiderLabel}.Apply(ctx)
				}),
				TargetCreature("target legendary creature you control", YouControl(), Legendary())),
		},
	})
}
