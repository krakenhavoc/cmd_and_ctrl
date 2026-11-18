package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Circle of Loyalty — Legendary Artifact {4}{W}{W}:
//
//	"Affinity for Knights
//	 Creatures you control get +1/+1.
//	 Whenever you cast a legendary spell, create a 2/2 white Knight
//	 creature token with vigilance.
//	 {3}{W}, {T}: Create a 2/2 white Knight creature token with
//	 vigilance."
//
// Affinity is the shared self cost modifier (Allies at Last), the anthem
// is Amazing Alliance's layer 7c grant over creatures you control, the
// cast trigger is WheneverYouCast over a legendary spell, and the
// activation makes Aryel's Knight token.
//
// No simplification.
func init() {
	knight := func() Applier {
		return CreateToken{Template: TokenCard("2/2 white Knight with vigilance"), N: 1}
	}
	Register(Spec{
		OracleID:     "fd1b4523-2399-4771-89bf-1d65f5d67056",
		Name:         "The Circle of Loyalty",
		Completeness: CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			AffinityFor("Affinity for Knights", HasSubtype("Knight")),
		},
		Static: []game.StaticAbility{{
			Layer:    game.Layer7PT,
			Label:    "Creatures you control get +1/+1.",
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power++
				c.Toughness++
			},
		}},
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Legendary(), "The Circle of Loyalty — create a 2/2 white Knight creature token with vigilance",
				Do(knight())),
		},
		Activated: []ActivatedAbility{{
			Label:   "{3}{W}, {T}: Create a 2/2 white Knight creature token with vigilance.",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Cost:    Plus(ManaCost("{3}{W}"), TapCost()),
			Effect:  Do(knight()),
		}},
	})
}
