package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:     "2f9a5e7c-f463-4773-bf67-a07339ce9b5d",
		Name:         "Orcish Medicine",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			target := game.TargetRef{}
			for _, t := range ctx.LegalTargets() {
				target = t
				break
			}
			keywords := []string{"lifelink", "indestructible"}
			return PickOption{
				Question: "Orcish Medicine — the creature gains which until end of turn?",
				Options:  []game.ChoiceOption{{Label: "Lifelink"}, {Label: "Indestructible"}},
				Then: func(ctx *Context, index int) error {
					if index >= 0 && index < len(keywords) && target.ID != uuid.Nil {
						if err := (GrantKeywordUntilEOT{
							Target:   target.ID,
							Keywords: []string{keywords[index]},
							Label:    "Orcish Medicine — " + keywords[index],
						}).Apply(ctx); err != nil {
							return err
						}
					}
					return Amass{Subtype: "Orc", N: 1}.Apply(ctx)
				},
			}.Apply(ctx)
		},
	})
}
