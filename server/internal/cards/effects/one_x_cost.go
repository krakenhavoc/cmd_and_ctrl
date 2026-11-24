package effects

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// one_x_cost.go — one announced X paying an {X} mana symbol and a
// non-mana X in the same activation cost (#6085).
//
// CR 107.3a: "While an activated ability is on the stack, any X in its
// activation cost equals the announced value." Two printed shapes:
//
//	Necron Overlord  {X}, {T}, Tap X untapped artifacts you control:
//	                 TapXUntapped beside ManaCost("{X}"). The engine
//	                 already counts the taps and prices the {X} from the
//	                 one ActivateAbilityParams.XValue, so Register stops
//	                 refusing the pair.
//	Talon of Pain    {X}, {T}, Remove X charge counters from this
//	                 artifact: RemoveCountersXFromThisSameX, whose count
//	                 is the announced X rather than a second number the
//	                 activator names (game.CounterRemovalCost.CountIsX).
//
// The other pairs ({X} beside "Sacrifice X", "Discard X", "Reveal X")
// are still refused where they were: no printed card asks for them.

// RemoveCountersXFromThisSameX is "Remove X <kind> counters from this
// permanent" in a cost that also prints {X} in its mana component
// (Talon of Pain): the counters removed are the announced X, and the
// effect may read either ctx.X() or ctx.CountersRemoved(), which agree.
// X = 0 is legal and removes nothing; an X above the counters held is
// refused at announce. Register holds it to an activated ability whose
// mana cost has an {X}.
func RemoveCountersXFromThisSameX(kind string) game.AbilityCost {
	c := RemoveCountersXFromThis(kind, 0)
	c.RemoveCounters.CountIsX = true
	return c
}

// checkOneXCost is Register's guard for the CountIsX removal on an
// activated ability: it needs an {X} in the mana cost to share (without
// one there is no announced X to read, and RemoveCountersXFromThis is
// the shape), and it is the self, printed-kind, variable form with a
// floor of 0 — the only one a printed card has.
func checkOneXCost(card string, i int, cost game.AbilityCost) {
	rc := cost.RemoveCounters
	if !game.CounterCountIsX(rc) {
		return
	}
	switch {
	case cost.XSlots() == 0:
		panic(fmt.Sprintf("effects.Register: %q ability %d removes X counters as the announced X but its mana cost %q has no {X} — use RemoveCountersXFromThis", card, i, cost.Mana))
	case !rc.Variable || rc.N != 0 || rc.From != nil || rc.Among || rc.All || rc.OnePermanent || rc.Counter == "":
		panic(fmt.Sprintf("effects.Register: %q ability %d marks a counter removal CountIsX outside \"Remove X <kind> counters from this permanent\" — build it with RemoveCountersXFromThisSameX", card, i))
	case game.SacrificeCountFromX(cost.SacrificeOther) || game.TapOthersCountFromX(cost.TapOthers) ||
		game.DiscardCountFromX(cost.DiscardCards) || game.RevealCardsCountFromX(cost.RevealCards) ||
		cost.EnergyX || cost.LifeX || cost.LoyaltyX:
		panic(fmt.Sprintf("effects.Register: %q ability %d removes X counters beside another X-counted component — no printed card does", card, i))
	}
}

// refuseCounterCountIsX refuses the CountIsX removal where no X is
// announced: a mana ability announces none (CR 605.3b).
func refuseCounterCountIsX(card, where string, rc *game.CounterRemovalCost) {
	if game.CounterCountIsX(rc) {
		panic(fmt.Sprintf("effects.Register: %q %s removes X counters as an announced X, and it announces none — use RemoveCountersXFromThis", card, where))
	}
}

// checkTriggerXFromEvent holds game.TriggeredAbility.XFromEventAmount
// ("destroy up to X target artifacts", X the cycling cost's X that the
// triggering event carries) to the one shape the engine sizes: a
// non-modal row whose own clause counts from X. On any other row the flag
// would size nothing and the clause would read as unbounded (#259).
func checkTriggerXFromEvent(card string, t game.TriggeredAbility) {
	if !t.XFromEventAmount {
		return
	}
	switch {
	case t.XFromCast:
		panic(fmt.Sprintf("effects.Register: %q trigger %q reads X from both its source's cast and its event — one X", card, t.Key))
	case t.Modes != nil || t.Targets == nil || !t.Targets.CountFromX || t.TargetsFrom != nil:
		panic(fmt.Sprintf("effects.Register: %q trigger %q sets XFromEventAmount without its own CountFromX target clause — build it with UpToXTargetsFromTheEventAmount", card, t.Key))
	}
}

// UpToXTargetsFromTheEventAmount marks a trigger's clause as "up to X
// target …" where X is the triggering event's Amount (#6085): Rampaging
// War Mammoth's "When you cycle this card, destroy up to X target
// artifacts", X being the cycling cost's X that EventCycle carries. The
// engine sizes the clause as the trigger is put on the stack
// (game.TriggeredAbility.XFromEventAmount), and X = 0 has no targets.
func UpToXTargetsFromTheEventAmount(t game.TriggeredAbility) game.TriggeredAbility {
	if t.Targets == nil {
		panic("effects.UpToXTargetsFromTheEventAmount: the trigger has no target clause")
	}
	spec := *t.Targets
	spec.CountFromX, spec.UpToX = true, true
	spec.Min, spec.Max = 0, 0
	t.Targets = &spec
	t.XFromEventAmount = true
	return t
}
