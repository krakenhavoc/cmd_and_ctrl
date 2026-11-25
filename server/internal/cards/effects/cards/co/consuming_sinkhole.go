package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Consuming Sinkhole — {3}{R} Instant (Oath of the Gatewatch, S82 OLD b295):
//
//	"Devoid (This card has no color.)
//	 Choose one —
//	 • Exile target land creature.
//	 • Consuming Sinkhole deals 4 damage to target player or planeswalker."
//
// A two-mode charm with its own target per mode (Temur Charm, Boros Charm).
// "Land creature" is a creature that is also a land, read through the layers;
// the damage is Boros Charm's four to a player or planeswalker. A mode's
// target is judged again as the spell resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "03fde598-5b6b-4c7f-ba13-c06ed1fa918a",
		Name:            "Consuming Sinkhole",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.ModeDoing("Exile target land creature.",
				effects.TargetCreature("target land creature", effects.Land()),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					for _, t := range ctx.ModeTargets(occ) {
						if t.Kind == game.TargetCard && ctx.IsTargetLegal(t) {
							return effects.ExileTarget{Target: t.ID}.Apply(ctx)
						}
					}
					return nil
				}), game.Purpose{Answers: game.AnswerRemove}),
			effects.ModeWithPurpose(effects.ModeDoing("Consuming Sinkhole deals 4 damage to target player or planeswalker.",
				effects.TargetPlayerOrPlaneswalker(),
				func(item *game.StackItem, ctx *effects.Context, occ int) error {
					for _, t := range ctx.ModeTargets(occ) {
						if ctx.IsTargetLegal(t) {
							return effects.DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: 4}.Apply(ctx)
						}
					}
					return nil
				}), effects.ForTargets(effects.DamageToTarget(0, 4))),
		),
	})
}
