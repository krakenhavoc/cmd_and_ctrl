package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "f4a73cde-58bf-4719-8eb0-e039fdde4f62",
		Name:         "Against All Odds",
		Completeness: CompletenessFull,
		Modes: ChooseN("Choose one or both", 1, 2,
			ModeDoing("Exile target artifact or creature you control, then return it to the battlefield under its owner's control.",
				TargetPermanent("target artifact or creature you control", Or(Artifact(), Creature()), YouControl()),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok || t.Kind != game.TargetCard {
						return nil
					}
					return Flicker{Target: t.ID}.Apply(ctx)
				}),
			ModeDoing("Return target artifact or creature card with mana value 3 or less from your graveyard to the battlefield.",
				TargetCardInGraveyard("target artifact or creature card with mana value 3 or less from your graveyard",
					YouOwn(), Or(Artifact(), Creature()), ManaValueLE(3)),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok || t.Kind != game.TargetCard {
						return nil
					}
					return ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneBattlefield}.Apply(ctx)
				}),
		),
	})
}
