package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

const (
	ungluedPeaBrainedDinosaurMana    = "unglued-pea-brained-dinosaur/tap-mana"
	ungluedPeaBrainedDinosaurAnimate = "unglued-pea-brained-dinosaur/combat-animate"
)

// Unglued Pea-Brained Dinosaur — Stickers (an Unfinity sticker sheet, CR 123.2; ADR 0208, #4307):
//
//	Name stickers "Unglued", "Pea-Brained", "Dinosaur"; three art stickers.
//	{TK}{TK} — {T}: Add {2}. Spend this mana only to cast creature spells.
//	{TK}{TK}{TK}{TK} — At the beginning of combat on your turn, target noncreature artifact you control becomes a 4/4 artifact creature with flying until end of turn.
//	{TK}{TK} — 2/4
//	{TK}{TK}{TK}{TK}{TK} — 8/6
//
// The first sticker gives a mana ability with a spending restriction (CR 106.12): the two colorless mana can pay only for creature spells. The second gives a beginning-of-combat trigger on its controller's turn that animates a target noncreature artifact they control: it becomes a 4/4 artifact creature with flying until end of turn (layers 4, 6 and 7b).
//
// The name, power and toughness, and art stickers are the rules' (CR 123.6, 123.8, 123.9).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "eb1c4f0a-b993-4aa1-8b13-1695b5f63653",
		Name:         "Unglued Pea-Brained Dinosaur",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{
			{
				Key: ungluedPeaBrainedDinosaurMana,
				Mana: []ManaAbility{{
					Cost:         ManaAbilityCost{Tap: true},
					Produced:     "{C}{C}",
					Label:        "{T}: Add {2}. Spend this mana only to cast creature spells.",
					Restrictions: []string{ManaRestrictCast, ManaRestrictType("Creature")},
				}},
				Text: "{T}: Add {2}. Spend this mana only to cast creature spells.",
			},
			{
				Key: ungluedPeaBrainedDinosaurAnimate,
				Triggered: []game.TriggeredAbility{
					Targeting(
						AtBeginningOfYourCombat("Unglued Pea-Brained Dinosaur sticker — target noncreature artifact you control becomes a 4/4 artifact creature with flying until end of turn",
							func(g *game.Game, item *game.StackItem) error {
								ctx := NewContext(g, item)
								id, ok := stickerFirstLegalTarget(ctx)
								if !ok {
									return nil
								}
								if err := (BecomeCreatureUntilEOT{
									Target: id, Types: []string{"Artifact", "Creature"}, SetPower: 4, SetToughness: 4,
									Label: "Unglued Pea-Brained Dinosaur sticker — 4/4 artifact creature",
								}).Apply(ctx); err != nil {
									return err
								}
								return GrantKeywordUntilEOT{Target: id, Keywords: []string{"flying"}, Label: "Unglued Pea-Brained Dinosaur sticker — flying"}.Apply(ctx)
							}),
						TargetPermanent("target noncreature artifact you control", Artifact(), Not(Creature()), YouControl())),
				},
				Text: "At the beginning of combat on your turn, target noncreature artifact you control becomes a 4/4 artifact creature with flying until end of turn.",
			},
		},
		StickerSheet: &StickerSheetSpec{Abilities: []StickerAbilitySpec{
			{Grant: ungluedPeaBrainedDinosaurMana},
			{Grant: ungluedPeaBrainedDinosaurAnimate},
		}},
	})
}
