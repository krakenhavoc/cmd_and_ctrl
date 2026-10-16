package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Jirina, Dauntless General — Legendary Creature — Human Soldier
// {W}{B}, 2/2 (the Squall deck request, #3918):
//
//	"When Jirina enters, exile target player's graveyard.
//	 Sacrifice Jirina: Humans you control gain hexproof and
//	 indestructible until end of turn."
//
// The entry trigger is Bojuka Bog's, on a creature: the player is
// chosen as the trigger goes on the stack and the graveyard exiled is
// the one that player has as it resolves.
//
// The sacrifice is Ainok Strike Leader's shape. "Humans you control" is
// the set as the ability resolves (CR 611.2c), read through effective
// subtypes, so a changeling counts and a Human that enters later in the
// turn does not. It names Humans, not Human creatures, so a Kindred
// Human permanent counts too. Jirina is a Human too, but it has been sacrificed by
// then.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fda76032-7e6f-416f-8874-2a0642837522",
		Name:         "Jirina, Dauntless General",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Jirina, Dauntless General — exile target player's graveyard", exileTargetPlayersGraveyard),
				TargetPlayer("target player")),
		},
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice Jirina: Humans you control gain hexproof and indestructible until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive, Saves: &game.Saves{Matches: game.SweepCreatures, Partial: true}},
			Cost:    game.AbilityCost{SacrificeSelf: true},
			Effect: Do(GrantKeywordUntilEOT{
				Match:    And(HasSubtype("Human"), YouControl()),
				Keywords: []string{"hexproof", "indestructible"},
				Label:    "Jirina, Dauntless General — hexproof and indestructible",
			}),
		}},
	})
}
