package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

const notoriousSliverWarPump = "notorious-sliver-war/five-anthem"

// Notorious Sliver War — Stickers (an Unfinity sticker sheet, CR 123.2; ADR 0208, #4307):
//
//	Name stickers "Notorious", "Sliver", "War"; three art stickers.
//	{TK}{TK} — {5}: Creatures you control get +1/+1 until end of turn.
//	{TK}{TK}{TK} — Protection from creatures with two or more creature types
//	{TK}{TK} — 3/3
//	{TK}{TK}{TK}{TK}{TK} — 9/6
//
// The first sticker gives an ability that pumps the creatures its controller controls; the set is fixed as it resolves (CR 611.2c). The second gives protection from creatures with two or more creature types (CR 702.16): a creature source with at least two creature types, a changeling being all of them (CR 702.73a).
//
// The name, power and toughness, and art stickers are the rules' (CR 123.6, 123.8, 123.9).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "06f8efdd-4e8e-4524-859b-71f2f1732ad5",
		Name:         "Notorious Sliver War",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{
			{
				Key: notoriousSliverWarPump,
				Activated: []ActivatedAbility{{
					Label:   "{5}: Creatures you control get +1/+1 until end of turn.",
					Cost:    ManaCost("{5}"),
					Purpose: game.Purpose{Answers: game.AnswerPump},
					Effect: func(g *game.Game, item *game.StackItem) error {
						return BoostUntilEOT{Match: And(Creature(), YouControl()), Power: 1, Toughness: 1, Label: "Notorious Sliver War sticker — +1/+1"}.Apply(NewContext(g, item))
					},
				}},
				Text: "{5}: Creatures you control get +1/+1 until end of turn.",
			},
		},
		StickerSheet: &StickerSheetSpec{Abilities: []StickerAbilitySpec{
			{Grant: notoriousSliverWarPump},
			{Keywords: []string{"protection from creatures with two or more creature types"}},
		}},
	})
}
