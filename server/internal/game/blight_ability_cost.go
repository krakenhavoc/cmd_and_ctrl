package game

import "github.com/google/uuid"

// blight_ability_cost.go — #4684: "Blight N" as part of an activated
// ability's cost (CR 701.68a, paid at CR 602.2b's run of CR 601.2h).
//
//	Gristle Glutton       "{T}, Blight 1: Discard a card. If you do,
//	                      draw a card."
//	Sting-Slinger         "{1}{R}, {T}, Blight 1: This creature deals 2
//	                      damage to each opponent."
//	Evershrike's Gift     "{1}{W}, Blight 2: Return this card from your
//	                      graveyard to your hand." (from the graveyard)
//	Spiral into Solitude  "{1}{W}, Blight 1, Sacrifice this Aura: Exile
//	                      enchanted creature."
//
// It is the cast path's blight cost (teamwork_blight_cost.go, #1703) on
// an ability, and shares every piece of it:
//
//   - the candidates are BlightOptionsForEffect, the creatures the
//     ACTIVATOR controls that can have counters put on them. Any of
//     them, the source included when it is one: "a creature you
//     control", not "this creature" (that is AddCounter, Devoted
//     Druid's). Not a targeting walk, so hexproof doesn't matter;
//   - one creature, checked by validateBlightCreatureLocked;
//   - paid by blightLocked, which marks the counters CounterFromCost.
//     A cost is not "the effect of a resolving spell or ability" (CR
//     614.16), so a replacement worded "if an effect would put"
//     (Doubling Season) leaves them alone, and one that names no effect
//     (Vizier of Remedies, Winding Constrictor) replaces the event as it
//     would any other. A cost can't pause, so two such replacements
//     apply in gathered order.
//
// CR 701.68b: a player who can't put the counters on a creature they
// control can't choose to blight, so with no candidate the cost can't
// be paid and the ability can't be activated (CR 118.3). A 1/1 is a
// candidate for blight 2: it dies of the counters at the next
// state-based check (CR 704.5f), after the cost is paid.
//
// The source's zone doesn't matter: a graveyard ability (Evershrike's
// Gift) blights a creature on the battlefield.
//
// THE COMPONENT IS DATA (an int), so no func-typed route from Game is
// added (ADR 0041's closure ratchet).

// validateAbilityBlightLocked checks the creature an activation names
// for its Blight component, without placing anything.
//
//   - IDs sent for an ability with no blight are refused rather than
//     ignored, as a stray unattach_ids is.
//   - With nothing the activator could blight, the cost can't be paid
//     (CR 701.68b): ErrCantPayCounterCost, the counter cost's refusal.
//   - Otherwise exactly one creature, on the battlefield, the
//     activator's, that can have counters put on it.
//
// Caller must hold g.mu.
func (g *Game) validateAbilityBlightLocked(playerID uuid.UUID, n int, ids []uuid.UUID) error {
	if n <= 0 {
		if len(ids) > 0 {
			return ErrInvalidParam
		}
		return nil
	}
	if len(g.BlightOptionsForEffect(playerID)) == 0 {
		return ErrCantPayCounterCost
	}
	if len(ids) != 1 {
		return ErrInvalidParam
	}
	return g.validateBlightCreatureLocked(playerID, ids[0])
}

// payAbilityBlightLocked puts the cost's n -1/-1 counters on the named
// creature through blightLocked (CounterFromCost, settled at once). The
// cost is paid whatever the window settles on: a Vizier of Remedies
// that takes the one counter away still leaves the blight paid.
// Validated above; a creature that has left in between (nothing can
// move it under one lock today) is skipped rather than erroring.
//
// Caller must hold g.mu.
func (g *Game) payAbilityBlightLocked(playerID uuid.UUID, n int, ids []uuid.UUID) error {
	if n <= 0 || len(ids) != 1 || findBattlefieldCard(g, ids[0]) == nil {
		return nil
	}
	_, err := g.blightLocked(playerID, ids[0], n)
	return err
}
