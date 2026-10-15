package effects

// Hypnotic Sprite // Mesmeric Glare — an Adventure card (CR 715):
//
//	Hypnotic Sprite — Creature — Faerie {U}{U}, 2/1
//	  "Flying"
//	Mesmeric Glare — Instant — Adventure {2}{U}
//	  "Counter target spell with mana value 3 or less. (Then exile this
//	   card. You may cast the creature later from exile.)"
//
// Two keys, one card: the creature keeps the bare oracle_id and the
// Adventure takes "<oracle_id>#1" (see Bonecrusher Giant). The counter
// is Disdainful Stroke's shape with the bound the other way round: the
// spell's mana value is read on the stack (CR 202.3e), so an X spell
// counts what was paid for X, and the clause is re-checked on
// resolution.
//
// No simplification.
const hypnoticSpriteOracleID = "1f9efad1-9f51-47aa-a5a7-9c8103435a01"

func init() {
	Register(Spec{
		OracleID:        hypnoticSpriteOracleID,
		Name:            "Hypnotic Sprite",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
	Register(Spec{
		OracleID:     hypnoticSpriteOracleID + "#1",
		Name:         "Mesmeric Glare",
		Completeness: CompletenessFull,
		Targets:      TargetSpell("target spell with mana value 3 or less", ManaValueLE(3)),
		OnResolve:    counterTheTargetSpell,
	})
}
