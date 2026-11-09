package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Molten Exhale — Sorcery {1}{R} (Tarkir: Dragonstorm, #4463):
//
//	"You may cast this spell as though it had flash if you behold a
//	 Dragon as an additional cost to cast it. (To behold a Dragon,
//	 choose a Dragon you control or reveal a Dragon card from your
//	 hand.)
//	 Molten Exhale deals 4 damage to target creature or planeswalker."
//
// CR 601.3c: a spell that may be cast as though it had flash only if an
// additional cost is paid may BEGIN to be cast at instant speed. The
// behold is an optional additional cost (OptionalBeholdOpeningFlash,
// #5207) whose announcement opens the instant-speed window for that cast
// (AdditionalCost.AsThoughFlash, Quantum Reduction's shape): the caster
// names one Dragon they control or one Dragon card in their hand on
// reveal_ids, a card from hand is revealed, and nothing moves (CR
// 701.4a). Cast without it, the spell keeps sorcery timing. It is not an
// alternative cost, so a flash cast may still claim one (CR 118.9a).
// Nothing on the card reads whether a Dragon was beheld.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "82aa1402-4b1c-4343-aba5-dd3271cc1ac5",
		Name:          "Molten Exhale",
		Completeness:  CompletenessFull,
		Purpose:       ForTargets(DamageToTarget(0, 4)),
		OptionalCosts: []game.AdditionalCost{OptionalBeholdOpeningFlash("a", "Dragon")},
		Targets:       TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
		OnResolve:     damageToFirstTarget(4),
	})
}
