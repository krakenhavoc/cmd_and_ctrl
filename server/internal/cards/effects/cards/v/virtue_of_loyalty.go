package cardsv

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Virtue of Loyalty // Ardenvale Fealty — Enchantment {3}{W}{W} //
// Instant — Adventure {1}{W} (Wilds of Eldraine, #4471):
//
//	Virtue of Loyalty — "At the beginning of your end step, put a +1/+1
//	 counter on each creature you control. Untap those creatures."
//	Ardenvale Fealty — "Create a 2/2 white Knight creature token with
//	 vigilance. (Then exile this card. You may cast the enchantment later
//	 from exile.)"
//
// Both faces register under the Adventure keyspace (the bare oracle ID
// and "#1"). The end-step trigger reads the creatures its controller
// controls as it resolves and puts one +1/+1 counter on each, one
// placement per creature (Black Sun's Zenith's avB24PutCountersOnEachThen,
// so a CR 616 ordering prompt is a restore point). "Those creatures" are
// that same list, carried on a keyed tail that untaps each one still on
// the battlefield once the last counter has settled. Ardenvale Fealty is
// the token table's vigilant Knight. The Adventure lifecycle is the
// engine's (CR 715.3d).
//
// No simplification.
const VirtueOfLoyaltyOracleID = "f61f4307-4eec-476a-97e5-d4d3da2b54c3"

func init() {
	effects.Register(effects.Spec{
		OracleID:     VirtueOfLoyaltyOracleID,
		Name:         "Virtue of Loyalty",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStep("Virtue of Loyalty — put a +1/+1 counter on each creature you control, then untap them",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					var ids []uuid.UUID
					for _, c := range effects.MatchingBattlefield(ctx, effects.And(effects.Creature(), effects.YouControl())) {
						ids = append(ids, c.InstanceID)
					}
					untap := ctx.Tail(virtueOfLoyaltyUntapThose)
					untap.Carry = append([]uuid.UUID(nil), ids...)
					return effects.AvB24PutCountersOnEachThen(ctx, ids, game.CounterPlusOne, 1, untap)
				}),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     VirtueOfLoyaltyOracleID + "#1",
		Name:         "Ardenvale Fealty",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.CreateToken{Controller: item.Controller, Template: effects.TokenCard("2/2 white Knight with vigilance"), N: 1}.Apply(ctx)
		},
	})
}

// virtueOfLoyaltyUntapThose is "Untap those creatures": Carry is the list
// the counters went on, and each one still on the battlefield is untapped.
// An on-disk key: never renamed, never reused.
var virtueOfLoyaltyUntapThose = effects.CardTail("virtue-of-loyalty/untap-those-creatures", func(ctx *effects.Context, o game.TailOutcome) error {
	group := ctx.AsGroupMember()
	for _, id := range o.Tail.Carry {
		if !effects.OnBattlefield(ctx.Game, id) {
			continue
		}
		if err := (effects.UntapTarget{Target: id}).Apply(group); err != nil {
			return err
		}
	}
	return nil
})
