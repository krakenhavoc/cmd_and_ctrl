package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

const (
	sassyGremlinBloodTreasure = "sassy-gremlin-blood/attacks-treasure"
	sassyGremlinBloodFlying   = "sassy-gremlin-blood/give-flying"
)

// Sassy Gremlin Blood — Stickers (an Unfinity sticker sheet, CR 123.2; ADR 0208, #4307):
//
//	Name stickers "Sassy", "Gremlin", "Blood"; three art stickers.
//	{TK}{TK} — Whenever this creature attacks, create a Treasure token.
//	{TK}{TK}{TK}{TK}{TK} — {3}: Target creature gains flying until end of turn.
//	{TK}{TK} — 3/2
//	{TK}{TK}{TK}{TK}{TK}{TK} — 10/10
//
// The first sticker gives an attack trigger that creates a Treasure. The second gives an ability that grants flying to a target creature.
//
// The name, power and toughness, and art stickers are the rules' (CR 123.6, 123.8, 123.9).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "16663933-c8e2-4411-8eed-673c52fa3ecb",
		Name:         "Sassy Gremlin Blood",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{
			{
				Key: sassyGremlinBloodTreasure,
				Triggered: []game.TriggeredAbility{
					effects.WheneverThisAttacks("Sassy Gremlin Blood sticker — create a Treasure token", effects.Do(effects.CreateToken{Template: effects.TreasureToken(), N: 1})),
				},
				Text: "Whenever this creature attacks, create a Treasure token.",
			},
			{
				Key: sassyGremlinBloodFlying,
				Activated: []effects.ActivatedAbility{{
					Label:   "{3}: Target creature gains flying until end of turn.",
					Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "flying")),
					Cost:    effects.ManaCost("{3}"),
					Targets: effects.TargetCreature("target creature"),
					Effect:  effects.EbTargetCreatureGainsUntilEOT("flying", "Sassy Gremlin Blood sticker — flying until end of turn"),
				}},
				Text: "{3}: Target creature gains flying until end of turn.",
			},
		},
		StickerSheet: &effects.StickerSheetSpec{Abilities: []effects.StickerAbilitySpec{
			{Grant: sassyGremlinBloodTreasure},
			{Grant: sassyGremlinBloodFlying},
		}},
	})
}
