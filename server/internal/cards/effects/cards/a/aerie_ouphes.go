package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aerie Ouphes — Creature — Ouphe {4}{G}, 3/3:
//
//	"Sacrifice this creature: It deals damage equal to its power to
//	 target creature with flying.
//	 Persist"
//
// The sacrifice is the cost, so "its power" is the power it had as it
// last existed on the battlefield (CR 608.2h): 3, then 2 once persist
// has returned it with a -1/-1 counter. Persist is PrintedKeywords
// (#2075).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "70cdd24a-8dc5-47b5-8729-9ebf486f4821",
		Name:            "Aerie Ouphes",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordPersist},
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice this creature: It deals damage equal to its power to target creature with flying.",
			Cost:    effects.SacrificeThis(),
			Targets: effects.TargetCreature("target creature with flying", effects.HasKeyword("flying")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				info, ok := ctx.SourcePermanent()
				ref, refOK := ctx.SourceRef()
				legal := ctx.LegalTargets()
				if !ok || !refOK || info.Power <= 0 || len(legal) == 0 {
					return nil
				}
				return effects.DealDamage{SourceObject: &ref, Target: legal[0].ID, Amount: info.Power}.Apply(ctx)
			},
		}},
	})
}
