package effects

// Smaug, the Great Calamity // Spew Flame — an Adventure card (CR 715,
// The Hobbit, #4454), oracle fc75cfec…:
//
//	Smaug, the Great Calamity — Legendary Creature — Dragon {5}{R}{R}, 5/5
//	  "Flying"
//	Spew Flame — Sorcery — Adventure {4}{R}
//	  "Spew Flame deals 5 damage to target creature. (Then exile this
//	   card. You may cast the creature later from exile.)"
//
// A single creature target. A target that left in response counters the
// spell (CR 608.2b).
//
// No simplification.

// smaugTheGreatCalamityOracleID is shared by both faces of the card.
const smaugTheGreatCalamityOracleID = "fc75cfec-e1f6-4cb5-8a3c-2f7b2c39c03f"

func init() {
	Register(Spec{
		OracleID:        smaugTheGreatCalamityOracleID,
		Name:            "Smaug, the Great Calamity",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
	Register(Spec{
		OracleID:     smaugTheGreatCalamityOracleID + "#1",
		Name:         "Spew Flame",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      ForTargets(DamageToTarget(0, 5)),
		OnResolve:    damageToFirstTarget(5),
	})
}
