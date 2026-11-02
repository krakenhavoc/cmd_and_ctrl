package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Silk, Web Weaver — Legendary Creature — Spider Human Hero {2}{G}{W}, 3/5
// (Marvel's Spider-Man):
//
//	"Web-slinging {1}{G}{W} (You may cast this spell for {1}{G}{W} if you
//	 also return a tapped creature you control to its owner's hand.)
//	 Whenever you cast a creature spell, create a 1/1 green and white
//	 Human Citizen creature token.
//	 {3}{G}{W}: Creatures you control get +2/+2 and gain vigilance until
//	 end of turn."
//
// Web-slinging is Spider-Sense's alternative cost (the returned creature is a
// cost, chosen as the spell is cast). The trigger watches creature spells
// the controller casts; Silk herself is on the stack when she is cast, so
// she never triggers off her own casting. The activated ability affects the
// creatures you control as it resolves (CR 611.2c): both halves read that
// same set.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "50477dc8-9a2a-41eb-984d-c9e1e78e08a2",
		Name:         "Silk, Web Weaver",
		Completeness: CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{{
			Key:          "web_slinging",
			Label:        "Web-slinging {1}{G}{W}",
			ManaCost:     "{1}{G}{W}",
			ReturnToHand: PermanentYouControl("a tapped creature you control", Creature(), b751Tapped()),
			PayLabel:     "a tapped creature you control",
		}},
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Creature(), "Silk, Web Weaver — create a 1/1 green and white Human Citizen creature token",
				Do(CreateToken{Template: TokenCard("1/1 green and white Human Citizen"), N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{3}{G}{W}: Creatures you control get +2/+2 and gain vigilance until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerCombatGrant},
			Cost:    ManaCost("{3}{G}{W}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				yours := And(Creature(), YouControl())
				if err := (BoostUntilEOT{Match: yours, Power: 2, Toughness: 2, Label: "Silk, Web Weaver — +2/+2"}).Apply(ctx); err != nil {
					return err
				}
				return GrantKeywordUntilEOT{Match: yours, Keywords: []string{"vigilance"}, Label: "Silk, Web Weaver — vigilance"}.Apply(ctx)
			},
		}},
	})
}
