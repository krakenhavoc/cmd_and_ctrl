package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Krenko, Tin Street Kingpin — 1/2 Legendary Creature — Goblin for
// {2}{R}:
//
//	"Whenever Krenko attacks, put a +1/+1 counter on it, then create
//	a number of 1/1 red Goblin creature tokens equal to Krenko's
//	power."
//
// S22 attack triggers: the "whenever THIS creature attacks" shape,
// and the one where the printed word "then" is load-bearing. The
// counter goes on first and the token count is read AFTER it, so a
// freshly-cast 1/2 Krenko makes two Goblins on its first swing, not
// one. Reading power before the counter is the classic mis-read.
//
// Power is read live off the battlefield at resolution rather than
// captured in Build, because everything between declaration and
// resolution counts — a pump spell, an anthem, a second Krenko
// trigger from an earlier combat phase.
//
// A Krenko that is not the attacking object any more — removed in
// response, or flickered and back as a NEW object (CR 400.7) — takes
// no counter, and the Goblins are made from its last-known power
// (CR 608.2h), read through Context.SourcePermanent (#1418, #1432).
// That is the paper answer: the counter has nothing to go on, and
// "Krenko's power" is the power it had as it last existed.
func init() {
	Register(Spec{
		OracleID:     "e8065e1d-e937-4b56-8011-78f0d07328a0",
		Name:         "Krenko, Tin Street Kingpin",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Krenko, Tin Street Kingpin — +1/+1 counter, then Goblins", func(g *game.Game, item *game.StackItem) error {
				// #1290: the power that matters is Krenko's power AFTER
				// the +1/+1 counter LANDS, not on the next line — a
				// Doubling Season / Hardened Scales board pauses the
				// placement on a CR 616 prompt, and reading power
				// before that resumes would make Goblins equal to the
				// pre-placement power.
				ctx := NewContext(g, item)
				krenko, ok := ctx.SourcePermanent()
				if !ok {
					return nil
				}
				if krenko.Left {
					return CreateToken{
						Controller: item.Controller,
						Template:   RedGoblinToken(),
						N:          max(krenko.Power, 0),
					}.Apply(ctx)
				}
				return g.AddCounterThenForEffect(item.SourceCardID, game.CounterPlusOne, 1, func(g *game.Game, _ int) error {
					krenko, ok := g.LookupCardForEffect(item.SourceCardID)
					if !ok {
						return nil
					}
					return CreateToken{
						Controller: item.Controller,
						Template:   RedGoblinToken(),
						N:          krenko.CurrentPower(),
					}.Apply(NewContext(g, item))
				})
			}),
		},
	})
}
