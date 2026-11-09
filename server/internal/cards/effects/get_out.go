package effects

// Get Out — Instant {U}{U} (Duskmourn):
//
//	"Choose one —
//	 • Counter target creature or enchantment spell.
//	 • Return one or two target creatures and/or enchantments you own to
//	   your hand."
//
// A one-of modal spell (CR 700.2), each bullet with its own clause. The
// return bullet's clause is ownership, not control (YouOwn): it names a
// permanent of yours another player controls, and it goes to its owner's
// hand, which is yours either way. A target that is no longer a legal one is
// skipped at resolution (CR 608.2b); the rest still return.
func init() {
	Register(Spec{
		OracleID:     "b2f3c27e-b646-4449-901a-8b809cdfa84e",
		Name:         "Get Out",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeDoing("Counter target creature or enchantment spell.",
				TargetSpell("target creature or enchantment spell", Or(Creature(), Enchantment())),
				CounterTheModesTarget),
			ModeDoing("Return one or two target creatures and/or enchantments you own to your hand.",
				TargetPermanent("one or two target creatures and/or enchantments you own",
					Or(Creature(), Enchantment()), YouOwn()).WithCount(1, 2),
				bounceTheModesTargets),
		),
	})
}
