package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nightpack Ambusher — Creature — Wolf {2}{G}{G}, 4/4:
//
//	"Flash
//	 Other Wolves and Werewolves you control get +1/+1.
//	 At the beginning of your end step, if you didn't cast a spell this
//	 turn, create a 2/2 green Wolf creature token."
//
// The end-step trigger is an intervening "if" (CR 603.4), Edgar,
// Moonlit Sovereign's shape: the cast tally is read when the step
// begins and again as the trigger resolves, so a spell cast in response
// (flash) turns it off. The tally counts the Ambusher's own cast, so one
// flashed in this turn makes no Wolf. The anthem is a TribalAnthem over
// OTHER Wolves and Werewolves you control. Flash is PrintedKeywords.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4d67357c-5939-45c8-a656-aaf76728a5a6",
		Name:            "Nightpack Ambusher",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Static: []game.StaticAbility{
			effects.TribalAnthem(effects.TribeFilter{Tribes: []string{"Wolf", "Werewolf"}, Others: true, YoursOnly: true}, 1, 1),
		},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginEndStep, effects.AllOf(effects.ByYou, effects.RfCreatureBYouCastNoSpellThisTurn),
				"Nightpack Ambusher — create a 2/2 green Wolf token",
				effects.Do(effects.If{
					Cond: effects.CondNoSpellsCast{},
					Then: []effects.Applier{effects.CreateToken{Template: effects.TokenCard("2/2 green Wolf"), N: 1}},
				})),
		},
	})
}
