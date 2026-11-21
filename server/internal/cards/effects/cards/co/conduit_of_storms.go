package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Conduit of Storms // Conduit of Emrakul — {2}{R} Creature — Werewolf
// Horror 2/3 // Creature — Eldrazi Werewolf 5/4 (Innistrad Remastered,
// S82 GAP, #5948):
//
//	Front: "Whenever this creature attacks, add {R} at the beginning of
//	        your next main phase this turn.
//	        {3}{R}{R}: Transform this creature."
//	Back:  "Whenever this creature attacks, add {C}{C} at the beginning
//	        of your next main phase this turn."
//
// The attack trigger schedules a delayed trigger (CR 603.7) for the next
// main phase still to begin this turn (the postcombat main, as the
// attack is declared in combat), Pygmy Hippo's scheduling; a turn with no
// main phase left adds nothing. The transform is the permanent turning
// over in place (Herbology Instructor's shape, at instant speed).
//
// No simplification.
func init() {
	const oracle = "1c4cf7c9-9dc5-49da-8d43-1ecd41c74ad3"
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Conduit of Storms",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Conduit of Storms — add {R} at the beginning of your next main phase this turn",
				func(g *game.Game, item *game.StackItem) error {
					return effects.ScheduleManaAtYourNextMainPhase(g, item, "Conduit of Storms — add {R}", conduitOfStormsManaBody)
				}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{R}{R}: Transform this creature.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{3}{R}{R}"),
			Effect:  effects.Do(effects.TransformThis{}),
		}},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Conduit of Emrakul",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Conduit of Emrakul — add {C}{C} at the beginning of your next main phase this turn",
				func(g *game.Game, item *game.StackItem) error {
					return effects.ScheduleManaAtYourNextMainPhase(g, item, "Conduit of Emrakul — add {C}{C}", conduitOfEmrakulManaBody)
				}),
		},
	})
}

// The two delayed bodies: the mana a Conduit's attack promised. On-disk
// keys: never renamed, never reused.
var (
	conduitOfStormsManaBody = game.DelayedBody("conduit-of-storms/add-red",
		func(g *game.Game, item *game.StackItem, _ game.EffectParams) error {
			return effects.AddMana{Produced: "{R}"}.Apply(effects.NewContext(g, item))
		})
	conduitOfEmrakulManaBody = game.DelayedBody("conduit-of-emrakul/add-colorless",
		func(g *game.Game, item *game.StackItem, _ game.EffectParams) error {
			return effects.AddMana{Produced: "{C}{C}"}.Apply(effects.NewContext(g, item))
		})
)
