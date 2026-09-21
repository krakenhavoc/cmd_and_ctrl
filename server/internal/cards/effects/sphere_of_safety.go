package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sphere of Safety — Enchantment {4}{W}:
//
//	"Creatures can't attack you or planeswalkers you control unless
//	 their controller pays {X} for each of those creatures, where X is
//	 the number of enchantments you control."
//
// The two things Propaganda does not exercise, both in one card.
//
// # X is counted, not printed
//
// AttackTaxCounting renders the count as a generic cost string once
// per attacking creature, so a Sphere beside two other enchantments
// charges "{3}" per attacker and three attackers pay "{3}{3}{3}". The
// count is taken at DECLARATION, which is when CR 508.1a asks: an
// enchantment that enters after attackers are declared changes
// nothing, and one that leaves in response to the declaration being
// announced is already too late.
//
// It counts the Sphere ITSELF — it is an enchantment its controller
// controls — so the floor is {1}, never free. AttackTaxCounting's
// zero-means-free branch is therefore unreachable from this card, and
// is there for a count that really can bottom out.
//
// # "Or planeswalkers you control"
//
// ProtectingPlaneswalkers is the whole of that clause. Without it the
// tax would cover only direct attacks on the seat, which is
// Propaganda's narrower reading and the default — the default is the
// narrow one so a card file that forgets errs weaker than printed.
//
// Battles are not covered, and that is not an omission: no card in
// this family mentions them, and a battle's defending player is its
// PROTECTOR (CR 310.9), who is usually not the player whose
// enchantment this is.
func init() {
	Register(Spec{
		OracleID:     "92f6c063-a740-4c3c-a60a-569fd298854d",
		Name:         "Sphere of Safety",
		Completeness: CompletenessFull,
		AttackTaxes: []game.AttackTax{
			ProtectingPlaneswalkers(AttackTaxCounting(
				func(q game.AttackTaxQuery) int { return sphereOfSafetyEnchantments(q) },
				"Creatures can't attack you or planeswalkers you control unless their controller pays {X} for each of those creatures, where X is the number of enchantments you control.",
			)),
		},
	})
}

// sphereOfSafetyEnchantments is "the number of enchantments you
// control", counted off the live battlefield with post-layer types —
// so an Opalescence'd enchantment creature and a land that became an
// enchantment both count, and a permanent that stopped being one does
// not.
//
// Walks the live slice rather than a copy, the contract every
// board-counting predicate in the catalog has: it runs inside the
// declaration path under the write lock, and q.Game is read-only.
func sphereOfSafetyEnchantments(q game.AttackTaxQuery) int {
	if q.Game == nil || q.Game.Battlefield == nil {
		return 0
	}
	n := 0
	for i := range q.Game.Battlefield.Cards {
		c := &q.Game.Battlefield.Cards[i]
		if c.Controller == q.Defender && c.IsEnchantment() {
			n++
		}
	}
	return n
}
