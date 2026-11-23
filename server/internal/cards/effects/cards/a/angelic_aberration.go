package cardsa

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Angelic Aberration — Creature — Eldrazi Angel {5}{W}, 4/4 (Modern
// Horizons 3 Commander, #5948):
//
//	"Devoid (This card has no color.)
//	 Flying, vigilance
//	 When this creature enters, sacrifice any number of creatures each with
//	 base power or toughness 1 or less. Create that many 4/4 colorless
//	 Eldrazi Angel creature tokens with flying and vigilance."
//
// The sacrifice is a choice made as the trigger resolves (CR 608.2c, no
// target): every creature its controller controls whose BASE power or
// toughness is 1 or less (CR 208.4b) is a candidate, zero to all of
// them, and they leave as one simultaneous exit. One token is made for
// each creature actually sacrificed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "60d0c941-39f6-4144-abef-02b0a84d602e",
		Name:            "Angelic Aberration",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid, "flying", "vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Angelic Aberration — sacrifice any number of creatures each with base power or toughness 1 or less, then create that many 4/4 Eldrazi Angels",
				angelicAberrationSacrifice),
		},
	})
}

// angelicAberrationSacrificed is the chosen creatures sacrificed and a
// 4/4 Eldrazi Angel made for each that really left.
var angelicAberrationSacrificed = effects.RegisterPermanentsPick("prompt-run/angelic-aberration-sacrifice", effects.PermanentsDef{
	Sacrifice: true,
	Then: func(ctx *effects.Context, picked game.PromptedPicks) error {
		n := picked.Count()
		if n == 0 {
			return nil
		}
		return effects.CreateToken{Template: effects.TokenCard("4/4 colorless Eldrazi Angel with flying and vigilance"), N: n}.Apply(ctx)
	},
})

func angelicAberrationSacrifice(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	return effects.ChoosePermanents{
		Question: "Angelic Aberration — sacrifice any number of creatures each with base power or toughness 1 or less",
		Keyed:    angelicAberrationSacrificed,
		Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
			var out []uuid.UUID
			for _, c := range g.BattlefieldCardsForEffect() {
				if c.Controller == of && c.IsCreature() && (c.BasePower() <= 1 || c.BaseToughness() <= 1) {
					out = append(out, c.InstanceID)
				}
			}
			return out, 0, 0 // "any number"
		},
	}.Apply(ctx)
}
