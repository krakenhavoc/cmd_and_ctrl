package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Thorn Lieutenant — Creature — Elf Warrior {1}{G}, 2/3 (Core Set 2019):
//
//	"Whenever this creature becomes the target of a spell or ability an
//	 opponent controls, create a 1/1 green Elf Warrior creature token.
//	 {5}{G}: This creature gets +4/+4 until end of turn."
//
// The trigger watches EventBecomesTarget for this creature, fired once per
// target slot, from a spell or ability whose controller is not the Lieutenant's
// controller. The pump is a layer 7c +4/+4 on the Lieutenant if it is still
// the permanent the ability came from.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2b761ce6-4fb1-4f22-b51c-d9f25f62b39a",
		Name:         "Thorn Lieutenant",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBecomesTarget, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID && ev.Actor != uuid.Nil && ev.Actor != source.Controller
			}, "Thorn Lieutenant — create a 1/1 green Elf Warrior creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 green Elf Warrior"), N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{5}{G}: This creature gets +4/+4 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 4, Toughness: 4}},
			Cost:    effects.ManaCost("{5}{G}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				return effects.BoostUntilEOT{Target: item.SourceCardID, Power: 4, Toughness: 4,
					Label: "Thorn Lieutenant — +4/+4 until end of turn"}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
