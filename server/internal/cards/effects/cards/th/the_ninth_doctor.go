package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Ninth Doctor — Legendary Creature — Time Lord Doctor {1}{U}{R}, 2/4:
//
//	"Haste
//	 Into the TARDIS — Whenever The Ninth Doctor becomes untapped during
//	 your untap step, you get an additional upkeep step after this step."
//
// The trigger watches The Ninth Doctor's own untap (EventUntapCard), in
// the untap step of its controller's turn: an untap effect at another
// time, or another player's untap step, does not trigger it. Nothing goes
// on the stack in the untap step (CR 502.4), so the trigger is put on the
// stack in the upkeep and resolves there; "this step" is then that upkeep,
// and the additional upkeep step comes directly after it, before the draw
// step (CR 500.9). It is an ordinary upkeep: "at the beginning of your
// upkeep" triggers again and its controller gets priority. "You get" adds
// nothing on another player's turn (CR 500.10a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c99f59c0-5cce-41a4-908f-d34e2c697419",
		Name:            "The Ninth Doctor",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventUntapCard, theNinthDoctorUntappedInYourUntapStep,
				"Into the TARDIS — you get an additional upkeep step after this step", theNinthDoctorExtraUpkeep),
		},
	})
}

// theNinthDoctorUntappedInYourUntapStep is "The Ninth Doctor becomes
// untapped during your untap step".
func theNinthDoctorUntappedInYourUntapStep(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return ev.CardID == source.InstanceID && g.Turn.Step == game.StepUntap && effects.IsYourTurn(g, source.Controller)
}

// theNinthDoctorExtraUpkeep is the trigger's body.
func theNinthDoctorExtraUpkeep(g *game.Game, item *game.StackItem) error {
	if !effects.IsYourTurn(g, item.Controller) {
		return nil
	}
	return effects.AddStepAfterThisStep{Step: game.StepUpkeep}.Apply(effects.NewContext(g, item))
}
