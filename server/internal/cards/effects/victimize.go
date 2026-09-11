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
// Sandbox simplification, declared in full because it is the one
// that changes how the card plays: the sacrifice is modelled as an
// ADDITIONAL COST TO CAST rather than as a resolution-time action.
// The engine has no "sacrifice a creature, and if you do, continue"
// prompt mid-resolution (PendingChoiceSacrifice has no
// continuation), and a cost is the only sacrifice-with-a-choice
// shape that exists. The observable differences all run the weaker
// way:
//
//   - The creature dies at announce, so its dies-triggers resolve
//     BEFORE Victimize does, and countering Victimize does not give
//     it back.
//   - With no creature to sacrifice the spell cannot be cast at all
//     (printed, it can be cast to do nothing).
//   - The sacrificed creature can never be one of the two targets.
//     That is true of the printed card too, and the engine keeps it
//     so: CastSpell validates the targets BEFORE it pays the
//     additional cost, so the creature is still on the battlefield
//     when the graveyard is scanned for legal targets.
//
// Second simplification: the returned creatures enter untapped and
// are tapped a beat later, because ReturnFromGraveyard has no tapped
// flag (its exile-side twin does). Anything watching for a tap
// event sees one; nothing watching for an untapped permanent
// entering gets a window to act, because both happen inside one
// resolution.
func init() {
	Register(Spec{
		OracleID: "240e85d3-e495-4877-8609-4b4056c402f7",
		Name:     "Victimize",
		Targets: TargetCardInGraveyard("two target creature cards in your graveyard",
			Creature(), YouOwn()).WithCount(2, 2),
		AdditionalCost: SacrificeCost("a creature", Creature()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (ReturnFromGraveyard{
					Target: t.ID,
					Dest:   game.ZoneBattlefield,
				}).Apply(ctx); err != nil {
					return err
				}
				if err := (TapTarget{Target: t.ID}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
