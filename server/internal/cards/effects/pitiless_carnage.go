package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Pitiless Carnage — Sorcery {3}{B} (Outlaws of Thunder Junction,
// #4468):
//
//	"Sacrifice any number of permanents you control, then draw that
//	 many cards.
//	 Plot {1}{B}{B} (You may pay {1}{B}{B} and exile this card from your
//	 hand. Cast it as a sorcery on a later turn without paying its mana
//	 cost. Plot only as a sorcery.)"
//
// God-Eternal Bontu's pick without "other": one prompt over every
// permanent you control, lands and tokens included, from none to all of
// them. You draw one card for each permanent sacrificed. Plot is
// Beastbond Outcaster's special action (CR 702.170a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "186bfdac-ce9c-476f-89b6-ef467487fca5",
		Name:           "Pitiless Carnage",
		Completeness:   CompletenessFull,
		SpecialActions: []game.SpecialAction{Plot("{1}{B}{B}")},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return ChoosePermanents{
				Question: "Pitiless Carnage — sacrifice any number of permanents you control",
				Keyed:    pitilessCarnageSacrificed,
				Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
					var out []uuid.UUID
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller == of {
							out = append(out, c.InstanceID)
						}
					}
					return out, 0, 0 // "any number": no floor, no ceiling
				},
			}.Apply(ctx)
		},
	})
}

var pitilessCarnageSacrificed = RegisterPermanentsPick("prompt-run/pitiless-carnage-sacrifice", PermanentsDef{
	Sacrifice: true,
	Then: func(ctx *Context, picked game.PromptedPicks) error {
		return drawOnePerPicked(ctx.Controller())(ctx, picked)
	},
})
