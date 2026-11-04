package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Black Panther, Wakandan King — Legendary Creature — Human Noble Hero
// {G}{W}, 2/2:
//
//	"First strike
//	 Survey the Realm — Whenever Black Panther or another creature you
//	 control enters, put a +1/+1 counter on target land you control.
//	 Mine Vibranium — {3}: Move all +1/+1 counters from target land you
//	 control onto target creature. If one or more +1/+1 counters are
//	 moved this way, you gain that much life and draw a card."
//
// Survey the Realm is Mabel's trigger (CreatureEnteredUnderYourControl
// includes Black Panther itself) aimed at a land you control; the
// target is chosen as the ability goes on the stack, and the trigger is
// dropped if there is no land to choose.
//
// Mine Vibranium has two target clauses. Each is re-checked as it
// resolves (CR 608.2b): if either is gone or no longer legal, no counter
// moves, so there is no life and no card. Otherwise every +1/+1 counter
// on the land comes off and goes onto the creature, and the controller
// gains that many life and draws one card, in printed order and each
// step waiting for any prompt the one before it asks (CR 608.2c). A land
// with no +1/+1 counters moves nothing and gains and draws nothing. The
// creature may be any creature, including an opponent's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "48deff27-b421-430b-a484-55b4a593010a",
		Name:            "Black Panther, Wakandan King",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				On(game.EventETB, CreatureEnteredUnderYourControl,
					"Black Panther, Wakandan King — Survey the Realm: a +1/+1 counter on target land you control",
					putACounterOnTheTarget(game.CounterPlusOne)),
				TargetPermanent("target land you control", Land(), YouControl())),
		},
		Activated: []ActivatedAbility{{
			Label: "Mine Vibranium — {3}: Move all +1/+1 counters from target land you control onto target creature. If one or more +1/+1 counters are moved this way, you gain that much life and draw a card.",
			Cost:  ManaCost("{3}"),
			Targets: Clauses(
				TargetPermanent("target land you control", Land(), YouControl()),
				TargetCreature("target creature"),
			),
			Effect: Do(
				Instruction(func(ctx *Context) error {
					land, ok := ctx.ClauseTarget(0)
					creature, ok2 := ctx.ClauseTarget(1)
					if !ok || !ok2 || land.Kind != game.TargetCard || creature.Kind != game.TargetCard {
						return nil
					}
					card, found := ctx.Game.LookupCardForEffect(land.ID)
					if !found {
						return nil
					}
					n := card.Counters[game.CounterPlusOne]
					if n <= 0 {
						return nil
					}
					if err := (AddCounter{Target: land.ID, Kind: game.CounterPlusOne, N: -n}).Apply(ctx); err != nil {
						return err
					}
					if err := (AddCounter{Target: creature.ID, Kind: game.CounterPlusOne, N: n}).Apply(ctx); err != nil {
						return err
					}
					ctx.RecordStep(blackPantherMoved, nil, n)
					return nil
				}),
				Instruction(func(ctx *Context) error {
					if n := ctx.StepRecord(blackPantherMoved).N(); n > 0 {
						return GainLife{Amount: n}.Apply(ctx)
					}
					return nil
				}),
				Instruction(func(ctx *Context) error {
					if ctx.StepRecord(blackPantherMoved).N() > 0 {
						return DrawCards{N: 1}.Apply(ctx)
					}
					return nil
				}),
			),
		}},
	})
}

const blackPantherMoved = "black-panther-wakandan-king/moved"
