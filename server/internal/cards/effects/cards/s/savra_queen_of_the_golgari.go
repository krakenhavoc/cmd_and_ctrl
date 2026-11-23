package cardss

import (
	"errors"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Savra, Queen of the Golgari — Legendary Creature — Elf Shaman {2}{B}{G},
// 2/2 (Ravnica Remastered, #5948):
//
//	"Whenever you sacrifice a black creature, you may pay 2 life. If you do,
//	 each other player sacrifices a creature of their choice.
//	 Whenever you sacrifice a green creature, you may gain 2 life."
//
// Both triggers watch the sacrifice event, which is emitted before the
// permanent leaves, so its colours are the ones it has as it is sacrificed
// (and Savra herself, black and green, fires both). The life is paid as the
// "if you do" of the may, so a player who can't pay it (CR 119.4) skips the
// edict; each other player then picks their own creature (EachPlayerSacrifices
// ExceptController). A creature that is both colours fires both triggers.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "641db975-cf94-44c6-a391-f540a952e0d8",
		Name:         "Savra, Queen of the Golgari",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerEdict(game.TriggeredAbility{
				Watches:   []game.EventKind{game.EventSacrifice},
				AppliesTo: savraYouSacrificedACreatureOfColor("B"),
				Key:       "Savra, Queen of the Golgari — you may pay 2 life; if you do, each other player sacrifices a creature",
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.MayChoice{
						Question: "Savra, Queen of the Golgari — pay 2 life so each other player sacrifices a creature?",
						YesLabel: "Pay 2 life",
						NoLabel:  "Don't pay",
						LifeCost: 2,
						Then:     savraPayLifeEdict,
					}.Apply(effects.NewContext(g, item))
				},
			}),
			{
				Watches:   []game.EventKind{game.EventSacrifice},
				AppliesTo: savraYouSacrificedACreatureOfColor("G"),
				Key:       "Savra, Queen of the Golgari — you may gain 2 life",
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.MayChoice{
						Question: "Savra, Queen of the Golgari — gain 2 life?",
						Then:     effects.MayGainLife,
						Word:     "2",
					}.Apply(effects.NewContext(g, item))
				},
			},
		},
	})
}

func savraYouSacrificedACreatureOfColor(color string) func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		if !effects.B23YouSacrificedAPermanent(ev, source, game.Characteristic{}, g) {
			return false
		}
		c, ok := g.LookupCardForEffect(ev.CardID)
		return ok && c.IsCreature() && c.HasColor(source.TextColor(color))
	}
}

var savraPayLifeEdict = effects.MayYes("confirm/savra-pay-life-edict", func(ctx *effects.Context, r game.ConfirmAnswer) error {
	if r.LifeCost <= 0 || !effects.CanPayLifeNow(ctx, r.LifeCost) {
		return nil
	}
	if err := ctx.Game.PayLifeForEffect(ctx.Source(), ctx.Controller(), r.LifeCost); err != nil {
		if errors.Is(err, game.ErrInvalidParam) {
			return nil // a payment that can't be made is not made (CR 119.8)
		}
		return err
	}
	return effects.EachPlayerSacrifices{ExceptController: true, Match: effects.Creature(), Label: "a creature"}.Apply(ctx)
})
