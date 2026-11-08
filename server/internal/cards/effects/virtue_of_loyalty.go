package effects

import (
	"github.com/google/uuid"

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
const virtueOfLoyaltyOracleID = "f61f4307-4eec-476a-97e5-d4d3da2b54c3"

func init() {
	Register(Spec{
		OracleID:     virtueOfLoyaltyOracleID,
		Name:         "Virtue of Loyalty",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourEndStep("Virtue of Loyalty — put a +1/+1 counter on each creature you control, then untap them",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					var ids []uuid.UUID
					for _, c := range MatchingBattlefield(ctx, And(Creature(), YouControl())) {
						ids = append(ids, c.InstanceID)
					}
					untap := ctx.Tail(virtueOfLoyaltyUntapThose)
					untap.Carry = append([]uuid.UUID(nil), ids...)
					return avB24PutCountersOnEachThen(ctx, ids, game.CounterPlusOne, 1, untap)
				}),
		},
	})
	Register(Spec{
		OracleID:     virtueOfLoyaltyOracleID + "#1",
		Name:         "Ardenvale Fealty",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return CreateToken{Controller: item.Controller, Template: TokenCard("2/2 white Knight with vigilance"), N: 1}.Apply(ctx)
		},
	})
}

// virtueOfLoyaltyUntapThose is "Untap those creatures": Carry is the list
// the counters went on, and each one still on the battlefield is untapped.
// An on-disk key: never renamed, never reused.
var virtueOfLoyaltyUntapThose = CardTail("virtue-of-loyalty/untap-those-creatures", func(ctx *Context, o game.TailOutcome) error {
	group := ctx.asGroupMember()
	for _, id := range o.Tail.Carry {
		if !onBattlefield(ctx.Game, id) {
			continue
		}
		if err := (UntapTarget{Target: id}).Apply(group); err != nil {
			return err
		}
	}
	return nil
})
