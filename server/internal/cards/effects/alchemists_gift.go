package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Alchemist's Gift — {B} Instant:
//
//	"Target creature gets +1/+1 and gains your choice of deathtouch or
//	lifelink until end of turn."
//
// The pump lands first; the keyword is chosen as the spell resolves.
func init() {
	Register(Spec{
		OracleID:     "4e698713-6203-44f2-b438-1b26b951536d",
		Name:         "Alchemist's Gift",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(BenefitUntilEOT(0, 1, 1, "deathtouch")),
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
				return nil
			}
			if err := (BoostUntilEOT{Target: t.ID, Power: 1, Toughness: 1, Label: "Alchemist's Gift — +1/+1"}).Apply(ctx); err != nil {
				return err
			}
			return PickOption{
				Question: "Alchemist's Gift — the creature gains which until end of turn?",
				Options:  []game.ChoiceOption{{Label: "Deathtouch"}, {Label: "Lifelink"}},
				ThenKey:  alchemistsGiftKeyword,
				Carry:    []uuid.UUID{t.ID},
			}.Apply(ctx)
		},
	})
}

var alchemistsGiftKeywords = []string{"deathtouch", "lifelink"}

var alchemistsGiftKeyword = OptionPickThen("option-pick/s82-bns-b45-alchemists-gift-keyword", func(ctx *Context, r game.OptionPicked) error {
	if r.Index < 0 || r.Index >= len(alchemistsGiftKeywords) || len(r.Carry) == 0 {
		return nil
	}
	return GrantKeywordUntilEOT{Target: r.Carry[0], Keywords: []string{alchemistsGiftKeywords[r.Index]}, Label: "Alchemist's Gift — keyword"}.Apply(ctx)
})
