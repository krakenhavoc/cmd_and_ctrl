package cardsb

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Barge In — Instant {R} (Throne of Eldraine, #5516):
//
//	"Target attacking creature gets +2/+2 until end of turn. Each attacking
//	 non-Human creature gains trample until end of turn."
//
// The target clause is "attacking" (checked at announce and resolution). The
// trample set is fixed when the spell resolves (CR 611.2c) and is every
// attacking non-Human creature, yours or not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "32eb5d30-f4e7-48ef-b5f6-051ddddd4175",
		Name:         "Barge In",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2)),
		Targets:      effects.TargetCreature("target attacking creature", effects.AttackingCreature()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if id, ok := effects.B16FirstLegalTargetCard(ctx); ok {
				if err := (effects.BoostUntilEOT{Target: id, Power: 2, Toughness: 2, Label: "Barge In — +2/+2"}).Apply(ctx); err != nil {
					return err
				}
			}
			word := ctx.TextSubtype("Human")
			return effects.GrantKeywordUntilEOT{
				Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
					return c.IsCreature() && c.AttackingTarget != uuid.Nil && !c.HasSubtype(word)
				},
				Keywords: []string{"trample"},
				Label:    "Barge In — attacking non-Human creatures gain trample",
			}.Apply(ctx)
		},
	})
}
