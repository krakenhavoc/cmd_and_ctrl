package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Earthshape — Instant {2}{W}:
//
//	"Earthbend 3. Then each creature you control with power less than
//	 or equal to that land's power gains hexproof and indestructible
//	 until end of turn. You gain hexproof until end of turn."
//
// The #1282 proof card for earthbend. ADR 0081 deferred it because it
// "reads the animated land's power afterwards", and afterwards is the
// whole problem: the counters are the one part of earthbend that can
// PAUSE (a CR 616 ordering prompt between two counter replacements —
// a Doubling Season beside a Hardened Scales), and read on the next
// line "that land's power" is the 0/0 body with no counters on it. The
// sweep is `Earthbend.Then`, which rides the counter placement's
// continuation and runs when the counters are really there.
//
// "That land's power" is read ONCE, when the sentence runs, and the
// affected set is fixed then too (CR 611.2c): a creature that grows
// past the land later in the turn keeps its hexproof. The land itself
// is a creature you control with power equal to its own, so it is
// shielded as well — which is the printed result, not an accident.
//
// "You gain hexproof" is GainPlayerKeyword, the player-level grant
// Dawn's Truce uses, for the printed duration.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "300f17a2-1594-40cb-9187-952898a83622",
		Name:         "Earthshape",
		Completeness: CompletenessFull,
		Targets:      EarthbendTargets(),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			land := FirstLegalBattlefieldTarget(ctx)
			if land == uuid.Nil {
				return nil
			}
			return Earthbend{Target: land, N: 3, Then: func(ctx *Context) error {
				return earthshapeShield(ctx, land)
			}}.Apply(ctx)
		},
	})
}

// earthshapeShield is everything after "Then".
func earthshapeShield(ctx *Context, land uuid.UUID) error {
	g := ctx.Game
	// The counters just landed; CurrentPower adds them on top of the
	// layered 0/0, and the recompute makes the layered half current.
	g.RecomputeLayersIfStaleLocked()
	if c, ok := g.LookupCardForEffect(land); ok && onBattlefield(g, land) {
		limit := c.CurrentPower()
		if err := (GrantKeywordUntilEOT{
			Match: And(Creature(), YouControl(), func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
				return c.CurrentPower() <= limit
			}),
			Keywords: []string{"hexproof", "indestructible"},
			Label:    "Earthshape — hexproof and indestructible",
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return GainPlayerKeyword{
		Player:   ctx.Controller(),
		Keyword:  KeywordHexproof,
		Label:    "Earthshape — you gain hexproof",
		Duration: DurationUntilEndOfTurn(ctx),
	}.Apply(ctx)
}
