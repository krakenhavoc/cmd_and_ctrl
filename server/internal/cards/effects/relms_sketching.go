package effects

// Relm's Sketching — Sorcery {2}{U}{U} (EDHREC rank 4046):
//
//	"Create a token that's a copy of target artifact, creature, or
//	 land."
//
// A Clone that is not restricted to creatures, which is the whole
// reason a deck plays it over one: the best copy target in a
// Commander game is often a Sol Ring, a Gaea's Cradle or a Dockside
// Extortionist that has already left. It is in the batch as the
// widest copy target clause in the catalog — three card types in one
// slot, anybody's permanent.
//
// # The token is a copy, not a clone of the board state
//
// It copies the printed, COPIABLE values (CR 707.2): the original's
// name, mana cost, types, abilities and printed power and toughness.
// It does not copy counters, Auras and Equipment attached to it, the
// damage it has taken, or anything a continuous effect is doing to
// it. So a token copy of a creature carrying four +1/+1 counters is
// the printed size, and a token copy of a land animated by a Nissa is
// just the land.
//
// A copy of a LEGENDARY permanent you already control dies to the
// legend rule immediately — the card prints no "except it isn't
// legendary" clause, unlike Vesuvan Duplimancy, and the difference
// matters enough that the two are worth reading side by side.
//
// The token runs the ordinary entry pipeline, so a copied permanent's
// enters-the-battlefield triggers fire and its "as this enters,
// choose …" clause is asked.
//
// "Target" means the usual re-check: a permanent that leaves in
// response takes the spell with it, and the token is never created.
// That body is Cackling Counterpart's word for word — the target
// clause is the only thing the two cards do not share — so both go
// through TokenCopyOfSingleTarget.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "dd6601c6-810d-4883-89d2-5d31419fb1cc",
		Name:         "Relm's Sketching",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target artifact, creature, or land", Or(Artifact(), Creature(), Land())),
		OnResolve:    TokenCopyOfSingleTarget,
	})
}
