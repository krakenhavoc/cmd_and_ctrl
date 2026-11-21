package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// It That Heralds the End — Creature — Eldrazi Drone {1}{C}, 2/2 (Modern
// Horizons 3, #5948):
//
//	"Colorless spells you cast with mana value 7 or greater cost {1} less to
//	 cast.
//	 Other colorless creatures you control get +1/+1."
//
// The discount is Eye of Ugin's generic-only reduction (CostsLess) read off the
// spell being cast: it must have no colour and mana value 7 or greater, and the
// reduction can't take the generic part below zero (CR 601.2f). The anthem is
// Kozilek, the Broken Reality's, narrowed to +1/+1: it reads layer 7c after
// layer 5, so a face-down 2/2 is colourless and a creature and gets it, a
// colour-changed creature does not, and this creature itself ("other") never
// does.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c14ac951-329d-452b-be75-5c72166cd9c7",
		Name:         "It That Heralds the End",
		Completeness: CompletenessFull,
		CostModifiers: []game.CostModifier{
			CostsLess(1, "Colorless spells you cast with mana value 7 or greater cost {1} less to cast.",
				YourSpell(), func(q game.CostQuery) bool {
					if !q.Card.IsColorless() {
						return false
					}
					mv, ok := q.Game.ManaValueForEffect(q.Card)
					return ok && mv >= 7
				}),
		},
		Static: []game.StaticAbility{
			Labelled(b16Anthem(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller &&
					target.IsColorless() && target.InstanceID != source.InstanceID
			}, 1, 1), "Other colorless creatures you control get +1/+1."),
		},
	})
}
