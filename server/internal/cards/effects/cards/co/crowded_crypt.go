package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crowded Crypt — Artifact {2}{B}:
//
//	"{T}: Add {B}.
//	 Whenever a creature you control dies, put a corpse counter on this
//	 artifact.
//	 {4}{B}{B}, {T}, Sacrifice this artifact: Create a 2/2 black Zombie
//	 creature token with decayed for each corpse counter on this
//	 artifact."
//
// The count is read as the ability resolves from the Crypt as it last
// existed on the battlefield (CountersOnThisAsItResolves, CR 608.2h):
// it was sacrificed to pay, so its corpse counters are last-known
// information. The death trigger adds a counter only while the Crypt
// is still on the battlefield to hold it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "25712827-626a-484d-bc52-f3bd6087857f",
		Name:         "Crowded Crypt",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{B}",
			Label:    "Add {B}",
		}},
		Triggered: []game.TriggeredAbility{
			effects.WheneverACreatureYouControlDies("Crowded Crypt — put a corpse counter on it",
				func(g *game.Game, item *game.StackItem) error {
					if !effects.OnBattlefield(g, item.SourceCardID) {
						return nil
					}
					return effects.AddCounter{Target: item.SourceCardID, Kind: "corpse", N: 1}.Apply(effects.NewContext(g, item))
				}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}{B}{B}, {T}, Sacrifice this artifact: Create a 2/2 black Zombie creature token with decayed for each corpse counter on this artifact.",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Cost:    effects.Plus(effects.ManaCost("{4}{B}{B}"), effects.TapCost(), effects.SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				n := effects.CountersOnThisAsItResolves("corpse")(ctx)
				return effects.CreateToken{Template: effects.TokenCard("2/2 black Zombie with decayed"), N: n}.Apply(ctx)
			},
		}},
	})
}
