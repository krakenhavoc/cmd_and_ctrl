package game

import (
	"github.com/google/uuid"
)

// mana_ability_resolved.go — EventManaAbilityResolved (#5644): one event
// per mana ability that finished resolving, carrying how much mana that
// ability produced.
//
// EventManaAbilityActivated is emitted when the cost is paid, before the
// mana exists, and EventManaAdded is one event per mana with no ability
// around it. Neither can answer "whenever a mana ability of this creature
// resolves, put a number of +1/+1 counters on it equal to the amount of
// mana this creature produced" (Tyvar the Bellicose). This event can.
//
// # What is counted
//
// Amount is the mana that ability put in a pool, after the CR 106.12b
// window: a Forest under Mana Reflection resolved for 2, an Ancient Tomb
// under Damping Sphere for 1. Mana a DIFFERENT ability added because of
// it (Wild Growth's triggered mana ability, CR 605.1b) is not in it; that
// ability resolves on its own and announces its own event, with the
// trigger's permanent as Source. An ability that resolved and added
// nothing (CR 903.4f, a commander with no colour identity) still
// resolved, with Amount 0.
//
// # Where it is emitted
//
// Once per resolution, by the five places a mana ability finishes:
//
//   - ActivateManaAbility, after the rider, when no colour was left to
//     pick;
//   - ResolveManaChoice, on the last answer of an ability whose colour
//     was picked (PendingChoice.ManaAbilityRun);
//   - materializePlanLocked and materializeExiledManaSourceLocked, the
//     auto-tapper's two arms, once per planned source;
//   - fireManaTriggersLocked, once per triggered mana ability.
//
// An activated ability's event comes before the triggered mana abilities
// its production fires, which is the order they resolve in.

// emitManaAbilityResolvedLocked announces that a mana ability of `source`
// finished resolving and produced `amount` mana. `actor` is the player
// who activated it, or for a triggered mana ability the player whose
// pool its mana went to. Caller must hold g.mu in write mode.
func (g *Game) emitManaAbilityResolvedLocked(actor, source uuid.UUID, label string, amount int) {
	if amount < 0 {
		amount = 0
	}
	g.EmitEvent(Event{
		Kind:   EventManaAbilityResolved,
		Actor:  actor,
		Source: source,
		CardID: source,
		Label:  label,
		Amount: amount,
	})
}

// manaAbilityPickAnsweredLocked closes one colour pick of a mana ability
// (`run` is its PendingChoice.ManaAbilityRun) that added `added` mana on
// top of the `base` the pick carried. While another pick of the same
// ability is still open the count moves onto it; the last one announces
// EventManaAbilityResolved with the total. The answered pick is already
// off the queue. Caller must hold g.mu in write mode.
func (g *Game) manaAbilityPickAnsweredLocked(run string, actor, source uuid.UUID, label string, base, added int) {
	if run == "" {
		return
	}
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == PendingChoiceMana && c.ManaAbilityRun == run {
			c.ManaAbilityAmount += base + added
			return
		}
	}
	g.emitManaAbilityResolvedLocked(actor, source, label, base+added)
}

// manaTriggerResolvedLocked is fireManaTriggersLocked's announcement for
// one triggered mana ability. `poolBefore` is the size of the player's
// pool and `choicesBefore` the length of the choice queue before the
// ability added its mana. With no colour left to pick the ability has
// resolved, and what the pool gained is what it produced. Otherwise each
// pick it queued joins one run, the first carrying what already landed,
// and the last answer announces it (manaAbilityPickAnsweredLocked).
// Caller must hold g.mu in write mode.
func (g *Game) manaTriggerResolvedLocked(playerID, source uuid.UUID, label string, poolBefore, choicesBefore int) {
	p := g.playerByIDLocked(playerID)
	if p == nil || p.Eliminated {
		return
	}
	added := p.ManaPool.Total() - poolBefore
	if choicesBefore >= len(g.PendingChoices) {
		g.emitManaAbilityResolvedLocked(playerID, source, label, added)
		return
	}
	run := uuid.NewString()
	for _, c := range g.PendingChoices[choicesBefore:] {
		if c == nil || c.Kind != PendingChoiceMana {
			continue
		}
		c.ManaAbilityRun = run
		if c.ManaLabel == "" {
			c.ManaLabel = label
		}
		c.ManaAbilityAmount, added = added, 0
	}
}
