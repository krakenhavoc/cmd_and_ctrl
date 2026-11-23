package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sutina, Speaker of the Tajuru — Legendary Creature — Elf Scout {2}{G}, 2/2
// (Foundations Jumpstart, #5929):
//
//	"When Sutina enters, search your library for a basic land card, put it
//	 onto the battlefield tapped, then shuffle.
//	 Whenever Sutina attacks, you may return a land you control to its owner's
//	 hand. When you do, put a +1/+1 counter on target creature."
//
// The entry search is Kodama of the East Tree's tapped basic fetch. The attack
// trigger is Kefnet the Mindful's land pick (none or one, so the "may" is the
// empty pick), made at resolution; once the land is in its owner's hand the
// reflexive trigger (CR 603.12) goes on the stack above it with its target
// chosen as it goes there, so no pick, or a land that could not be returned,
// makes none. Any land you control may be returned, tapped or not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "91f410c5-4313-462f-8d21-da63988d919f",
		Name:         "Sutina, Speaker of the Tajuru",
		Completeness: effects.CompletenessFull,
		// The entry search: a basic land onto the battlefield tapped (ADR 0126 §6).
		Purpose: game.Purpose{Lands: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Sutina — search your library for a basic land card, put it onto the battlefield tapped, then shuffle",
				func(g *game.Game, item *game.StackItem) error {
					return effects.B07SearchBasicOntoBattlefield(g, item, item.Controller, true, false,
						"Sutina — a basic land, onto the battlefield tapped")
				}),
			effects.WheneverThisAttacks("Sutina — you may return a land you control to its owner's hand; when you do, put a +1/+1 counter on target creature",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayReturnALandYouControl("Sutina — you may return a land you control to its owner's hand",
						b52SutinaLandReturned).Apply(effects.NewContext(g, item))
				}),
		},
	})
}

// b52SutinaLandReturned is the chosen land returning to its owner's hand and
// then the reflexive counter, keyed (#4290). An on-disk key.
var b52SutinaLandReturned = effects.RegisterPermanentsPick("prompt-run/bns-b52-sutina-return-land", effects.PermanentsDef{
	Then: func(ctx *effects.Context, picked game.PromptedPicks) error {
		for _, id := range picked.Cards() {
			if !effects.OnBattlefield(ctx.Game, id) {
				continue
			}
			if err := (effects.BounceToHand{Target: id}).Apply(ctx); err != nil {
				return err
			}
			return effects.ReflexiveTrigger{
				Label: "Sutina — put a +1/+1 counter on target creature",
				Body:  effects.AntManCounterBody,
			}.Apply(ctx)
		}
		return nil
	},
})
