package effects

// Transcendent Master — Creature — Human Cleric Avatar {1}{W}{W}, 3/3:
//
//	"Level up {1}
//	 LEVEL 6-11: 6/6, Lifelink
//	 LEVEL 12+: 9/9, Lifelink, indestructible"
//
// A leveler (CR 702.87, 711; leveler.go): level up is the sorcery-speed
// activation, each band the statics a card with that sentence printed in
// full would carry, so below level 6 it is its printed 3/3 (CR 711.5).
func init() {
	Register(Spec{
		OracleID:     "11332981-f06a-48c7-a7b4-fce68e629e82",
		Name:         "Transcendent Master",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{LevelUpKeyword("{1}")},
		Static: LevelBands(
			LevelBand{From: 6, To: 11, Power: 6, Toughness: 6, Keywords: []string{"lifelink"}},
			LevelBand{From: 12, Power: 9, Toughness: 9, Keywords: []string{"lifelink", "indestructible"}},
		),
	})
}
