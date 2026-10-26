package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

const (
	squishySphinxNinjaWard    = "squishy-sphinx-ninja/ward-2"
	squishySphinxNinjaProvoke = "squishy-sphinx-ninja/provoke"
)

// Squishy Sphinx Ninja — Stickers (an Unfinity sticker sheet, CR 123.2; ADR 0208, #4307):
//
//	Name stickers "Squishy", "Sphinx", "Ninja"; three art stickers.
//	{TK}{TK} — Ward {2}
//	{TK}{TK}{TK}{TK} — Provoke
//	{TK}{TK}{TK} — 5/3
//	{TK}{TK}{TK}{TK}{TK} — 7/7
//
// Ward {2} (CR 702.21a) is the engine's triggered ability (Ward): whenever the creature becomes the target of a spell or ability an opponent controls, counter it unless that player pays {2}. Provoke (CR 702.39a) is the engine's attack trigger (Provoke).
//
// The name, power and toughness, and art stickers are the rules' (CR 123.6, 123.8, 123.9).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d198b39d-2c5f-4150-8c54-a00d33d74a28",
		Name:         "Squishy Sphinx Ninja",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{
			{
				Key:       squishySphinxNinjaWard,
				Triggered: []game.TriggeredAbility{Ward(WardMana("{2}"), "Squishy Sphinx Ninja sticker — ward {2}")},
				Text:      "Ward {2} (Whenever this creature becomes the target of a spell or ability an opponent controls, counter it unless that player pays {2}.)",
			},
			{
				Key:       squishySphinxNinjaProvoke,
				Triggered: []game.TriggeredAbility{Provoke()},
				Text:      "Provoke (Whenever this creature attacks, you may have target creature defending player controls untap and block it if able.)",
			},
		},
		StickerSheet: &StickerSheetSpec{Abilities: []StickerAbilitySpec{
			{Grant: squishySphinxNinjaWard},
			{Grant: squishySphinxNinjaProvoke},
		}},
	})
}
