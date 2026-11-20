package cardso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	effects.Register(effects.Spec{
		OracleID:     "884eaa3f-b6af-4f0e-b0e7-cd40666986ce",
		Name:         "Obscura Charm",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Return target multicolored permanent card with mana value 3 or less from your graveyard to the battlefield tapped.",
				effects.TargetCardInGraveyard("target multicolored permanent card with mana value 3 or less from your graveyard",
					effects.YouOwn(), effects.Permanent(), effects.Multicolored(), effects.ManaValueLE(3)),
				effects.RfCreatureBModeTargetDoes(func(ctx *effects.Context, id uuid.UUID) error {
					return effects.ReturnFromGraveyard{Target: id, Dest: game.ZoneBattlefield, Tapped: true}.Apply(ctx)
				})),
			effects.ModeDoing("Counter target instant or sorcery spell.",
				effects.TargetSpell("target instant or sorcery spell", effects.Or(effects.Instant(), effects.Sorcery())),
				effects.CounterTheModesTarget),
			effects.ModeDoing("Destroy target creature or planeswalker with mana value 3 or less.",
				effects.Harmful(effects.TargetPermanent("target creature or planeswalker with mana value 3 or less",
					effects.Or(effects.Creature(), effects.Planeswalker()), effects.ManaValueLE(3))),
				effects.DestroyTheModesTarget),
		),
	})
}
