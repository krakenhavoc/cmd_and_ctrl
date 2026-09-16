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
// The white X-drop that grows itself and then the team. The X body,
// the tap-to-grow and the team pump are live. The pump's cost removes
// a +1/+1 counter from Mikaeus at announce (RemoveCountersFromThis,
// #625), with the tap, so a response cannot see Mikaeus still holding
// a counter he has already spent. Paying with his LAST counter kills
// him, as printed: the removal marks him Card.LostLastCounter, so the
// toughness state check does not mistake the 0/0 left behind for a
// placeholder, and he is put into the graveyard (CR 704.5f) with the
// pump still on the stack, where it resolves anyway.
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
// One engine-side gap, not the card's: cast for X=0 Mikaeus is a
// printed 0/0 that never had a counter, which the toughness state
// check deliberately skips as a placeholder, so he stays on the
// battlefield and can be grown with the tap. Declared to players as
// the second caveat: it is the one way this card plays stronger than
// printed.
func init() {
	Register(Spec{
		OracleID:     "82f3faa8-39fa-450b-843f-d60a4c36d8f7",
		Name:         "Mikaeus, the Lunarch",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The X +1/+1 counters are put on Mikaeus as the spell resolves, a beat before he enters, so effects that watch you put counters on a permanent don't see them.",
			"Cast for X=0, Mikaeus stays on the battlefield as a 0/0 instead of dying at once, and can tap to grow.",
		},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return AddCounter{Target: item.SourceCardID, Kind: "+1/+1", N: ctx.X()}.Apply(ctx)
		},
		Activated: []ActivatedAbility{
			{
				Label:  "{T}: Put a +1/+1 counter on Mikaeus.",
				Cost:   TapCost(),
				Effect: putCounterOnSelf,
			},
			{
				Label:  "{T}, Remove a +1/+1 counter from Mikaeus: Put a +1/+1 counter on each other creature you control.",
				Cost:   Plus(TapCost(), RemoveCountersFromThis(game.CounterPlusOne, 1)),
				Effect: putCounterOnEachOtherCreatureYouControl,
			},
		},
	})
}
