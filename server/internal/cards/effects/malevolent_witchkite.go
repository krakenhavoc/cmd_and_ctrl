package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Malevolent Witchkite — Creature — Dragon Warlock {4}{B}{B}, 5/4 (Wilds of
// Eldraine, #4471):
//
//	"Flying
//	 When this creature enters, sacrifice any number of artifacts,
//	 enchantments, and/or tokens, then draw that many cards."
//
// The sacrifice is a choice made as the ability resolves (CR 608.2), Plunge
// into Darkness's "any number" pick over your own permanents, zero allowed.
// The draw counts the permanents that really left. Abilities that trigger on
// the sacrifice wait until after the draw, which the keyed pick's tail gives.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "aee2cd1b-9bf9-4b91-83a4-2f1dbc82dcf4",
		Name:            "Malevolent Witchkite",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Malevolent Witchkite — sacrifice any number of artifacts, enchantments, and/or tokens, then draw that many cards", Do(ChoosePermanents{
				Question: "Malevolent Witchkite — sacrifice any number of artifacts, enchantments, and/or tokens (you draw that many cards)",
				Keyed:    malevolentWitchkiteSacrificed,
				Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
					var out []uuid.UUID
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller == of && (c.IsArtifact() || c.IsEnchantment() || c.IsToken()) {
							out = append(out, c.InstanceID)
						}
					}
					return out, 0, 0
				},
			})),
		},
	})
}

// malevolentWitchkiteSacrificed is the chosen permanents sacrificed and a
// card drawn for each that really left, keyed.
var malevolentWitchkiteSacrificed = RegisterPermanentsPick("prompt-run/malevolent-witchkite-sacrifice", PermanentsDef{
	Sacrifice: true,
	Then: func(ctx *Context, picked game.PromptedPicks) error {
		return DrawCards{Player: ctx.Controller(), N: picked.Count()}.Apply(ctx)
	},
})
