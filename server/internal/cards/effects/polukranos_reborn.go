package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Polukranos Reborn // Polukranos, Engine of Ruin — {G}{G}{G} Legendary
// Creature — Hydra 4/5 // Legendary Creature — Phyrexian Hydra 6/6 (March
// of the Machine):
//
//	Front: "Reach
//	        {6}{W/P}: Transform Polukranos Reborn. Activate only as a
//	        sorcery."
//	Back:  "Reach, lifelink
//	        Whenever Polukranos or another nontoken Hydra you control dies,
//	        create a 3/3 green and white Phyrexian Hydra creature token
//	        with reach and a 3/3 green and white Phyrexian Hydra creature
//	        token with lifelink."
//
// One trigger with two conditions: Polukranos itself dying, or another
// nontoken Hydra you controlled dying. Hydra is read from the dead
// creature's last-known subtypes (CR 603.10a), so a changeling or a
// type-granting effect counts as it did on the battlefield. Both tokens
// are made by the one resolution.
//
// No simplification.
func init() {
	const oracle = "7a9fa906-b43e-4d98-b134-5936d3085851"
	Register(Spec{
		OracleID:        oracle,
		Name:            "Polukranos Reborn",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Activated: []ActivatedAbility{{
			Label:        "{6}{W/P}: Transform Polukranos Reborn. Activate only as a sorcery.",
			Cost:         ManaCost("{6}{W/P}"),
			SorcerySpeed: true,
			Effect:       Do(TransformThis{}),
		}},
	})
	Register(Spec{
		OracleID:        oracle + "#1",
		Name:            "Polukranos, Engine of Ruin",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach", "lifelink"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
				if ThisDied(ev, source, ch, g) {
					return true
				}
				dead, ok := diedCreature(ev, g)
				return ok && !IsToken(dead) && leftUnderControlOf(ev, dead) == source.Controller &&
					leftAsSubtype(ev, dead, source.TextSubtype("Hydra"))
			},
			Key: "Polukranos, Engine of Ruin — create a 3/3 Phyrexian Hydra with reach and a 3/3 Phyrexian Hydra with lifelink",
			Effect: Do(
				CreateToken{Template: TokenCard("3/3 green and white Phyrexian Hydra with reach"), N: 1},
				CreateToken{Template: TokenCard("3/3 green and white Phyrexian Hydra with lifelink"), N: 1},
			),
		}},
	})
}
