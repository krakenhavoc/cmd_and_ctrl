package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Reanimate — Sorcery {B}:
//
//	"Put target creature card from a graveyard onto the battlefield
//	 under your control. You lose life equal to that card's mana
//	 value."
//
// The card the whole archetype is named after, and the one that
// exercises the two clauses that make reanimation more than a
// zone move.
//
// "From A GRAVEYARD" — every pile at the table, not just yours.
// Discarding is one way to fill a graveyard; an opponent's board
// wipe is another, and Reanimate is how you charge them for it.
//
// "UNDER YOUR CONTROL" — the creature changes hands. That clause is
// why ReturnFromGraveyardUnderControlForEffect exists: the primitive
// used to route the card to the graveyard's owner, so reanimating
// across the table handed the creature straight back.
//
// The life loss is the card's mana value, read off the card in the
// graveyard before it moves — a creature that would be bigger or
// smaller as a permanent (a CDA, an anthem) does not change what you
// pay, and {X} in a graveyard is 0 (CR 202.3e).
//
// "You lose life" is not damage: no lifelink, no prevention, no
// damage triggers. ChangePlayerLifeForEffect with a negative delta
// is the right primitive and DealDamage is not.
func init() {
	Register(Spec{
		OracleID: "a044474a-cd72-4e9d-bd8d-a08f2de9cdc0",
		Name:     "Reanimate",
		Targets:  targetCreatureInAnyGraveyard(),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			card, ok := reanimateSingleTarget(ctx, ctx.Controller())
			if !ok {
				return nil
			}
			mv := card.ManaValue()
			if mv <= 0 {
				return nil
			}
			return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), ctx.Controller(), -mv)
		},
	})
}
