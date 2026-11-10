package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Karakyk Guardian — Creature — Dragon {3}{G}{U}{R}, 6/5 (Tarkir:
// Dragonstorm, #4463):
//
//	"Flying, vigilance, trample
//	 This creature has hexproof as long as it hasn't dealt damage yet."
//
// Ruric Thar's gate (ruric_thar_magecrusher.go, #3172) for damage of any
// kind: the hexproof exists until Card.DealtDamage is stamped, when
// damage this permanent was the source of actually lands (combat or not,
// to a player or a permanent; damage fully prevented is not dealt), and
// the stamp schedules a layer pass so the hexproof is gone before a
// trigger put on the stack by the same damage chooses its targets. It
// stays gone until the Guardian leaves the battlefield and returns as a
// new object (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "41129dc1-be88-4198-94f9-032817a79bc7",
		Name:            "Karakyk Guardian",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance", "trample"},
		Static: []game.StaticAbility{{
			Label:      "Has hexproof as long as it hasn't dealt damage yet.",
			Layer:      game.Layer6Ability,
			ActiveWhen: game.HasntDealtDamage(),
			AppliesTo:  selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				appendKeywordsTo(c, []string{"hexproof"})
			},
		}},
	})
}
