package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Otterball Antics — Instant {1}{U} (Bloomburrow, #4467):
//
//	"Create a 1/1 blue and red Otter creature token with prowess. If this
//	 spell was cast from anywhere other than your hand, put a +1/+1 counter
//	 on that creature.
//	 Flashback {3}{U}"
//
// The token enters with the counter (the clause is one instruction over the
// token it just made). "Cast from anywhere other than your hand" reads the
// item's cast zone: a flashback cast from the graveyard has it, a hand cast
// does not, and a copy was never cast so it does not either (CR 707.10).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "43a6de3e-152e-4af1-acec-53405742425c",
		Name:             "Otterball Antics",
		Completeness:     CompletenessFull,
		Purpose:          game.Purpose{Tokens: 1},
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{3}{U}")},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			spec := Token(TokenCard("1/1 blue and red Otter with prowess"))
			if item.CastFromZone != "" && item.CastFromZone != game.ZoneHand {
				spec = spec.WithCounters(game.CounterPlusOne, 1)
			}
			return CreateTokenAdvanced{Spec: spec, N: 1}.Apply(ctx)
		},
	})
}
