package cardsbl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Blizzard — Enchantment {G}{G} (Ice Age, S82 OLD batch 838, #5636):
//
//	"Cast this spell only if you control a snow land.
//	 Cumulative upkeep {2}
//	 Creatures with flying don't untap during their controllers' untap steps."
//
// The cast condition reads the caster's battlefield when the spell is cast
// (CR 601.2). The untap lock is Back to Basics' with a flying test, read through
// the layered view (CR 502.3).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "698feaaa-2fff-4809-81d1-4caad44c6c0a",
		Name:         "Blizzard",
		Completeness: effects.CompletenessFull,
		CastCondition: func(g *game.Game, player uuid.UUID, _ game.Card) bool {
			for i := range g.Battlefield.Cards {
				if c := g.Battlefield.Cards[i]; c.Controller == player && c.IsLand() && c.HasSupertype("snow") {
					return true
				}
			}
			return false
		},
		CastConditionLabel:    "Cast this spell only if you control a snow land.",
		Triggered:             []game.TriggeredAbility{effects.CumulativeUpkeep("Blizzard — cumulative upkeep {2}", "{2}")},
		UntapStepRestrictions: []game.UntapStepRestriction{effects.DoesntUntapDuringTheirControllersUntapSteps(effects.And(effects.Creature(), effects.HasKeyword("flying")))},
	})
}
