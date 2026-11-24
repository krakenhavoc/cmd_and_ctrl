package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Against All Odds — Sorcery {3}{W} (Phyrexia: All Will Be One):
//
//	"Choose one or both —
//	 • Exile target artifact or creature you control, then return it to the
//	   battlefield under its owner's control.
//	 • Return target artifact or creature card with mana value 3 or less from
//	   your graveyard to the battlefield."
//
// The bullets resolve in printed order (CR 608.2c) and each reads its own
// target, skipping it if it is no longer legal (CR 608.2b). The blink returns
// a new object under its OWNER's control (Flicker with no Controller), so a
// stolen creature you were controlling goes home. A token exiled by the first
// bullet ceases to exist and does not come back (CR 111.8). The second bullet
// puts a card from your own graveyard onto the battlefield under your
// control; the first bullet cannot return a card to the graveyard, so a
// creature you blink is not a target of the second.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f4a73cde-58bf-4719-8eb0-e039fdde4f62",
		Name:         "Against All Odds",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseN("Choose one or both", 1, 2,
			effects.ModeDoing("Exile target artifact or creature you control, then return it to the battlefield under its owner's control.",
				effects.TargetPermanent("target artifact or creature you control", effects.Or(effects.Artifact(), effects.Creature()), effects.YouControl()),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok || t.Kind != game.TargetCard {
						return nil
					}
					return effects.Flicker{Target: t.ID}.Apply(ctx)
				}),
			effects.ModeDoing("Return target artifact or creature card with mana value 3 or less from your graveyard to the battlefield.",
				effects.TargetCardInGraveyard("target artifact or creature card with mana value 3 or less from your graveyard",
					effects.YouOwn(), effects.Or(effects.Artifact(), effects.Creature()), effects.ManaValueLE(3)),
				effects.ModeReturnTargetToTheBattlefield),
		),
	})
}
