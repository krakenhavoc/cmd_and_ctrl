package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

const (
	familiarBeebleUntap = "familiar-beeble-mascot/attacks-untap"
	familiarBeebleAnthem = "familiar-beeble-mascot/creature-enters-pump"
)

// Familiar Beeble Mascot — Stickers (an Unfinity sticker sheet, CR 123.2; ADR 0208, #4307):
//
//	Name stickers "Familiar", "Beeble", "Mascot"; three art stickers.
//	{TK}{TK} — Whenever this creature attacks, untap target permanent.
//	{TK}{TK}{TK}{TK} — Whenever a creature enters under your control, creatures you control get +1/+1 until end of turn.
//	{TK}{TK} — 2/3
//	{TK}{TK}{TK} — 5/3
//
// The first sticker gives an attack trigger that untaps a target permanent. The second gives a trigger on each creature entering under your control (the creature itself included) that pumps the creatures you control; the set is fixed as it resolves (CR 611.2c).
//
// The name, power and toughness, and art stickers are the rules' (CR 123.6, 123.8, 123.9).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "41c0e1f1-8ebe-4cd2-96fe-e4bb625fe6ee",
		Name:         "Familiar Beeble Mascot",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{
			{
				Key: familiarBeebleUntap,
				Triggered: []game.TriggeredAbility{
					Targeting(
						WheneverThisAttacks("Familiar Beeble Mascot sticker — untap target permanent", UntapEachLegalTarget),
						TargetPermanent("target permanent")),
				},
				Text: "Whenever this creature attacks, untap target permanent.",
			},
			{
				Key: familiarBeebleAnthem,
				Triggered: []game.TriggeredAbility{
					On(game.EventETB, CreatureEnteredUnderYourControl,
						"Familiar Beeble Mascot sticker — creatures you control get +1/+1 until end of turn",
						func(g *game.Game, item *game.StackItem) error {
							return BoostUntilEOT{Match: And(Creature(), YouControl()), Power: 1, Toughness: 1, Label: "Familiar Beeble Mascot sticker — +1/+1"}.Apply(NewContext(g, item))
						}),
				},
				Text: "Whenever a creature enters under your control, creatures you control get +1/+1 until end of turn.",
			},
		},
		StickerSheet: &StickerSheetSpec{Abilities: []StickerAbilitySpec{
			{Grant: familiarBeebleUntap},
			{Grant: familiarBeebleAnthem},
		}},
	})
}
