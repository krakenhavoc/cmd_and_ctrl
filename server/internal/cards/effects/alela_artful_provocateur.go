package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Alela, Artful Provocateur — Legendary Creature — Faerie Warlock
// {1}{W}{U}{B}, 2/3:
//
//	"Flying, deathtouch, lifelink
//	 Other creatures you control with flying get +1/+0.
//	 Whenever you cast an artifact or enchantment spell, create a 1/1
//	 blue Faerie creature token with flying."
//
// The anthem is Windstorm Drake's layer 7c static (flying is granted in
// layer 6, so a creature that has flying by another effect is pumped).
// The cast trigger fires on the cast, so the Faerie is on the
// battlefield before the spell resolves (CR 603.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "936ae5dc-9838-47f3-ba1f-66523a7f5b76",
		Name:            "Alela, Artful Provocateur",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "deathtouch", "lifelink"},
		Static: []game.StaticAbility{{
			Label:     "Other creatures you control with flying get +1/+0.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: b29OtherCreaturesYouControlWithFlying,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power++
			},
		}},
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Or(Artifact(), Enchantment()), "Alela — create a 1/1 blue Faerie creature token with flying",
				Do(CreateToken{Template: TokenCard("1/1 blue Faerie with flying"), N: 1})),
		},
	})
}
