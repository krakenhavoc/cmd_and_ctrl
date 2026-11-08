package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blood Spatter Analysis — Enchantment {B}{R} (Murders at Karlov Manor,
// #4469):
//
//	"When this enchantment enters, it deals 3 damage to target creature
//	 an opponent controls.
//	 Whenever one or more creatures die, mill a card and put a bloodstain
//	 counter on this enchantment. Then sacrifice it if it has five or more
//	 bloodstain counters on it. When you do, return target creature card
//	 from your graveyard to your hand."
//
// The enters trigger is DealDamageToTheTarget. The death trigger fires
// once for a batch of deaths, anyone's creatures (OncePerBatch over
// EventLTB, as Morbid Opportunist's is, without "other"). It mills one,
// then puts the counter on this object (CounterOnThis, nothing if the
// enchantment has left or come back as a new object), then counts the
// counters as they stand and sacrifices it at five or more. The "when
// you do" rides the sacrifice's own continuation (ifItWentThen), so the
// reflexive trigger exists only if the enchantment really left; its
// target is chosen as it goes on the stack (CR 603.12), and is Undead
// Butler's registered "return target creature card from your graveyard
// to your hand" body.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "aa929252-5dcc-4c9b-9e7c-61d0bef98d6d",
		Name:         "Blood Spatter Analysis",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Blood Spatter Analysis — 3 damage to target creature an opponent controls",
				DealDamageToTheTarget(3)),
				TargetCreature("target creature an opponent controls", OpponentControls())),
			OncePerBatch(On(game.EventLTB, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, ok := diedCreature(ev, g)
				return ok
			}, "Blood Spatter Analysis — mill a card and put a bloodstain counter on it; at five, sacrifice it",
				Do(MillCards{N: 1}, CounterOnThis{Kind: bloodSpatterCounter, N: 1}, Instruction(bloodSpatterSacrificeAtFive)))),
		},
	})
}

// bloodSpatterCounter is the counter kind the enchantment collects.
const bloodSpatterCounter = "bloodstain"

// bloodSpatterSacrificeAtFive is "Then sacrifice it if it has five or
// more bloodstain counters on it", with "When you do" on the sacrifice's
// outcome.
func bloodSpatterSacrificeAtFive(ctx *Context) error {
	if sourceIsNewObject(ctx.Game, ctx.Item) {
		return nil
	}
	c, ok := ctx.Game.LookupCardForEffect(ctx.Source())
	if !ok || !onBattlefield(ctx.Game, ctx.Source()) || c.Counters[bloodSpatterCounter] < 5 {
		return nil
	}
	return SacrificePermanent{Target: ctx.Source(), ThenTail: ctx.ifItWentThen(bloodSpatterReturn)}.Apply(ctx)
}

// bloodSpatterReturn is "When you do, return target creature card from
// your graveyard to your hand."
var bloodSpatterReturn = RegisterThenSteps("then-steps/blood-spatter-analysis-return",
	Instruction(func(ctx *Context) error {
		return WhenYouDo("Blood Spatter Analysis — return target creature card from your graveyard to your hand",
			undeadButlerReturnBody).Apply(ctx)
	}),
)
