package cardslu

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Martyr's Bond — Enchantment {4}{W}{W} (Commander 2011, #5574):
//
//	"Whenever this enchantment or another nonland permanent you control is put
//	 into a graveyard from the battlefield, each opponent sacrifices a permanent
//	 of their choice that shares a card type with it."
//
// One trigger covers both halves (the Bond is itself a nonland permanent its
// controller controls): any card of the Bond's controller going from the
// battlefield to a graveyard that was not a land as it last stood (CR 603.10a,
// the event's last-known types and controller, never the card in the
// graveyard). Each opponent then chooses, in APNAP order, one permanent they
// control that shares at least one of THOSE card types with it, and sacrifices
// it; an opponent with none sacrifices nothing (CR 701.21a). A creature
// artifact that died asks for a creature or an artifact.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7de8ffa5-f6e5-4987-b3f8-6d6fc83f8fa8",
		Name:         "Martyr's Bond",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, martyrsBondWatches,
				"Martyr's Bond — each opponent sacrifices a permanent that shares a card type with it",
				func(g *game.Game, item *game.StackItem) error {
					var types []string
					for _, t := range item.Trigger.Event.LastKnownTypes {
						types = append(types, strings.ToLower(t))
					}
					return effects.EachPlayerSacrifices{
						ExceptController: true,
						Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
							for _, t := range types {
								if c.HasCardType(t) {
									return true
								}
							}
							return false
						},
						Label: "a permanent that shares a card type with it",
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

// martyrsBondWatches is "this enchantment or another nonland permanent you
// control is put into a graveyard from the battlefield".
func martyrsBondWatches(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventLTB || ev.NewZone != game.ZoneGraveyard {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || effects.LeftUnderControlOf(ev, c) != source.Controller || effects.LeftAsType(ev, c, "land") {
		return false
	}
	return len(ev.LastKnownTypes) > 0 || !c.HasCardType("land")
}
