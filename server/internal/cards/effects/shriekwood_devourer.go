package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Shriekwood Devourer — Creature — Treefolk {5}{G}{G}, 7/5 (Duskmourn
// Commander, #5368):
//
//	"Trample
//	 Whenever you attack with one or more creatures, untap up to X lands, where X
//	 is the greatest power among those creatures."
//
// One trigger per attack declaration (a "one or more" ability, OncePerBatch),
// whatever the number of attackers. X is the greatest power among the creatures
// you declared as attackers, read as the trigger resolves; "up to X lands" is
// the shared UntapUpToLands pick (any player's lands, as printed: there is no
// "you control").
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "56314980-5f18-4b93-a9a0-6002ee8a880c",
		Name:            "Shriekwood Devourer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventAttack,
				func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return source != nil && attackDeclaredByYou(ev, source.Controller)
				},
				"Shriekwood Devourer — untap up to X lands, where X is the greatest power among the attacking creatures",
				func(g *game.Game, item *game.StackItem) error {
					return UntapUpToLands{
						N:        shriekwoodGreatestAttackerPower(g, item.Controller),
						Question: "Shriekwood Devourer — untap up to X lands",
					}.Apply(NewContext(g, item))
				})),
		},
	})
}

// shriekwoodGreatestAttackerPower is the greatest power among the creatures
// `controller` declared as attackers this combat that are still on the
// battlefield; zero when none is.
func shriekwoodGreatestAttackerPower(g *game.Game, controller uuid.UUID) int {
	declared := map[uuid.UUID]bool{}
	for _, ref := range g.AttackersThisCombatForEffect() {
		declared[ref.ID] = true
	}
	best := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != controller || !declared[c.InstanceID] {
			continue
		}
		if p := c.CurrentPower(); p > best {
			best = p
		}
	}
	return best
}
