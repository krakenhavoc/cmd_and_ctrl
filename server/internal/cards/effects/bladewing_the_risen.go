package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bladewing the Risen — Legendary Creature — Zombie Dragon {3}{B}{B}{R}{R},
// 4/4 (Scourge):
//
//	"Flying
//	 When Bladewing enters, you may return target Dragon permanent card
//	 from your graveyard to the battlefield.
//	 {B}{R}: Dragon creatures get +1/+1 until end of turn."
//
// The enters trigger is Renegade Rallier's reanimation over a Dragon
// permanent card in your graveyard, as a "you may" (Aphetto Vulture's
// shape): the target is chosen as the trigger goes on the stack, so with no
// Dragon card there it is removed (CR 603.3d), and the "may" is asked as it
// resolves (CR 603.5). The pump is Magnify's locked-in set over every Dragon
// creature on the battlefield, any player's, fixed as it resolves (CR
// 611.2c), so a Dragon that arrives later is not pumped.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c22cc70a-9533-4bfc-b3ce-b2d43a73f0c1",
		Name:            "Bladewing the Risen",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				Optional(
					WhenThisEnters("Bladewing the Risen — return target Dragon permanent card from your graveyard to the battlefield",
						sheoldredReanimate),
					"Bladewing the Risen — return target Dragon permanent card from your graveyard to the battlefield?"),
				TargetCardInGraveyard("target Dragon permanent card from your graveyard", YouOwn(), Permanent(), HasSubtype("Dragon")),
			),
		},
		Activated: []ActivatedAbility{{
			Label:   "{B}{R}: Dragon creatures get +1/+1 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{B}{R}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return BoostUntilEOT{
					Match:     OfCreatureType(ctx.TextSubtype("Dragon")),
					Power:     1,
					Toughness: 1,
					Label:     "Bladewing the Risen — Dragon creatures get +1/+1",
				}.Apply(ctx)
			},
		}},
	})
}
