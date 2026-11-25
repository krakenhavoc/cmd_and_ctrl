package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Astral Drift — {2}{W} Enchantment (Commander 2020):
//
//	"Whenever you cycle this card or cycle another card while this
//	 enchantment is on the battlefield, you may exile target creature. If you
//	 do, return that card to the battlefield under its owner's control at the
//	 beginning of the next end step.
//	 Cycling {2}{W}"
//
// One printed ability, one registered trigger watching two zones: from the
// graveyard for the card's own cycling (CR 702.29c; "you" is its owner there),
// and from the battlefield for any other card the controller cycles. Which of
// the two applies is read from where the source is as the event happens, so
// the enchantment on the battlefield never answers to itself and a copy in a
// graveyard never answers to another card's cycling. The exile and the delayed
// return are Flickerwisp's (ExileTargetsThenScheduleReturn): the card comes
// back under its owner's control at the next end step (CR 400.7).
func init() {
	trig := effects.Targeting(
		effects.Optional(
			effects.On(game.EventCycle, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Kind != game.EventCycle || source == nil || ev.Actor != source.Controller {
					return false
				}
				z := g.FindCardZoneForEffect(source.InstanceID)
				if z == nil {
					return false
				}
				if ev.CardID == source.InstanceID {
					return z.Kind == game.ZoneGraveyard
				}
				return z.Kind == game.ZoneBattlefield
			}, "Astral Drift — you may exile target creature; return it at the beginning of the next end step",
				func(g *game.Game, item *game.StackItem) error {
					return effects.ExileTargetsThenScheduleReturn(effects.NewContext(g, item), "Astral Drift — return the exiled creature to the battlefield")
				}),
			"Astral Drift — exile target creature until the next end step?"),
		effects.TargetCreature("target creature"))
	trig.Zones = []game.ZoneKind{game.ZoneBattlefield, game.ZoneGraveyard}
	effects.Register(effects.Spec{
		OracleID:     "7aa5db50-5057-45fd-bce9-fb4a5e874263",
		Name:         "Astral Drift",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.Cycling("{2}{W}")},
		Triggered:    []game.TriggeredAbility{trig},
	})
}
