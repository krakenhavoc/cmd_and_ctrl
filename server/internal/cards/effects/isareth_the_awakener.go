package effects

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Isareth the Awakener — Legendary Creature — Human Wizard {1}{B}{B}, 3/3
// (Foundations Jumpstart):
//
//	"Deathtouch
//	 Whenever Isareth attacks, you may pay {X}. When you do, return target
//	 creature card with mana value X from your graveyard to the battlefield
//	 with a corpse counter on it. If that creature would leave the
//	 battlefield, exile it instead of putting it anywhere else."
//
// Halo Forager's pay-X-then-reflexive shape (MayPayX, CR 603.12): X is
// chosen as the attack trigger resolves and the target, a creature card in
// YOUR graveyard with exactly that mana value, as the reflexive trigger goes
// on the stack. The entry is From the Catacombs': the corpse counter rides
// the entry (CR 614.1c) and the exile redirect is registered from the
// entry's continuation, for the creature that arrived.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7af0e2da-9163-42d9-bf69-439cc61cd28d",
		Name:            "Isareth the Awakener",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Isareth the Awakener — pay {X}; return a creature card with mana value X from your graveyard", s82BnsB51IsarethAttacks),
		},
	})
}

var isarethPaid = PayUnlessBranches("pay-unless/isareth-x-reanimate",
	func(ctx *Context, r game.PayUnlessAnswer) error {
		return ReflexiveTrigger{
			Label:  "Isareth the Awakener — return a creature card with mana value " + strconv.Itoa(r.Number) + " from your graveyard",
			Body:   s82BnsB51IsarethReturnBody,
			Params: game.EffectParams{Amount: r.Number},
		}.Apply(ctx)
	}, nil)

func s82BnsB51IsarethAttacks(g *game.Game, item *game.StackItem) error {
	return MayPayX{
		Cost:  "{X}",
		Label: "Isareth the Awakener",
		Buys:  "return a creature card with mana value X from your graveyard to the battlefield with a corpse counter on it",
		Goal:  XAsHighAsYouCan,
		Then:  isarethPaid,
	}.Apply(NewContext(g, item))
}
