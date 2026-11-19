package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sigurd, Jarl of Ravensthorpe — Legendary Creature — Human Warrior {R}{G}{W},
// 3/3 (Assassin's Creed, #5948):
//
//	"Vigilance, trample, lifelink
//	 Boast — {1}: Put a lore counter on target Saga you control or remove one
//	 from it. (Activate only if this creature attacked this turn and only once
//	 each turn.)
//	 Whenever you put a lore counter on a Saga you control, put a +1/+1 counter
//	 on up to one other target creature."
//
// Boast is Axgard Braggart's constructor (the engine reads the attack record
// and the activation tally). The Saga is the ability's target; whether the
// counter goes on or comes off is chosen as it resolves, a two-way pick
// (Archfiend of Spite's), and removing one with none on it does nothing.
// Removing a lore counter puts nothing on the stack (CR 714.2b); putting one
// triggers the Saga's next chapter in the ordinary way.
//
// The trigger watches counters placed: a lore counter you put on a Saga you
// control, from this boast, a turn-based precombat-main lore counter, or any
// other effect, once per placement. Its target (up to one other creature) is
// chosen as it goes on the stack.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "eedd74c4-42d1-41ba-9051-dc518e45ebbb",
		Name:            "Sigurd, Jarl of Ravensthorpe",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance", "trample", "lifelink"},
		Activated: []ActivatedAbility{
			BoastTargeting("{1}: Put a lore counter on target Saga you control or remove one from it.",
				ManaCost("{1}"),
				TargetPermanent("target Saga you control", OfSubtype("Saga"), YouControl()),
				sigurdBoast),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(On(game.EventCounterPlaced, sigurdLorePlaced,
				"Sigurd, Jarl of Ravensthorpe — put a +1/+1 counter on up to one other target creature",
				b36CounterOnChosenAnimal),
				Another(TargetCreature("up to one other target creature").WithCount(0, 1))),
		},
	})
}

// sigurdLorePlaced is "you put a lore counter on a Saga you control".
func sigurdLorePlaced(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventCounterPlaced || ev.Label != game.CounterLore || ev.Amount <= 0 || ev.Actor != source.Controller {
		return false
	}
	saga, ok := g.LookupCardForEffect(ev.Target)
	return ok && onBattlefield(g, ev.Target) && saga.Controller == source.Controller && saga.HasSubtype("Saga")
}

func sigurdBoast(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return PickOption{
		Question: "Sigurd, Jarl of Ravensthorpe — put a lore counter on the Saga, or remove one from it?",
		Options: []game.ChoiceOption{
			{Label: "Put a lore counter on it"},
			{Label: "Remove a lore counter from it"},
		},
		ThenKey: sigurdBoastAnswered,
		Carry:   []uuid.UUID{id},
	}.Apply(ctx)
}

// sigurdBoastAnswered is the choice, keyed. An on-disk key.
var sigurdBoastAnswered = OptionPickThen("option-pick/sigurd-boast-lore", func(ctx *Context, r game.OptionPicked) error {
	if len(r.Carry) == 0 || !onBattlefield(ctx.Game, r.Carry[0]) {
		return nil
	}
	delta := 1
	if r.Index == 1 {
		c, ok := ctx.Game.LookupCardForEffect(r.Carry[0])
		if !ok || c.Counters[game.CounterLore] <= 0 {
			return nil
		}
		delta = -1
	}
	return AddCounter{Target: r.Carry[0], Kind: game.CounterLore, N: delta}.Apply(ctx)
})
