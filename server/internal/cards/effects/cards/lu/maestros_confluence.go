package cardslu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Maestros Confluence — Sorcery {3}{U}{B}{R} (Streets of New Capenna
// Commander):
//
//	"Choose three. You may choose the same mode more than once.
//	 • Return target monocolored instant or sorcery card from your
//	   graveyard to your hand.
//	 • Target creature gets -3/-3 until end of turn.
//	 • Goad each creature target player controls."
//
// Mystic Confluence's shape (CR 700.2d). The monocolored test reads the
// card's colours in the graveyard (CR 105.2); a colourless card is not
// monocolored. The goad (CR 701.15) takes the creatures the target player
// controls as it resolves; it lasts until your next turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ef67a41b-a2a6-4f1b-a0d5-15560e99e387",
		Name:         "Maestros Confluence",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseNRepeating("Choose three (you may choose the same mode more than once)", 3, 3,
			effects.ModeDoing("Return target monocolored instant or sorcery card from your graveyard to your hand.",
				effects.TargetCardInGraveyard("target monocolored instant or sorcery card from your graveyard",
					effects.YouOwn(), effects.Or(effects.Instant(), effects.Sorcery()), effects.B41Monocolored()),
				effects.RfCreatureBModeTargetDoes(func(ctx *effects.Context, id uuid.UUID) error {
					return effects.ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}.Apply(ctx)
				})),
			effects.ModeWithPurpose(effects.ModeDoing("Target creature gets -3/-3 until end of turn.",
				effects.Harmful(effects.TargetCreature("target creature")),
				effects.BoostTheModesTarget(-3, -3, "Maestros Confluence — -3/-3")),
				effects.ForTargets(effects.HarmUntilEOT(0, -3, -3))),
			effects.ModeDoing("Goad each creature target player controls.",
				effects.TargetPlayer("target player"),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok || t.Kind != game.TargetPlayer {
						return nil
					}
					return effects.GoadAllMatching(ctx, func(_ *game.Game, c game.Card) bool { return c.Controller == t.ID })
				}),
		),
	})
}
