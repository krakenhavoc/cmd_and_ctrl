package effects

// Wall of Denial — Creature — Wall {1}{W}{U}, 0/8 (EDHREC rank
// 4326):
//
//	"Defender, flying
//	 Shroud (This creature can't be the target of spells or
//	 abilities.)"
//
// Three keywords and nothing else, which is the whole card: an
// eight-toughness flying blocker that removal cannot touch. In a
// format whose usual answer to a wall is a Swords to Plowshares,
// shroud is what makes the Wall stay.
//
// All three are canonical engine keywords, so PrintedKeywords is the
// entire Spec. Defender is read by the attack-declaration gate,
// flying by the block-pair check, and shroud by CanBeTargetedBy —
// including against its own controller, which is the drawback half of
// the keyword and the reason you cannot put an Aura on it.
//
// It is in the batch because the importer stamps Scryfall's keyword
// array onto every card, so a vanilla three-keyword creature needs no
// Spec at all — and a catalog entry is how that claim gets a test
// rather than a shrug.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9cc092f9-e8f0-40ea-847d-813b70f889db",
		Name:            "Wall of Denial",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender", "flying", "shroud"},
	})
}
