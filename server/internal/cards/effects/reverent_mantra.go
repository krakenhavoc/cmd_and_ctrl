package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Reverent Mantra — Instant {3}{W} (Mercadian Masques, #5619):
//
//	"You may exile a white card from your hand rather than pay this
//	 spell's mana cost.
//	 Choose a color. All creatures gain protection from the chosen color
//	 until end of turn."
//
// Cave-In's pitch with a white card (Pitch, no life): the card is exiled
// as a cost with the Mantra already on the stack (CR 601.2h), so the
// Mantra itself can't pay it. The body is Akroma's Blessing's over every
// creature on the battlefield, not only the caster's: the set is read as
// the spell resolves (CR 611.2c), so a creature that enters later is not
// protected. The colour is chosen then too, as the colour they are
// protected FROM (ColorForProtection), and the protection is Bathe in
// Light's tail over that set. With no creature on the battlefield
// nothing is asked.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b4f7b1eb-6cd4-40f9-b259-3c39291edf27",
		Name:         "Reverent Mantra",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerProtect},
		AlternativeCosts: []game.AlternativeCost{
			Pitch("Exile a white card from your hand", 0,
				CardInYourHand("a white card from your hand", OfColor("W")),
				"a white card from your hand"),
		},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			var ids []uuid.UUID
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				if c.IsCreature() {
					ids = append(ids, c.InstanceID)
				}
			}
			if len(ids) == 0 {
				return nil
			}
			then := ctx.ItemTail(radianceProtectionFromChosenColor)
			then.Carry = ids
			then.Words = []string{"Reverent Mantra"}
			ChooseColorTail(game.ColorForProtection, ctx.Game, item.Controller, ctx.Source(), "Reverent Mantra — choose a color", then)
			return nil
		},
	})
}
