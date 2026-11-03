package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Brass Infiniscope — Artifact {4} (Secrets of Strixhaven Commander,
// #4834):
//
//	"{T}: Add {C}{C}. When you next cast a spell with {X} in its mana
//	 cost this turn, you draw a card and gain half X life, rounded down."
//
// Solar Array's shape: a mana ability whose effect also creates a
// delayed triggered ability (CR 603.7). The trigger is data, a
// registered condition and body, so it survives a restore point, and it
// ends at cleanup whether or not it fired. It triggers only once, the
// first time its event occurs (CR 603.7b); a cast of a spell with no {X}
// in its mana cost is not that event and does not use it up. Tapped
// while paying for an X spell, it is already waiting when that spell
// becomes cast (CR 601.2i), so it fires for that very spell.
//
// "{X} in its mana cost" is the spell's own mana cost, parsed: an {X} in
// an additional or alternative cost does not count. X is the value
// announced for the spell (CR 107.3a), read when the trigger resolves
// off the spell's stack item, or off its last-known record if it was
// countered in response (CR 608.2h, game.SpellXValueForEffect). Half of
// it is rounded down, as printed (CR 107.1a).
func init() {
	Register(Spec{
		OracleID:     "f74d8f53-a239-4641-90f1-bf786d57e253",
		Name:         "Brass Infiniscope",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}{C}",
			Label:    "{T}: Add {C}{C}. When you next cast a spell with {X} in its mana cost this turn, you draw a card and gain half X life, rounded down",
			Rider:    brassInfiniscopeNextXSpell,
		}},
	})
}

var (
	// The event condition of "when you next cast a spell with {X} in its
	// mana cost this turn": an EventCast by the trigger's controller whose
	// spell has an {X} in its mana cost.
	youNextCastAnXSpellCondition = game.DelayedCondition("cast/you-next-cast-x-in-mana-cost", youNextCastAnXSpell)

	// "You draw a card and gain half X life, rounded down."
	brassInfiniscopeDrawAndGainBody = game.SimpleDelayedBody("brass-infiniscope/draw-and-gain-half-x", brassInfiniscopeDrawAndGain)
)

// brassInfiniscopeNextXSpell schedules the delayed trigger.
func brassInfiniscopeNextXSpell(g *game.Game, controller, source uuid.UUID) error {
	g.ScheduleDelayedTriggerForEffect(game.DelayedTrigger{
		Controller:   controller,
		SourceCardID: source,
		Label:        "Brass Infiniscope — draw a card and gain half X life",
		On:           []game.EventKind{game.EventCast},
		Condition:    youNextCastAnXSpellCondition,
		Body:         brassInfiniscopeDrawAndGainBody,
	})
	return nil
}

// youNextCastAnXSpell is youNextCast whose spell has an {X} in its mana
// cost.
func youNextCastAnXSpell(ev game.Event, dt *game.DelayedTrigger, g *game.Game, p game.EffectParams) bool {
	if !youNextCast(ev, dt, g, p) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	if !ok {
		return false
	}
	cost, err := game.ParseCost(c.ManaCost)
	return err == nil && cost.XSlots > 0
}

func brassInfiniscopeDrawAndGain(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	x := 0
	if cast := ctx.PayloadCards(); len(cast) > 0 {
		x, _ = g.SpellXValueForEffect(cast[0])
	}
	if err := (DrawCards{Player: item.Controller, N: 1}).Apply(ctx); err != nil {
		return err
	}
	return GainLife{Player: item.Controller, Amount: x / 2}.Apply(ctx)
}
