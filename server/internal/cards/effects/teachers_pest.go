package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Teacher's Pest — Creature — Skeleton Pest {B}{G}, 1/1 (Secrets of
// Strixhaven, #4456):
//
//	"Menace
//	 Whenever this creature attacks, you gain 1 life.
//	 {B}{G}: Return this card from your graveyard to the battlefield
//	 tapped."
//
// Menace is a printed keyword; the attack trigger is the plain "whenever
// this creature attacks" (the card, unlike the Pest token, is not a token
// and has its own text). The graveyard ability is Reassembling Skeleton's
// (CR 113.6, ActivatedAbilityShape.Zones): a CR 602 activation from the
// graveyard that returns the card tapped through the entry replacement,
// doing nothing when the card has left the graveyard by resolution.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "12397dbf-99a9-40c3-aa7c-76d09baad008",
		Name:            "Teacher's Pest",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Teacher's Pest — you gain 1 life", Do(GainLife{Amount: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{B}{G}: Return this card from your graveyard to the battlefield tapped.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    ManaCost("{B}{G}"),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Effect:  returnThisFromGraveyardTapped,
		}},
	})
}
