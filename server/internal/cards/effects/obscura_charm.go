package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Obscura Charm — Instant {W}{U}{B} (Streets of New Capenna):
//
//	"Choose one —
//	 • Return target multicolored permanent card with mana value 3 or less
//	   from your graveyard to the battlefield tapped.
//	 • Counter target instant or sorcery spell.
//	 • Destroy target creature or planeswalker with mana value 3 or less."
//
// The reanimation's target is chosen on the cast; a card that left the
// graveyard in response does nothing (CR 608.2b). The multicolored test
// reads the card's colours in the graveyard (CR 105.2), and "permanent
// card" excludes instants and sorceries.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "884eaa3f-b6af-4f0e-b0e7-cd40666986ce",
		Name:         "Obscura Charm",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeDoing("Return target multicolored permanent card with mana value 3 or less from your graveyard to the battlefield tapped.",
				TargetCardInGraveyard("target multicolored permanent card with mana value 3 or less from your graveyard",
					YouOwn(), Permanent(), Multicolored(), ManaValueLE(3)),
				rfCreatureBModeTargetDoes(func(ctx *Context, id uuid.UUID) error {
					return ReturnFromGraveyard{Target: id, Dest: game.ZoneBattlefield, Tapped: true}.Apply(ctx)
				})),
			ModeDoing("Counter target instant or sorcery spell.",
				TargetSpell("target instant or sorcery spell", Or(Instant(), Sorcery())),
				CounterTheModesTarget),
			ModeDoing("Destroy target creature or planeswalker with mana value 3 or less.",
				Harmful(TargetPermanent("target creature or planeswalker with mana value 3 or less",
					Or(Creature(), Planeswalker()), ManaValueLE(3))),
				DestroyTheModesTarget),
		),
	})
}
