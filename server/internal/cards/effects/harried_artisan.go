package effects

// Harried Artisan // Phyrexian Skyflayer — a Phyrexian transformer (March of the Machine):
//
//	Harried Artisan {2}{R}, Creature — Human Artificer, 2/3
//	  "Haste
//	   {3}{W/P}: Transform this creature. Activate only as a sorcery."
//
//	Phyrexian Skyflayer, Creature — Phyrexian Artificer, 3/4
//	  "Flying, haste"
//
// The activation turns the permanent over in place (Transform, ADR
// 0079); the Phyrexian mana in its cost is paid with W or 2 life.
//
// No simplification.
func init() {
	const oracle = "c4e51cba-e5f7-4f44-9ae9-9223a2efd004"
	Register(Spec{
		OracleID:        oracle,
		Name:            "Harried Artisan",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Activated: []ActivatedAbility{{
			Label:        "{3}{W/P}: Transform this creature. Activate only as a sorcery.",
			Cost:         ManaCost("{3}{W/P}"),
			SorcerySpeed: true,
			Effect:       Do(TransformThis{}),
		}},
	})
	Register(Spec{
		OracleID:        oracle + "#1",
		Name:            "Phyrexian Skyflayer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
	})
}
