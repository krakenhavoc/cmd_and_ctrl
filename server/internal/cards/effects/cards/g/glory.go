package cardsg

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Glory — Creature — Incarnation {3}{W}{W}, 3/3 (Dominaria Remastered,
// #5948):
//
//	"Flying
//	 {2}{W}: Choose a color. Creatures you control gain protection from the
//	 chosen color until end of turn. Activate only if this card is in your
//	 graveyard."
//
// An ability that works from the graveyard (Loathsome Troll's shape, the
// ability's Zones). The creatures it affects are the ones its controller
// controls as it resolves (CR 611.2c), fixed before the colour is named,
// and the colour is chosen by the controller; the protection is Akroma's
// Blessing's keyed tail. With Glory in the graveyard it affects only the
// creatures then on the battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3853ef6a-6771-4708-8ed5-d667f646de5a",
		Name:            "Glory",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{W}: Choose a color. Creatures you control gain protection from the chosen color until end of turn. Activate only if this card is in your graveyard.",
			Purpose: game.Purpose{Answers: game.AnswerProtect},
			Cost:    effects.ManaCost("{2}{W}"),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				var ids []uuid.UUID
				for _, c := range g.BattlefieldCardsForEffect() {
					if c.IsCreature() && c.Controller == item.Controller {
						ids = append(ids, c.InstanceID)
					}
				}
				if len(ids) == 0 {
					return nil
				}
				then := ctx.ItemTail(effects.RadianceProtectionFromChosenColor)
				then.Carry = ids
				then.Words = []string{"Glory"}
				effects.ChooseColorTail(game.ColorForProtection, g, item.Controller, ctx.Source(), "Glory — choose a color", then)
				return nil
			},
		}},
	})
}
