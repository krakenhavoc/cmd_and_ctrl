package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Derevi, Empyrial Tactician — Legendary Creature — Bird Wizard {G}{W}{U},
// 2/3: "Flying. When Derevi enters and whenever a creature you control deals
// combat damage to a player, you may tap or untap target permanent.
// {1}{G}{W}{U}: Put Derevi onto the battlefield from the command zone."
//
// The trigger is one ability with two causes. The activated ability
// functions only from the command zone (CR 113.6, as commander ninjutsu
// does): it is not a cast, so there is no commander tax.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "afa49a09-146f-4439-850e-dd1938c93cef",
		Name:            "Derevi, Empyrial Tactician",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.OnAny([]game.EventKind{game.EventETB, game.EventDealDamage}, effects.Cmdr2S52DereviTapOrUntapCause,
					"Derevi — you may tap or untap target permanent",
					effects.B38TapOrUntapTarget("Derevi, Empyrial Tactician")),
				effects.TargetPermanent("target permanent")),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{G}{W}{U}: Put Derevi onto the battlefield from the command zone.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{1}{G}{W}{U}"),
			Zones:   []game.ZoneKind{game.ZoneCommand},
			Effect:  effects.Cmdr2S52PutThisFromCommandZone,
		}},
	})
}
