package effects

// Drift of Phantasms — Creature — Spirit {2}{U}, 0/5 (Ravnica: City of
// Guilds, #5600):
//
//	"Defender (This creature can't attack.)
//	 Flying
//	 Transmute {1}{U}{U} ({1}{U}{U}, Discard this card: Search your
//	 library for a card with the same mana value as this card, reveal it,
//	 put it into your hand, then shuffle. Transmute only as a sorcery.)"
//
// Two printed keywords and the transmute constructor (transmute.go, CR
// 702.53a), finding a mana value 3 card.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6ee51134-c4ad-4101-8ce2-23dde9359cd4",
		Name:            "Drift of Phantasms",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender", "flying"},
		Activated:       []ActivatedAbility{Transmute("{1}{U}{U}")},
	})
}
