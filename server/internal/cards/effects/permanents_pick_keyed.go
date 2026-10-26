package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// permanents_pick_keyed.go — ChoosePermanents with its set rule, its
// sacrifice and the rest of the card REGISTERED rather than held in
// closures (#4290), so a table waiting on any seat's pick is a restore
// point.
//
// A card registers its pick once, from a package-level var, under an
// on-disk key (ledgered "run <key>": never renamed, never reused), and
// passes it as ChoosePermanents.Keyed. What the closure captured is data on
// the run: the resolving item itself, which the body's Context is rebuilt
// from, Carry (IDs) and Nums.

// PermanentsDef is the registered part of a ChoosePermanents.
type PermanentsDef struct {
	// Validate is the rule over the SET one seat picks (Tragic Arrogance's
	// "an artifact, a creature, an enchantment, and a planeswalker"),
	// applied to every leg. It is a function of the cards alone, because
	// it runs under the enumerator's read lock; it captures nothing.
	Validate func(picked []game.Card) bool

	// Sacrifice sacrifices everything chosen, across every leg, as one
	// simultaneous exit before Then runs (ChoosePermanents.Sacrifice).
	Sacrifice bool

	// Then is everything printed after the choice, run once with what
	// every seat picked. Its Context is rebuilt from the resolving item.
	// ThenRun is Then for a body that also reads the run's Carry or Nums;
	// set at most one.
	Then    func(ctx *Context, picked game.PromptedPicks) error
	ThenRun func(ctx *Context, picked game.PromptedPicks, r game.RunDone) error
}

// PermanentsThen names a registered permanent pick. Hold one only from
// RegisterPermanentsPick.
type PermanentsThen struct{ run game.PromptRunThen }

// RegisterPermanentsPick registers a permanent pick under key. Call it
// once, from a package-level var, or from init when the body asks the
// same pick again.
func RegisterPermanentsPick(key string, def PermanentsDef) PermanentsThen {
	body := func(ctx *Context, picked game.PromptedPicks, r game.RunDone) error {
		switch {
		case def.ThenRun != nil:
			return def.ThenRun(ctx, picked, r)
		case def.Then != nil:
			return def.Then(ctx, picked)
		}
		return nil
	}
	return PermanentsThen{run: RunThenRule(key, def.Validate, func(ctx *Context, r game.RunDone) error {
		picked := r.Picks()
		if !def.Sacrifice || picked.Count() == 0 {
			return body(ctx, picked, r)
		}
		item := ctx.Item
		return ctx.Game.SacrificeAllThenForEffect(uuid.Nil, picked.Cards(), func(g *game.Game, _ []uuid.UUID) error {
			return body(NewContext(g, item), picked, r)
		})
	})}
}
