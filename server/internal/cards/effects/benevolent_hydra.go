package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Benevolent Hydra — Creature — Hydra {X}{G}{G}, 1/1 (EDHREC rank
// 2955):
//
//	"This creature enters with X +1/+1 counters on it.
//	 If one or more +1/+1 counters would be put on another creature
//	 you control, that many plus one +1/+1 counters are put on it
//	 instead.
//	 {T}, Remove a +1/+1 counter from this creature: Put a +1/+1
//	 counter on another target creature you control."
//
// Hardened Scales on an X-sized body. The X counters go on as the
// spell resolves (Goldvein Hydra's posture, declared below); the
// replacement is Conclave Mentor's with the Hydra itself excluded
// (b28OtherCreaturesYouControlGetAnExtraCounter — "another"), so a
// second Hydra's X counters get the bonus from the first and a
// Hardened Scales stacks in CR 616 order.
//
// The tap ability moves a counter: its cost removes a +1/+1 counter
// from the Hydra at announce (RemoveCountersFromThis, #625), and the
// counter it puts on the target is an ordinary placement, so the
// Hydra's own replacement makes it two. The Hydra is a printed 1/1,
// so spending its last counter leaves a 1/1, not a 0/0.
//
// Two declared simplifications, both weaker than printed:
//
//   - The X +1/+1 counters are placed as the spell resolves, a beat
//     before the card enters (an entry replacement cannot read the
//     spell's X), so a "whenever you put counters on a permanent"
//     payoff does not see them — Goldvein Hydra's gap.
//   - "Another target creature" is matched by NAME, the catalog's
//     convention for "another" on an activated ability's target
//     clause (Dour Port-Mage, Ioreth): a target predicate is not told
//     which permanent is activating, so it cannot exclude just this
//     one. The Hydra cannot target itself, as printed, and cannot
//     target a second Benevolent Hydra either, which the printed card
//     can.
func init() {
	Register(Spec{
		OracleID:     "01dbf1bc-ca62-4fb6-959c-ef7c0dc03bb0",
		Name:         "Benevolent Hydra",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The X +1/+1 counters are put on the Hydra as the spell resolves, a beat before it enters, so effects that watch you put counters on a permanent don't see them.",
			"The tap ability can't move a counter onto another Benevolent Hydra.",
		},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return AddCounter{Target: item.SourceCardID, Kind: "+1/+1", N: ctx.X()}.Apply(ctx)
		},
		Replacements: []game.ReplacementEffect{
			b28OtherCreaturesYouControlGetAnExtraCounter("Benevolent Hydra: +1 +1/+1 counter"),
		},
		Activated: []ActivatedAbility{{
			Label:   "{T}, Remove a +1/+1 counter from this creature: Put a +1/+1 counter on another target creature you control.",
			Cost:    Plus(TapCost(), RemoveCountersFromThis(game.CounterPlusOne, 1)),
			Targets: TargetCreature("another target creature you control", YouControl(), b03NotNamed("Benevolent Hydra")),
			Effect:  b36CounterOnChosenAnimal,
		}},
	})
}
