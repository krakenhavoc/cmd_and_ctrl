package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// whenACartoucheYouControlEntersReturnThis is the Trials' second ability:
// "When a Cartouche you control enters, return this enchantment to its
// owner's hand." The Trial must still be the permanent that was on the
// stack (CR 400.7), which stdB9ReturnThisToHand checks.
func whenACartoucheYouControlEntersReturnThis(name string) game.TriggeredAbility {
	return On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		c, ok := enteredUnderYourControl(ev, source, g, false)
		return ok && c.HasSubtype("Cartouche")
	}, name+" — return this enchantment to its owner's hand", Do(Instruction(stdB9ReturnThisToHand)))
}
