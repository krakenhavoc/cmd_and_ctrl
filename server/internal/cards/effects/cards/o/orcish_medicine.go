package cardso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Orcish Medicine — Instant {1}{B}:
//
//	"Target creature gains your choice of lifelink or indestructible
//	 until end of turn.
//	 Amass Orcs 1."
//
// The choice is made as the spell resolves (Lunar Avenger's PickOption
// shape), and the amass is the printed second sentence, so it runs in
// the choice's Then: Apply only queues the prompt, and anything printed
// after the choice has to wait for the answer. A target that is illegal
// on resolution makes the whole spell fizzle (CR 608.2b), the amass
// included, as printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2f9a5e7c-f463-4773-bf67-a07339ce9b5d",
		Name:         "Orcish Medicine",
		Purpose:      effects.WithRiders(effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0)), game.Purpose{Amass: 1}),
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			target := game.TargetRef{}
			for _, t := range ctx.LegalTargets() {
				target = t
				break
			}
			return effects.PickOption{
				Question: "Orcish Medicine — the creature gains which until end of turn?",
				Options:  []game.ChoiceOption{{Label: "Lifelink"}, {Label: "Indestructible"}},
				ThenKey:  orcishMedicineKeyword,
				Carry:    []uuid.UUID{target.ID},
			}.Apply(ctx)
		},
	})
}

// orcishMedicineKeyword is the keyword chosen, then the Army, keyed (#4253):
// the target rides the prompt as data.
var orcishMedicineKeyword = effects.OptionPickThen("option-pick/orcish-medicine-keyword", func(ctx *effects.Context, r game.OptionPicked) error {
	keywords := []string{"lifelink", "indestructible"}
	if r.Index >= 0 && r.Index < len(keywords) && len(r.Carry) > 0 && r.Carry[0] != uuid.Nil {
		if err := (effects.GrantKeywordUntilEOT{
			Target:   r.Carry[0],
			Keywords: []string{keywords[r.Index]},
			Label:    "Orcish Medicine — " + keywords[r.Index],
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return effects.Amass{Subtype: "Orc", N: 1}.Apply(ctx)
})
