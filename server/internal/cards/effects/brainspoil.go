package effects

// Brainspoil — Sorcery {3}{B}{B} (Ravnica: City of Guilds, #5600):
//
//	"Destroy target creature that isn't enchanted. It can't be
//	 regenerated.
//	 Transmute {1}{B}{B} ({1}{B}{B}, Discard this card: Search your
//	 library for a card with the same mana value as this card, reveal
//	 it, put it into your hand, then shuffle. Transmute only as a
//	 sorcery.)"
//
// Death Rattle's destroy with no regeneration (CR 701.19c), over a
// creature with no Aura attached (Enchanted, a Role included), judged at
// announce and again at resolution (CR 608.2b): a creature that has been
// enchanted in response is no longer a legal target. Transmute is the
// keyword's constructor (transmute.go, CR 702.53), finding a mana value 5
// card from the hand at sorcery speed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "dd464a76-5f96-4e21-b7c6-4a476e7f9ce6",
		Name:         "Brainspoil",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature that isn't enchanted", Not(Enchanted())),
		OnResolve:    destroyTheTargetPermanentNoRegen,
		Activated:    []ActivatedAbility{Transmute("{1}{B}{B}")},
	})
}
