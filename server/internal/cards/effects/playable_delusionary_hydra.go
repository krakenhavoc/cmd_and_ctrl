package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

const (
	playableDelusionaryHydraLoot   = "playable-delusionary-hydra/tap-loot"
	playableDelusionaryHydraAttack = "playable-delusionary-hydra/attacks-life-draw"
)

// Playable Delusionary Hydra — Stickers (an Unfinity sticker sheet, CR 123.2; ADR 0208, #4307):
//
//	Name stickers "Playable", "Delusionary", "Hydra"; three art stickers.
//	{TK}{TK} — {T}: Draw a card, then discard a card.
//	{TK}{TK}{TK}{TK} — Whenever this creature attacks, you gain 3 life and draw a card.
//	{TK}{TK} — 1/5
//	{TK}{TK}{TK} — 4/4
//
// The first sticker gives a loot ability that taps the permanent. The second gives an attack trigger that gains 3 life and draws a card.
//
// The name, power and toughness, and art stickers are the rules' (CR 123.6, 123.8, 123.9).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "065cdf8d-6874-4ab6-a08e-79bd88b245bd",
		Name:         "Playable Delusionary Hydra",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{
			{
				Key: playableDelusionaryHydraLoot,
				Activated: []ActivatedAbility{{
					Label:   "{T}: Draw a card, then discard a card.",
					Purpose: game.Purpose{Answers: game.AnswerValue},
					Cost:    TapCost(),
					Effect:  func(g *game.Game, item *game.StackItem) error { return lootOne(g, item, 1) },
				}},
				Text: "{T}: Draw a card, then discard a card.",
			},
			{
				Key: playableDelusionaryHydraAttack,
				Triggered: []game.TriggeredAbility{
					WheneverThisAttacks("Playable Delusionary Hydra sticker — you gain 3 life and draw a card",
						Do(GainLife{Amount: 3}, DrawCards{N: 1})),
				},
				Text: "Whenever this creature attacks, you gain 3 life and draw a card.",
			},
		},
		StickerSheet: &StickerSheetSpec{Abilities: []StickerAbilitySpec{
			{Grant: playableDelusionaryHydraLoot},
			{Grant: playableDelusionaryHydraAttack},
		}},
	})
}
