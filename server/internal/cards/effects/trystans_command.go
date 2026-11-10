package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Trystan's Command — Kindred Sorcery — Elf {4}{B}{G} (Lorwyn Eclipsed,
// #4458):
//
//	"Choose two —
//	 • Create a token that's a copy of target Elf you control.
//	 • Return one or two target permanent cards from your graveyard to
//	   your hand.
//	 • Destroy target creature or enchantment.
//	 • Creatures target player controls get +3/+3 until end of turn.
//	   Untap them."
//
// Ashling's Command's "choose two" template (ChooseN 2..2): each chosen
// bullet has its own targets, checked again as the spell resolves (CR
// 608.2b), and the bullets run in printed order (CR 608.2c). The copy
// bullet is KindredCommandCopyMode. The return is Archenemy's Charm's
// one clause with a count of one to two, over permanent cards you own in
// your graveyard; each pick still there goes to your hand. The last
// bullet fixes its set as it resolves (CR 611.2c): the target player's
// creatures get +3/+3, and those of them that are tapped untap.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ecc4e230-a546-4cb4-987c-87300060bbe2",
		Name:         "Trystan's Command",
		Completeness: CompletenessFull,
		Modes: ChooseN("Choose two", 2, 2,
			KindredCommandCopyMode("Elf"),
			ModeWithPurpose(ModeDoing("Return one or two target permanent cards from your graveyard to your hand.",
				TargetCardInGraveyard("one or two target permanent cards from your graveyard",
					YouOwn(), Permanent()).WithCount(1, 2),
				eoeABCharmReturn), game.Purpose{Answers: game.AnswerValue}),
			ModeWithPurpose(ModeDoing("Destroy target creature or enchantment.",
				TargetPermanent("target creature or enchantment", Or(Creature(), Enchantment())),
				DestroyTheModesTarget), game.Purpose{Answers: game.AnswerRemove}),
			ModeWithPurpose(ModeDoing("Creatures target player controls get +3/+3 until end of turn. Untap them.",
				TargetPlayer("target player"), trystansCommandPumpAndUntap),
				game.Purpose{Answers: game.AnswerPump}),
		),
	})
}

// trystansCommandPumpAndUntap is "creatures target player controls get
// +3/+3 until end of turn. Untap them.": the same creatures, read once.
func trystansCommandPumpAndUntap(_ *game.StackItem, ctx *Context, occ int) error {
	t, ok := ModeTarget(ctx, occ)
	if !ok || t.Kind != game.TargetPlayer {
		return nil
	}
	if err := (BoostUntilEOT{Match: And(Creature(), controlledBy(t.ID)), Power: 3, Toughness: 3, Label: "Trystan's Command — +3/+3"}).Apply(ctx); err != nil {
		return err
	}
	return b16UntapAllYouControlMatching(ctx, t.ID, func(c game.Card) bool { return c.IsCreature() })
}
