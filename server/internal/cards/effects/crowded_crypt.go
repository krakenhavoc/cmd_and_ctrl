package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "25712827-626a-484d-bc52-f3bd6087857f",
		Name:         "Crowded Crypt",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{B}",
			Label:    "Add {B}",
		}},
		Triggered: []game.TriggeredAbility{
			WheneverACreatureYouControlDies("Crowded Crypt — put a corpse counter on it",
				func(g *game.Game, item *game.StackItem) error {
					if !onBattlefield(g, item.SourceCardID) {
						return nil
					}
					return AddCounter{Target: item.SourceCardID, Kind: "corpse", N: 1}.Apply(NewContext(g, item))
				}),
		},
		Activated: []ActivatedAbility{{
			Label:   "{4}{B}{B}, {T}, Sacrifice this artifact: Create a 2/2 black Zombie creature token with decayed for each corpse counter on this artifact.",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Cost:    Plus(ManaCost("{4}{B}{B}"), TapCost(), SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				n := CountersOnThisAsItResolves("corpse")(ctx)
				return CreateToken{Template: TokenCard("2/2 black Zombie with decayed"), N: n}.Apply(ctx)
			},
		}},
	})
}
