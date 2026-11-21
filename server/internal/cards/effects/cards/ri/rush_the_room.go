package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rush the Room — Instant {R} (The Lord of the Rings: Tales of
// Middle-earth, #5948):
//
//	"Target creature gets +1/+0 and gains first strike until end of turn. If
//	 it's a Goblin or Orc, it also gains haste until end of turn."
//
// The creature types are read as the spell resolves, through the layers, so an
// Army that amass made an Orc counts.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "223fefa3-64a7-457f-90a2-89973640f554",
		Name:         "Rush the Room",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 1, 0, "first strike")),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			legal := ctx.LegalTargets()
			if len(legal) == 0 || legal[0].Kind != game.TargetCard {
				return nil
			}
			id := legal[0].ID
			if err := (effects.BoostUntilEOT{Target: id, Power: 1, Label: "Rush the Room — +1/+0"}).Apply(ctx); err != nil {
				return err
			}
			keywords := []string{"first strike"}
			if c, ok := ctx.Game.LookupCardForEffect(id); ok && (c.HasSubtypeWord("Goblin") || c.HasSubtypeWord("Orc")) {
				keywords = append(keywords, "haste")
			}
			return effects.GrantKeywordUntilEOT{Target: id, Keywords: keywords, Label: "Rush the Room — first strike"}.Apply(ctx)
		},
	})
}
