package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Akoum Hellkite — Creature — Dragon {4}{R}{R}, 4/4 (Commander Starter
// Decks):
//
//	"Flying
//	 Landfall — Whenever a land you control enters, this creature deals 1
//	 damage to any target. If that land is a Mountain, this creature deals 2
//	 damage instead."
//
// Landfall with a target clause. The Mountain check reads the entering land
// as the ability resolves (Koth's RfCreatureCEnteredLandHasSubtype, which
// follows it even if it has left the battlefield since), through the text
// reader so a changed land type is honoured. The damage is the Hellkite's
// own, and nothing is dealt if the target has become illegal (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "05b6f9e3-9acd-43a6-acff-caa811aaf0a1",
		Name:            "Akoum Hellkite",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.Landfall("Akoum Hellkite — 1 damage to any target, 2 if that land is a Mountain",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					amount := 1
					if effects.RfCreatureCEnteredLandHasSubtype(g, item, g.TextReaderForEffect(item).TextSubtype("Mountain")) {
						amount = 2
					}
					for _, t := range ctx.LegalTargets() {
						return effects.DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: amount}.Apply(ctx)
					}
					return nil
				}), effects.TargetAny()),
		},
	})
}
