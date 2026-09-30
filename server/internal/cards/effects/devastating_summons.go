package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Devastating Summons — Sorcery {R}:
//
//	"As an additional cost to cast this spell, sacrifice X lands.
//	 Create two X/X red Elemental creature tokens."
//
// "Sacrifice X" as a cast's additional cost (ADR 0100 §3): X is
// announced with the cast (CR 107.3a) — the mana cost prints no {X} —
// and the caster names exactly X lands, sacrificed with the spell on
// the stack. The tokens read the same X back at resolution (CR 107.3i).
// X = 0 is legal and makes two 0/0s that die to the next state-based
// check, as in paper.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "5eb6626a-1e0a-438e-9e95-a1e86be6489d",
		Name:           "Devastating Summons",
		Completeness:   CompletenessFull,
		XMatters:       true,
		AdditionalCost: SacrificeXCost("X lands", Land()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			x := ctx.X()
			return CreateToken{
				Controller: item.Controller,
				Template: game.Card{
					Name:      "Elemental",
					TypeLine:  "Token Creature — Elemental",
					Colors:    []string{"R"},
					Power:     x,
					Toughness: x,
				},
				N: 2,
			}.Apply(ctx)
		},
	})
}
