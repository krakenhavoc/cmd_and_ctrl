package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "16663933-c8e2-4411-8eed-673c52fa3ecb",
		Name:         "Sassy Gremlin Blood",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{
			{
				Key: sassyGremlinBloodTreasure,
				Triggered: []game.TriggeredAbility{
					WheneverThisAttacks("Sassy Gremlin Blood sticker — create a Treasure token", Do(CreateToken{Template: TreasureToken(), N: 1})),
				},
				Text: "Whenever this creature attacks, create a Treasure token.",
			},
			{
				Key: sassyGremlinBloodFlying,
				Activated: []ActivatedAbility{{
					Label:   "{3}: Target creature gains flying until end of turn.",
					Purpose: ForTargets(BenefitUntilEOT(0, 0, 0, "flying")),
					Cost:    ManaCost("{3}"),
					Targets: TargetCreature("target creature"),
					Effect:  ebTargetCreatureGainsUntilEOT("flying", "Sassy Gremlin Blood sticker — flying until end of turn"),
				}},
				Text: "{3}: Target creature gains flying until end of turn.",
			},
		},
		StickerSheet: &StickerSheetSpec{Abilities: []StickerAbilitySpec{
			{Grant: sassyGremlinBloodTreasure},
			{Grant: sassyGremlinBloodFlying},
		}},
	})
}
