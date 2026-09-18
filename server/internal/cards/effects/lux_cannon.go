package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lux Cannon — Artifact {4} (EDHREC rank 4118):
//
//	"{T}: Put a charge counter on this artifact.
//	 {T}, Remove three charge counters from this artifact: Destroy
//	 target permanent."
//
// Colourless, repeatable, unconditional removal that answers a land,
// an indestructible-granting enchantment or a commander in the
// command zone's way — anything at all, which is the reason a deck
// with no white and no black runs it. The price is that it is slow:
// three turns of tapping to charge, and the tap for the fourth turn
// is the shot. Every proliferate effect in the format is a turn off
// that clock, which is why Lux Cannon and Contagion Engine show up in
// the same deck lists.
//
// THE TWO ABILITIES BOTH TAP IT, so they cannot be used on the same
// turn, and the second one's tap is a REAL cost paid at announce —
// with an untapper (an Unwinding Clock, a Voltaic Key) the Cannon can
// charge and fire in one turn, which is the printed interaction and
// falls out of the costs being costs.
//
// "REMOVE THREE CHARGE COUNTERS" IS A COST, NOT AN EFFECT (#625): it
// is paid at announce alongside the tap, so an opponent responding to
// the activation sees a Cannon that has already spent its counters
// and cannot make the ability fizzle by shrinking it. A Cannon with
// two counters cannot announce the ability at all.
//
// "TARGET PERMANENT" is the widest target clause the engine has —
// lands, planeswalkers, the Cannon itself. Destruction, not exile, so
// indestructible and regeneration answer it as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ebdf26a4-77ee-4635-8d52-926bdca623f7",
		Name:         "Lux Cannon",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label: "{T}: Put a charge counter on this artifact.",
				Cost:  TapCost(),
				Effect: func(g *game.Game, item *game.StackItem) error {
					if !b09SourceStillOnBattlefield(g, item) {
						return nil
					}
					return AddCounter{
						Target: item.SourceCardID,
						Kind:   game.CounterCharge,
						N:      1,
					}.Apply(NewContext(g, item))
				},
			},
			{
				Label:   "{T}, Remove three charge counters from this artifact: Destroy target permanent.",
				Cost:    Plus(TapCost(), RemoveCountersFromThis(game.CounterCharge, 3)),
				Targets: TargetPermanent("target permanent"),
				// The shared single-target destroy body: a target that
				// became illegal in response is skipped, not errored.
				Effect: destroyFirstLegalTarget,
			},
		},
	})
}
