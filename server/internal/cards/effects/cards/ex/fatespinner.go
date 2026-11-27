package cardsex

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fatespinner — Creature — Human Wizard {1}{U}{U}, 1/2 (Mirrodin, S82 OLD,
// #5606):
//
//	"At the beginning of each opponent's upkeep, that player chooses draw
//	 step, main phase, or combat phase. The player skips each instance of
//	 the chosen step or phase this turn."
//
// The trigger reads the upkeep's player off its event (AtEachOpponentsUpkeep).
// That player, not the Fatespinner's controller, picks one of the three as
// the trigger resolves, and the skip is #6226's record
// (SkipThisTurnForEffect, CR 500.11): every instance of the chosen step or
// phase this turn, an added one included, is passed over, and one that has
// already begun is not (CR 614.10). The answer is an option pick, so a
// table waiting on it is a restore point.
//
// No simplification.
func init() {
	const label = "Fatespinner — that player chooses draw step, main phase, or combat phase and skips each instance of it this turn"
	effects.Register(effects.Spec{
		OracleID:     "676935c8-22e0-427f-9566-16778cea0873",
		Name:         "Fatespinner",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtEachOpponentsUpkeep(label, func(g *game.Game, item *game.StackItem) error {
				if item.Trigger == nil {
					return nil
				}
				player := item.Trigger.Event.Actor
				if player == uuid.Nil || g.PlayerByIDForEffect(player) == nil {
					return nil
				}
				g.QueueOptionPickForEffect(game.OptionPickPrompt{
					Chooser:  player,
					Source:   item.SourceCardID,
					Question: "Fatespinner — choose draw step, main phase, or combat phase to skip this turn",
					Options: []game.ChoiceOption{
						{Label: "Draw step"},
						{Label: "Main phase"},
						{Label: "Combat phase"},
					},
					ThenKey: fatespinnerChosen,
					Carry:   []uuid.UUID{player},
				})
				return nil
			}),
		},
	})
}

// fatespinnerChosen records the skip the upkeep's player chose. An on-disk
// key: never renamed, never reused.
var fatespinnerChosen = game.RegisterOptionPickThen("option-pick/s82-old-b602-fatespinner-skip", func(g *game.Game, r game.OptionPicked) error {
	if r.Option == nil || len(r.Carry) == 0 {
		return nil
	}
	part := ""
	switch r.Option.Label {
	case "Draw step":
		part = game.SkipThisTurnDraw
	case "Main phase":
		part = game.SkipThisTurnMain
	case "Combat phase":
		part = game.SkipThisTurnCombat
	default:
		return nil
	}
	g.SkipThisTurnForEffect(r.Carry[0], r.Source, part, "Fatespinner — "+r.Option.Label+" skipped this turn")
	return nil
})
