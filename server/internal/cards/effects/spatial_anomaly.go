package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spatial Anomaly — Land (Star Trek, #4973):
//
//	"{T}: Add {C}.
//	 {4}{R}, {T}, Sacrifice this land: It deals 2 damage to target
//	 creature. Create a tapped Planet land token named New Planet with
//	 "{T}: Add one mana of any color.""
//
// The land is sacrificed as a cost, so "it" deals the damage from the
// graveyard by last known information (CR 608.2h). A target gone
// illegal in response means the ability doesn't resolve at all (CR
// 608.2b), so no New Planet either.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a4f31512-a91c-42e0-9929-5069386c3c71",
		Name:         "Spatial Anomaly",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{4}{R}, {T}, Sacrifice this land: It deals 2 damage to target creature. Create a tapped Planet land token named New Planet.",
			Cost:    trekB29AnomalyCost("{4}{R}"),
			Targets: TargetCreature("target creature"),
			Purpose: ForTargets(DamageToTarget(0, 2)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if err := b33DamageChosenTargetFromSource(2)(g, item); err != nil {
					return err
				}
				return trekB29NewPlanet(NewContext(g, item))
			},
		}},
	})
}
