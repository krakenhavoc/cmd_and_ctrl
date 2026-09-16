package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mikaeus, the Lunarch — Legendary Creature — Human Cleric {X}{W},
// 0/0 (EDHREC rank 1959):
//
//	"Mikaeus enters with X +1/+1 counters on it.
//	 {T}: Put a +1/+1 counter on Mikaeus.
//	 {T}, Remove a +1/+1 counter from Mikaeus: Put a +1/+1 counter
//	 on each other creature you control."
//
// The white X-drop that grows itself. The X body and the tap-to-grow
// are live.
//
// Sandbox simplification, declared, for the X counters — Goldvein
// Hydra's: an entry replacement cannot see the X announced for the
// spell (the stack item is gone by the time the entry pipeline
// runs), so the counters go on as the spell RESOLVES, a beat before
// the card moves from the stack to the battlefield, which is the
// last moment X is readable. They are on the card when it lands, so
// the 0/0 body never meets the state-based check without them, and
// a counter doubler applies. The one observable difference is that a
// "whenever you put counters on a permanent" payoff does not see
// them — weaker, never stronger.
//
// DECLARED SIMPLIFICATION, the Dragon's Hoard posture: the third
// ability is not implemented. "Remove a +1/+1 counter from Mikaeus"
// is a cost component the engine cannot express — AbilityCost
// carries tap, sacrifice, mana, life, loyalty and crew, and nothing
// that removes a counter — and shipping the team pump with the
// removal deferred to resolution would let a response bank a counter
// the printed card had already spent, which is the #259 direction.
// Mikaeus grows and stops there until a counter-removal cost lands.
//
// One engine-side gap, not the card's: cast for X=0 Mikaeus is a
// printed 0/0 with no counters, which the toughness state check
// deliberately skips, so he stays on the battlefield and can be
// grown with the tap.
func init() {
	Register(Spec{
		OracleID:     "82f3faa8-39fa-450b-843f-d60a4c36d8f7",
		Name:         "Mikaeus, the Lunarch",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The X +1/+1 counters are put on Mikaeus as the spell resolves, a beat before he enters, so effects that watch you put counters on a permanent don't see them.",
			"Removing a +1/+1 counter to put one on each other creature you control isn't implemented — Mikaeus can only grow himself.",
		},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return AddCounter{Target: item.SourceCardID, Kind: "+1/+1", N: ctx.X()}.Apply(ctx)
		},
		Activated: []ActivatedAbility{{
			Label:  "{T}: Put a +1/+1 counter on Mikaeus.",
			Cost:   TapCost(),
			Effect: putCounterOnSelf,
		}},
	})
}
