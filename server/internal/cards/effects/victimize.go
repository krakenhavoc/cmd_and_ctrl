package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Victimize — Sorcery {2}{B} (EDHREC rank 128):
//
//	"Choose two target creature cards in your graveyard. Sacrifice a
//	 creature. If you do, return the chosen cards to the battlefield
//	 tapped."
//
// Two-for-one reanimation off a spare body. The two targets are one
// clause with a count of exactly two (CR 601.2c — the spell cannot
// be cast with one creature card in the graveyard), answered by two
// clicks in the zone browser, and each is re-checked at resolution
// so a target exiled in response is skipped while the other still
// returns (CR 608.2b — "does as much as it can").
//
// The sacrifice is part of the EFFECT, not a cost (#2863, ADR 0013
// amendment of 2026-10-09). It is a one-seat sacrifice run
// (PlayerSacrificesThenForEffect, ADR 0013 §5x): the caster picks a
// creature as the spell resolves, and "if you do" is the run's answer
// for the caster, read once the creature has really left the
// battlefield. So the spell can be cast with no creature at all (it
// then does nothing), the creature's death triggers go on the stack
// AFTER Victimize has resolved, and a countered Victimize costs no
// creature. The sacrificed creature can never be one of the two
// cards returned: the targets were chosen at cast, before it died.
//
// The two creatures enter together and tapped (#1867): one entry, so
// each sees the other enter (CR 603.6a), with the tapped clause on the
// entry event rather than a tap a beat later.
func init() {
	Register(Spec{
		OracleID:     "240e85d3-e495-4877-8609-4b4056c402f7",
		Name:         "Victimize",
		Completeness: CompletenessFull,
		Targets: TargetCardInGraveyard("two target creature cards in your graveyard",
			Creature(), YouOwn()).WithCount(2, 2),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			controller := ctx.Controller()
			return ctx.Game.PlayerSacrificesThenForEffect(
				ctx.Source(), controller,
				sacrificeSpec("a creature", Creature()),
				"Victimize — sacrifice a creature",
				1,
				func(g *game.Game, sacrificed game.PromptedSacrifices) error {
					if !sacrificed.Sacrificed(controller) {
						return nil
					}
					// A fresh Context on the live *Game (resumeClause's
					// contract), and the targets re-checked now.
					ctx := NewContext(g, item)
					return ReturnFromGraveyardTogether{Targets: legalTargetCardIDs(ctx), Tapped: true}.Apply(ctx)
				})
		},
	})
}
