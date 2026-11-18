package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sigarda's Splendor — Enchantment {2}{W}{W}:
//
//	"As this enchantment enters, note your life total.
//	 At the beginning of your upkeep, draw a card if your life total is
//	 greater than or equal to the last noted life total for this
//	 enchantment. Then note your life total.
//	 Whenever you cast a white spell, you gain 1 life."
//
// The note is the permanent's Card.ChosenNumber, written as it enters
// (CR 614.12a) and rewritten at each upkeep through
// Game.SetNotedNumberForEffect, which unlike the chosen-number setter does
// not clamp a negative life total at zero. The upkeep compares against the
// note as it last was and then notes the life total after any draw, as
// printed. An enchantment that left before the trigger resolves is read
// from its last-known information (CR 608.2h) and notes nothing.
func init() {
	Register(Spec{
		OracleID:     "71a3d0ce-052b-41eb-a70f-0c716149a55f",
		Name:         "Sigarda's Splendor",
		Completeness: CompletenessFull,
		AsEnters: func(card *game.Card, ctx *Context) error {
			if p := ctx.Game.PlayerByIDForEffect(card.Controller); p != nil {
				ctx.Game.SetNotedNumberForEffect(card.InstanceID, p.Life)
			}
			return nil
		},
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Sigarda's Splendor — draw a card if your life total is at least the last noted life total, then note your life total", sigardasSplendorUpkeep),
			WheneverYouCast(OfColor("W"), "Sigarda's Splendor — you gain 1 life", Do(GainLife{Amount: 1})),
		},
	})
}

func sigardasSplendorUpkeep(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	info, ok := ctx.SourcePermanent()
	p := g.PlayerByIDForEffect(ctx.Controller())
	if !ok || p == nil {
		return nil
	}
	if p.Life >= info.ChosenNumber {
		if err := (DrawCards{Player: p.ID, N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	if !info.Left {
		g.SetNotedNumberForEffect(item.SourceCardID, p.Life)
	}
	return nil
}
