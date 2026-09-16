package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hangarback Walker — Artifact Creature — Construct {X}{X}, 0/0
// (EDHREC rank 1528):
//
//	"This creature enters with X +1/+1 counters on it.
//	 When this creature dies, create a 1/1 colorless Thopter artifact
//	 creature token with flying for each +1/+1 counter on this
//	 creature.
//	 {1}, {T}: Put a +1/+1 counter on this creature."
//
// The artifact that turns into Thopters. Three abilities, all live:
// the X-sized body, the tap-to-grow, and the death payout — Thopters
// equal to the +1/+1 counters it had when it died.
//
// Sandbox simplification, declared, for the X counters — Goldvein
// Hydra's: an entry replacement cannot see the X announced for the
// spell (the stack item is gone by the time the entry pipeline
// runs), so the counters go on as the spell RESOLVES, a beat before
// the card moves from the stack to the battlefield, which is the last
// moment X is readable. They are on the card when it lands, so the
// 0/0 body never meets the state-based check without them, every ETB
// watcher sees the finished creature, and a counter doubler or
// Hardened Scales applies, exactly as it does to a printed "enters
// with". The one observable difference is that a "whenever you put
// counters on a permanent" payoff does not see them, because the card
// was not a permanent yet — weaker, never stronger.
//
// "For each +1/+1 counter on this creature" on death is last-known
// information (CR 603.10). The card's Counters are cleared by the
// move and the LKI characteristic the harvester hands the trigger
// carries no counter math, so the total is read back off the event
// log — b13LastKnownCounters — which is the count at the moment it
// died, a Corpsejack doubling included.
//
// One engine-side gap, not the card's: cast for X=0 the Walker is a
// printed 0/0 with no counters, which the toughness state check
// deliberately skips (the placeholder convention on Card.Power), so
// it stays on the battlefield instead of dying at once — and can
// then be grown with the tap ability, which is how the card is
// actually played, if not quite how it is printed. It makes no
// Thopters when it dies with no counters.
func init() {
	Register(Spec{
		OracleID:     "dde55256-5259-44e7-a267-fca45a7f0d04",
		Name:         "Hangarback Walker",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The X +1/+1 counters are put on the Walker as the spell resolves, a beat before it enters, so effects that watch you put counters on a permanent don't see them."},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return AddCounter{Target: item.SourceCardID, Kind: "+1/+1", N: ctx.X()}.Apply(ctx)
		},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return cardDied(ev, source)
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				n := b13LastKnownCounters(g, source.InstanceID, "+1/+1")
				return game.NewTriggeredItem(source, "Hangarback Walker — create a Thopter for each +1/+1 counter it had",
					func(g *game.Game, item *game.StackItem) error {
						if n <= 0 {
							return nil
						}
						return CreateToken{Controller: item.Controller, Template: TokenCard("1/1 colorless Thopter artifact with flying"), N: n}.Apply(NewContext(g, item))
					})
			},
		}},
		Activated: []ActivatedAbility{{
			Label: "{1}, {T}: Put a +1/+1 counter on this creature.",
			Cost:  Plus(ManaCost("{1}"), TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if z := g.FindCardZoneForEffect(item.SourceCardID); z == nil || z.Kind != game.ZoneBattlefield {
					return nil
				}
				return AddCounter{Target: item.SourceCardID, Kind: "+1/+1", N: 1}.Apply(NewContext(g, item))
			},
		}},
	})
}
