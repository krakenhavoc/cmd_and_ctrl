package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Black Tom Cassidy — Legendary Creature — Mutant Villain {1}{G}, 1/3
// (Marvel Super Heroes Commander, #4455):
//
//	"{T}: Add {G}. If you control another Mutant, you gain 1 life."
//
// A mana ability with a rider (CR 605.3b): the life is gained inside the same
// resolution as the mana, and only when the controller has ANOTHER Mutant
// then, read when the ability resolves. Black Tom is a Mutant himself, so the
// rider needs a second one.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c0d20e66-deb6-4a20-a2c1-8a79514e3792",
		Name:         "Black Tom Cassidy",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{G}",
			Label:    "{T}: Add {G}. If you control another Mutant, you gain 1 life.",
			Rider: func(g *game.Game, controller, source uuid.UUID) error {
				self, ok := g.LookupCardForEffect(source)
				if !ok {
					return nil
				}
				mutant := self.TextSubtype("Mutant")
				for _, c := range g.BattlefieldCardsForEffect() {
					if c.Controller == controller && c.InstanceID != source && c.HasSubtype(mutant) {
						return g.ChangePlayerLifeForEffect(source, controller, 1)
					}
				}
				return nil
			},
		}},
	})
}
