package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spit Flame — Instant {2}{R} (Tarkir: Dragonstorm Commander):
//
//	"Spit Flame deals 4 damage to target creature.
//	 Whenever a Dragon you control enters, you may pay {R}. If you do,
//	 return this card from your graveyard to your hand."
//
// The spell is a plain 4 damage. The second ability works from the
// graveyard (CR 113.6, InGraveyard) on a Dragon entering under your
// control; "you may pay {R}. If you do" is a MayPay asked as the trigger
// resolves, and the return is Furious Forebear's keyed branch: it needs
// the card to be in the graveyard still, as the same object (CR 400.7).
//
// No simplification.
func init() {
	t := On(game.EventETB, spitFlameDragonEntered,
		"Spit Flame — you may pay {R}; if you do, return this card from your graveyard to your hand",
		func(g *game.Game, item *game.StackItem) error {
			return MayPay{
				Chooser:  item.Controller,
				Cost:     "{R}",
				Question: "Spit Flame — pay {R} to return it from your graveyard to your hand?",
				Then:     furiousForebearPaid,
			}.Apply(NewContext(g, item))
		})
	Register(Spec{
		OracleID:     "f63278e0-3c67-421d-87ce-67725e5e74df",
		Name:         "Spit Flame",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			for _, tg := range ctx.LegalTargets() {
				if tg.Kind == game.TargetCard {
					return DealDamage{Source: ctx.Source(), Target: tg.ID, Amount: 4}.Apply(ctx)
				}
			}
			return nil
		},
		Triggered: []game.TriggeredAbility{InGraveyard(t)},
	})
}

// spitFlameDragonEntered is "a Dragon you control enters": a creature
// permanent under the card's owner's control with the Dragon type as it
// enters (read through the layers; Dragon is a word of the card's text).
func spitFlameDragonEntered(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := enteredUnderYourControl(ev, source, g, false)
	return ok && c.IsCreature() && c.HasSubtype(source.TextSubtype("Dragon"))
}
