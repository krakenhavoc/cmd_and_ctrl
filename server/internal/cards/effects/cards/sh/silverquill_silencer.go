package cardssh

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Silverquill Silencer — Creature — Human Cleric {W}{B}, 3/2 (Strixhaven:
// School of Mages, #5502):
//
//	"As this creature enters, choose a nonland card name.
//	 Whenever an opponent casts a spell with the chosen name, they lose 3 life
//	 and you draw a card."
//
// Kai Winn's name choice and cast trigger: the name is chosen as it enters
// (CR 614.12), per instance, and kept on the permanent. The trigger fires for
// a spell an opponent of the Silencer's controller casts whose name is the
// chosen one (a split card cast as one half has only that half's name); the
// caster loses the life and the Silencer's controller draws. A name that is
// never chosen (a restored game with none) matches nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "df9ab83c-850a-4be7-89cd-5122353081f6",
		Name:         "Silverquill Silencer",
		Completeness: effects.CompletenessFull,
		AsEnters:     effects.ChooseCardNameAsEnters("Silverquill Silencer — choose a nonland card name"),
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Actor == uuid.Nil || ev.Actor == source.Controller {
					return false
				}
				named := strings.TrimSpace(source.ChosenName)
				if named == "" {
					return false
				}
				spell, ok := g.LookupCardForEffect(ev.CardID)
				return ok && game.SpellOrLandHasName(spell, named)
			}, "Silverquill Silencer — they lose 3 life and you draw a card",
				func(g *game.Game, item *game.StackItem) error {
					if err := g.ChangePlayerLifeForEffect(item.SourceCardID, effects.TriggeringActor(item), -3); err != nil {
						return err
					}
					return effects.DrawCards{Player: item.Controller, N: 1}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
