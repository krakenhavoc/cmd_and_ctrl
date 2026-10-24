package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tato Farmer — Creature — Zombie Mutant Peasant {2}{G}, 1/4:
//
//	"Landfall — Whenever a land you control enters, you may get two rad
//	 counters.
//	 {T}: Put target land card in a graveyard that was milled this turn
//	 onto the battlefield tapped under your control."
//
// #2809. The landfall "you may" is asked as it resolves (CR 603.5). The
// tap ability targets a land card in ANY graveyard whose most recent
// arrival there this turn was a mill (MilledThisTurn): a land surveilled,
// discarded or put there by a search is not one, and neither is a land
// that was already there when the turn began. The card enters tapped
// under the Farmer's controller's control, whoever owns it, and its own
// landfall sees it. The {T} is a creature's, so the Farmer must have
// been under its controller's control since their turn began (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "71b5a3b4-31bb-44d8-a80c-36efff2a621e",
		Name:         "Tato Farmer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(Landfall("Tato Farmer — you get two rad counters", Do(youGetRadCounters{N: 2})),
				"Tato Farmer — get two rad counters?"),
		},
		Activated: []ActivatedAbility{{
			Label:   "{T}: Put target land card in a graveyard that was milled this turn onto the battlefield tapped under your control.",
			Cost:    TapCost(),
			Targets: TargetCardInGraveyard("target land card in a graveyard that was milled this turn", Land(), MilledThisTurn()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return ReturnFromGraveyard{Target: id, Dest: game.ZoneBattlefield, Controller: ctx.Controller(), Tapped: true}.Apply(ctx)
			},
		}},
	})
}
