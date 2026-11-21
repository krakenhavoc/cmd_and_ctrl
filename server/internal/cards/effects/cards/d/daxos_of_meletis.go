package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Daxos of Meletis — Legendary Creature — Human Soldier {1}{W}{U}, 2/2:
//
//	"Daxos can't be blocked by creatures with power 3 or greater.
//	 Whenever Daxos deals combat damage to a player, exile the top card
//	 of that player's library. You gain life equal to that card's mana
//	 value. Until end of turn, you may cast that card and you may spend
//	 mana as though it were mana of any color to cast that spell."
//
// The block rule is the printed evasion (CantBeBlockedBy). The trigger
// exiles the damaged player's top card with a cast-only, any-colour
// permission for this turn (a land can't be cast, so it is stranded);
// the life gain follows the exile, told the card that really landed in
// exile (CR 608.2c), and reads its mana value there (a land's is 0).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4cf10cb6-b082-4c0d-89e9-43cd39716201",
		Name:         "Daxos of Meletis",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedBy(effects.OnSelf(), effects.PowerGE(3), "creatures with power 3 or greater"),
		},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer(
				"Daxos of Meletis — exile the top card of that player's library, gain life equal to its mana value, and you may cast it this turn",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					perm := game.CastPermission{
						CastOnly: true,
						AnyColor: true,
						Duration: g.UntilEndOfTurnDuration(),
					}
					return g.ExileTopWithPermissionKeyedForEffect(item.Trigger.Event.Target, item.Controller, 1, perm,
						effects.NewContext(g, item).ItemTail(daxosGainLife))
				}),
		},
	})
}

// daxosGainLife is "you gain life equal to that card's mana value", told
// the card the exile landed. An on-disk key: never renamed, never reused.
var daxosGainLife = effects.CardTail("daxos-of-meletis/gain-life", func(ctx *effects.Context, o game.TailOutcome) error {
	for _, id := range o.IDs {
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok {
			continue
		}
		if err := (effects.GainLife{Amount: c.ManaValue()}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})
