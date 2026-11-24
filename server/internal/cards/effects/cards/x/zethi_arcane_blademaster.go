package cardsx

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Zethi, Arcane Blademaster — Legendary Creature — Human Soldier {1}{W}{U},
// 3/3 (Secret Lair Drop):
//
//	"Multikicker {W/U}
//	 When Zethi, Arcane Blademaster enters, exile up to X target instant cards
//	 from your graveyard, where X is the number of times Zethi was kicked. Put
//	 a kick counter on each of them.
//	 Whenever Zethi attacks, copy each exiled card you own with a kick counter
//	 on it. You may cast the copies."
//
// Batroc the Leaper's up-to-X targets (TargetsFrom reads the kicks as the
// trigger goes on the stack) and Roving Actuator's exile-then-copy: the exile
// is keyed, and its tail stamps the kick counter on the cards that really
// reached exile. The attack trigger copies every exiled card its controller
// owns that still holds a kick counter, whoever exiled it, and offers the
// casts for their printed mana costs (CopyAndCast, CR 707.12); the originals
// stay in exile with their counters, so each attack copies them again.
func init() {
	enters := effects.WhenThisEnters("Zethi, Arcane Blademaster — exile up to X target instant cards from your graveyard and put a kick counter on each",
		func(g *game.Game, item *game.StackItem) error {
			ids := effects.LegalTargetIDs(effects.NewContext(g, item))
			if len(ids) == 0 {
				return nil
			}
			return g.ExileCardsKeyedForEffect(ids, effects.NewContext(g, item).ItemTail(zethiKickCounters))
		})
	enters.TargetsFrom = func(_ game.TriggerContext, source *game.Card, _ *game.Game) *game.TargetSpec {
		x := game.CardKickedTimes(*source)
		if x <= 0 {
			return nil
		}
		return effects.TargetCardInGraveyard("up to X target instant cards from your graveyard",
			effects.YouOwn(), effects.Instant()).WithCount(0, x)
	}
	effects.Register(effects.Spec{
		OracleID:      "ac136d46-32b7-4222-a840-b8be011c1872",
		Name:          "Zethi, Arcane Blademaster",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Multikicker("{W/U}", 20)},
		Triggered: []game.TriggeredAbility{
			enters,
			effects.WheneverThisAttacks("Zethi, Arcane Blademaster — copy each exiled card you own with a kick counter on it; you may cast the copies",
				func(g *game.Game, item *game.StackItem) error {
					var ids []uuid.UUID
					for _, c := range g.Exile.Cards {
						if c.Owner == item.Controller && c.Counters[zethiKickCounter] > 0 {
							ids = append(ids, c.InstanceID)
						}
					}
					if len(ids) == 0 {
						return nil
					}
					return effects.CopyAndCast{Cards: ids}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

// zethiKickCounter is the kick counter Zethi stamps on the cards it exiles.
const zethiKickCounter = "kick"

var zethiKickCounters = effects.CardTail("zethi/kick-counters", func(ctx *effects.Context, o game.TailOutcome) error {
	for _, id := range o.IDs {
		if err := ctx.Game.AddCounterForEffect(id, zethiKickCounter, 1); err != nil {
			return err
		}
	}
	return nil
})
