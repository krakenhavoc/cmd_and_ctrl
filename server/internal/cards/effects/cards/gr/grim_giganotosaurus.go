package cardsgr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Grim Giganotosaurus — Creature — Dinosaur {5}{B}{G}, 10/10 (Jurassic World
// Collection, #5948):
//
//	"{10}{B}{G}: Monstrosity 10. This ability costs {1} less to activate for
//	 each creature with power 4 or greater your opponents control.
//	 When this creature becomes monstrous, destroy all artifacts and
//	 creatures other than this creature."
//
// Battlefield Butcher's cost-modifier shape on the shared Monstrosity
// ability (the reduction never touches the coloured {B}{G}). The trigger is
// one destroy over the artifact-or-creature class minus the Giganotosaurus
// itself; regeneration and indestructible apply as usual.
//
// No simplification.
func init() {
	mons := effects.Monstrosity(effects.ManaCost("{10}{B}{G}"), 10)
	mons.Label = "{10}{B}{G}: Monstrosity 10. This ability costs {1} less to activate for each creature with power 4 or greater your opponents control."
	mons.CostModifiers = []game.CostModifier{
		effects.CostsLessEach(effects.GapB82OpponentCreaturesWithPower4,
			"This ability costs {1} less to activate for each creature with power 4 or greater your opponents control."),
	}
	effects.Register(effects.Spec{
		OracleID:     "1c197b3c-f540-4766-aeaa-ae09d22214cb",
		Name:         "Grim Giganotosaurus",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{mons},
		Triggered: []game.TriggeredAbility{
			effects.WhenBecomesMonstrous("Grim Giganotosaurus — destroy all artifacts and creatures other than this creature",
				func(g *game.Game, item *game.StackItem) error {
					self := item.SourceCardID
					return effects.DestroyAllMatching{Match: effects.And(effects.Or(effects.Artifact(), effects.Creature()), func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.InstanceID != self })}.Apply(effects.NewContext(g, item))
				}),
		},
		Purpose: game.Purpose{Sweep: game.Sweep{Matches: game.SweepNonlandPermanents, How: game.SweepDestroy, Partial: true}},
	})
}
