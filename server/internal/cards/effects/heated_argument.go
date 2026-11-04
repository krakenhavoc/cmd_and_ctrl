package effects

// Heated Argument — Instant {4}{R} (Secrets of Strixhaven, #4456):
//
//	"Heated Argument deals 6 damage to target creature. You may exile a card
//	 from your graveyard. If you do, Heated Argument also deals 2 damage to
//	 that creature's controller."
//
// Two sentences in order. The damage is dealt first; a prompt it asks (a
// damage replacement) holds the exile offer until it is answered (CR 608.2c).
// The offer is a pick of up to one card from your graveyard (the spell itself
// is still on the stack, so it is never a candidate; with an empty graveyard
// nothing is asked). Declining ends the spell. "That creature's controller"
// is the controller of the creature the spell targeted, read as the offer is
// made, and the 2 damage is dealt by the spell to that player only if a card
// was actually exiled. If the target left in response, the spell has no legal
// target and does nothing at all (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "329f6434-8f1b-4a10-8feb-2a9dc5f2712b",
		Name:         "Heated Argument",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: ResolveSteps(
			Instruction(socB26HeatedArgumentDamage),
			Instruction(socB26HeatedArgumentOffer)),
	})
}
