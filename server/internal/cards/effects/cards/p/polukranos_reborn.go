package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        oracle,
		Name:            "Polukranos Reborn",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Activated: []effects.ActivatedAbility{{
			Label:        "{6}{W/P}: Transform Polukranos Reborn. Activate only as a sorcery.",
			Cost:         effects.ManaCost("{6}{W/P}"),
			SorcerySpeed: true,
			Effect:       effects.Do(effects.TransformThis{}),
		}},
	})
	effects.Register(effects.Spec{
		OracleID:        oracle + "#1",
		Name:            "Polukranos, Engine of Ruin",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach", "lifelink"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
				if effects.ThisDied(ev, source, ch, g) {
					return true
				}
				dead, ok := effects.DiedCreature(ev, g)
				return ok && !effects.IsToken(dead) && effects.LeftUnderControlOf(ev, dead) == source.Controller &&
					effects.LeftAsSubtype(ev, dead, source.TextSubtype("Hydra"))
			},
			Key: "Polukranos, Engine of Ruin — create a 3/3 Phyrexian Hydra with reach and a 3/3 Phyrexian Hydra with lifelink",
			Effect: effects.Do(
				effects.CreateToken{Template: effects.TokenCard("3/3 green and white Phyrexian Hydra with reach"), N: 1},
				effects.CreateToken{Template: effects.TokenCard("3/3 green and white Phyrexian Hydra with lifelink"), N: 1},
			),
		}},
	})
}
