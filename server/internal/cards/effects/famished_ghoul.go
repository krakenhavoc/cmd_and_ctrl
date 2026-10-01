package effects

// Famished Ghoul — Creature — Zombie {3}{B}, 3/2:
//
//	"{1}{B}, Sacrifice this creature: Exile up to two target cards from
//	 a single graveyard."
//
// #1807, ADR 0106 §5. The Ghoul is sacrificed as a cost, so it is gone
// before the ability resolves; the ability resolves anyway (CR 113.7a,
// an ability exists independently of its source).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "22c74e4b-b1fe-4a35-acd8-fa4840af3c06",
		Name:         "Famished Ghoul",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			ExileFromASingleGraveyardAbility("{1}{B}, Sacrifice this creature", Plus(ManaCost("{1}{B}"), SacrificeThis()), 2),
		},
	})
}
