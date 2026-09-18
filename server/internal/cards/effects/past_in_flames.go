package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Past in Flames — Sorcery {3}{R}:
//
//	"Each instant and sorcery card in your graveyard gains flashback
//	 until end of turn. The flashback cost is equal to its mana cost.
//	 Flashback {4}{R}"
//
// Snapcaster Mage's clause pointed at a whole graveyard, and the card
// that makes CR 611.2c matter: a one-shot continuous effect locks the
// set of objects it affects when it RESOLVES. So the resolution walks
// the graveyard once and writes down the instances; a card discarded,
// milled or killed a moment later is not in the set and has no
// flashback, however much the turn has left. That is not an
// optimisation — it is the rule, and a re-derived "every instant in
// your graveyard right now" would be strictly stronger than printed.
//
// The card also PRINTS flashback, which is the interaction worth
// pinning: a granted permission never overrides an offer the card
// itself makes, so Past in Flames flashed back under a second Past in
// Flames still costs its printed {4}{R} rather than its {3}{R} mana
// cost. The catalog is consulted first (ADR 0066 decision 3), and the
// direction that survives is the one that errs against the player.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "37a18736-5fe2-4897-809b-013497bdd890",
		Name:             "Past in Flames",
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{4}{R}")},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return GrantCastFromYourGraveyard{
				Filter:            game.PermissionFilter{InstantOrSorceryOnly: true},
				AltCostKey:        "flashback",
				ExileOnResolution: true,
				Label:             "Flashback — its mana cost (Past in Flames)",
			}.Apply(ctx)
		},
	})
}
