package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tyrant's Scorn — Instant {U}{B} (Breaking News, #5929):
//
//	"Choose one —
//	 • Destroy target creature with mana value 3 or less.
//	 • Return target creature to its owner's hand."
//
// Two bullets, each with its own target group, checked again as the spell
// resolves (CR 608.2b): a creature that has since grown past mana value 3
// is no longer a legal target of the first bullet. The destroy is the
// shared one (a regeneration shield or indestructible still saves it), the
// bounce goes to the owner's hand, a token ceasing to exist.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "afb278f1-0e21-4f6d-96bb-c34aa3ef96b0",
		Name:         "Tyrant's Scorn",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeWithPurpose(ModeDoing("Destroy target creature with mana value 3 or less.",
				TargetCreature("target creature with mana value 3 or less", ManaValueLE(3)),
				TitanOfIndustryDestroy), game.Purpose{Answers: game.AnswerRemove}),
			ModeDoing("Return target creature to its owner's hand.",
				TargetCreature("target creature"), TyrantsScornBounce),
		),
	})
}

func TyrantsScornBounce(_ *game.StackItem, ctx *Context, occ int) error {
	t, ok := ModeTarget(ctx, occ)
	if !ok || t.Kind != game.TargetCard {
		return nil
	}
	return BounceToHand{Target: t.ID}.Apply(ctx)
}
