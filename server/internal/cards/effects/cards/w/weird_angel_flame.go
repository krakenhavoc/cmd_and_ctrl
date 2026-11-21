package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

const weirdAngelFlameHeroic = "weird-angel-flame/heroic-counters"

// Weird Angel Flame — Stickers (an Unfinity sticker sheet, CR 123.2; ADR 0208, #4307):
//
//	Name stickers "Weird", "Angel", "Flame"; three art stickers.
//	{TK}{TK} — Heroic — Whenever you cast a spell that targets this permanent, put two +1/+1 counters on it.
//	{TK}{TK}{TK} — Protection from even mana values
//	{TK}{TK} — 2/3
//	{TK}{TK}{TK}{TK}{TK} — 7/8
//
// The first sticker gives heroic (CR 207.2c, an ability word): whenever its controller casts a spell that targets the permanent, two +1/+1 counters go on it. The second gives protection from even mana values (CR 702.16, 202.3): a source whose mana value is even, a spell counting the X it was cast with and a token that is no copy being 0.
//
// The name, power and toughness, and art stickers are the rules' (CR 123.6, 123.8, 123.9).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2c925531-63f8-4a6b-bfe0-bb8cd5d0d63d",
		Name:         "Weird Angel Flame",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{
			{
				Key: weirdAngelFlameHeroic,
				Triggered: []game.TriggeredAbility{
					effects.Heroic("Weird Angel Flame sticker — put two +1/+1 counters on this permanent (heroic)",
						effects.ThisGetsTwoPlusOneCounters),
				},
				Text: "Heroic — Whenever you cast a spell that targets this permanent, put two +1/+1 counters on it.",
			},
		},
		StickerSheet: &effects.StickerSheetSpec{Abilities: []effects.StickerAbilitySpec{
			{Grant: weirdAngelFlameHeroic},
			{Keywords: []string{"protection from even mana values"}},
		}},
	})
}
