package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Derevi, Empyrial Tactician — Legendary Creature — Bird Wizard {G}{W}{U},
// 2/3: "Flying. When Derevi enters and whenever a creature you control deals
// combat damage to a player, you may tap or untap target permanent.
// {1}{G}{W}{U}: Put Derevi onto the battlefield from the command zone."
//
// The trigger is one ability with two causes. The activated ability
// functions only from the command zone (CR 113.6, as commander ninjutsu
// does): it is not a cast, so there is no commander tax.
func init() {
	Register(Spec{
		OracleID:        "afa49a09-146f-4439-850e-dd1938c93cef",
		Name:            "Derevi, Empyrial Tactician",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				OnAny([]game.EventKind{game.EventETB, game.EventDealDamage}, cmdr2S52DereviTapOrUntapCause,
					"Derevi — you may tap or untap target permanent",
					b38TapOrUntapTarget("Derevi, Empyrial Tactician")),
				TargetPermanent("target permanent")),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}{G}{W}{U}: Put Derevi onto the battlefield from the command zone.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    ManaCost("{1}{G}{W}{U}"),
			Zones:   []game.ZoneKind{game.ZoneCommand},
			Effect:  cmdr2S52PutThisFromCommandZone,
		}},
	})
}
