package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

const (
	goblinCowardParadeMentor  = "goblin-coward-parade/mentor"
	goblinCowardParadeDestroy = "goblin-coward-parade/leaves-destroy"
)

// Goblin Coward Parade — Stickers (an Unfinity sticker sheet, CR 123.2; ADR 0208, #4307):
//
//	Name stickers "Goblin", "Coward", "Parade"; three art stickers.
//	{TK}{TK} — Mentor
//	{TK}{TK}{TK} — When this permanent leaves the battlefield, you may destroy target creature with power 4 or greater.
//	{TK}{TK} — 2/3
//	{TK}{TK}{TK}{TK} — 8/4
//
// Mentor (CR 702.134a) is the engine's keyword trigger (Mentor): a +1/+1 counter on a target attacking creature with lesser power. The second sticker gives a leaves-the-battlefield trigger with a "you may" and a target creature with power 4 or greater.
//
// The name, power and toughness, and art stickers are the rules' (CR 123.6, 123.8, 123.9).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7cad08c1-4bd9-4d9f-85d2-fb3ab718fbb8",
		Name:         "Goblin Coward Parade",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{
			{
				Key:       goblinCowardParadeMentor,
				Triggered: []game.TriggeredAbility{Mentor("Goblin Coward Parade sticker — mentor")},
				Text:      "Mentor (Whenever this creature attacks, put a +1/+1 counter on target attacking creature with lesser power.)",
			},
			{
				Key: goblinCowardParadeDestroy,
				Triggered: []game.TriggeredAbility{
					Optional(Targeting(
						WhenThisLeaves("Goblin Coward Parade sticker — you may destroy target creature with power 4 or greater", destroyFirstLegalCardTarget),
						TargetCreature("target creature with power 4 or greater", PowerGE(4))),
						"Destroy target creature with power 4 or greater?"),
				},
				Text: "When this permanent leaves the battlefield, you may destroy target creature with power 4 or greater.",
			},
		},
		StickerSheet: &StickerSheetSpec{Abilities: []StickerAbilitySpec{
			{Grant: goblinCowardParadeMentor},
			{Grant: goblinCowardParadeDestroy},
		}},
	})
}
