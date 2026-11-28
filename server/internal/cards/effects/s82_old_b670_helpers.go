package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// OldB670WheneverThisDealsCombatDamageToACreatureUnlessItsControllerPays is the
// shared shape of Death Charmer, Plague Fiend and Soul Charmer: "Whenever this
// creature deals combat damage to a creature, <consequence> unless that
// creature's controller pays <cost>." The damaged creature and its controller are
// read as the damage is dealt (CR 510.2, CR 608.2h), so the question is still
// asked of that player when the creature died to the damage. The consequence is
// the registered `decline` (PayUnlessElse): it reads the payer as r.Chooser, the
// trigger's controller as ctx.Controller(), and r.Carry, which holds the damaged
// creature's ID only while that object is still on the battlefield.
func OldB670WheneverThisDealsCombatDamageToACreatureUnlessItsControllerPays(label, cost, question string, decline game.PayUnlessElse) game.TriggeredAbility {
	return game.TriggeredAbility{
		Watches:   []game.EventKind{game.EventDealDamage},
		AppliesTo: ThisDealtCombatDamageToACreature,
		Key:       label,
		Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
			item := game.NewTriggeredItem(source, label)
			item.Params.Object = DamagedCreatureRef(g, ev.Target)
			if c, ok := g.LookupCardForEffect(ev.Target); ok {
				item.Params.Player = c.Controller
			}
			return item
		},
		Effect: func(g *game.Game, item *game.StackItem) error {
			var carry []uuid.UUID
			ref := item.Params.Object
			if c, ok := g.LookupCardForEffect(ref.ID); ok && c.ObjectEpoch == ref.Epoch {
				if z := g.FindCardZoneForEffect(ref.ID); z != nil && z.Kind == game.ZoneBattlefield {
					carry = append(carry, ref.ID)
				}
			}
			return PayUnless{
				Chooser:  item.Params.Player,
				Cost:     cost,
				Question: question,
				Else:     decline,
				Carry:    carry,
			}.Apply(NewContext(g, item))
		},
	}
}

// OldB670AnOpponentControlsCreaturesMoreThanYou is Avatar of Might's condition,
// "an opponent controls at least four more creatures than you" for n = 4: one
// opponent against the caster, not the opponents added together, counted through
// the layers as the cost is determined (CR 601.2f).
func OldB670AnOpponentControlsCreaturesMoreThanYou(n int) CostPredicate {
	return func(q game.CostQuery) bool {
		if q.Game == nil {
			return false
		}
		mine := CountControlled(q.Game, q.Controller, func(c game.Card) bool { return c.IsCreature() })
		return EachOpponent(q.Game, q.Controller, func(opponent uuid.UUID) bool {
			return CountControlled(q.Game, opponent, func(c game.Card) bool { return c.IsCreature() }) >= mine+n
		})
	}
}
