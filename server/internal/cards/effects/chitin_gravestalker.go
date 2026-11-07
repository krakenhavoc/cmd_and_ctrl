package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Chitin Gravestalker — Creature — Insect Warrior {5}{B}, 5/4
// (Aetherdrift, #4464):
//
//	"This spell costs {1} less to cast for each artifact and/or creature
//	 card in your graveyard.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// Ravenous Cosmozoan's discount (CR 601.2f) counting Aatchik's cards
// (dftABArtifactOrCreatureCardsInGraveyard): an artifact creature card is
// one card. The caster's own graveyard is counted as the cost is
// determined, and only the generic part is reduced, so {B} is always
// paid. Cycling is the hand ability (CR 702.29a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "acac15c8-877d-49d4-9b42-90181c05a684",
		Name:         "Chitin Gravestalker",
		Completeness: CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			CostsLessEach(func(q game.CostQuery) int {
				if q.Game == nil {
					return 0
				}
				return dftABArtifactOrCreatureCardsInGraveyard(q.Game, q.Controller)
			}, "This spell costs {1} less to cast for each artifact and/or creature card in your graveyard."),
		},
		Activated: []ActivatedAbility{Cycling("{2}")},
	})
}
