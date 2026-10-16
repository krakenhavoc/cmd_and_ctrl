package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Construct a Cosmic Cube — Enchantment — Plan {2}{B}:
//
//	"Whenever you draw your second card each turn, create a 2/1 black
//	 Villain creature token with menace and put a plan counter on this
//	 enchantment.
//	 When the seventh plan counter is put on this enchantment, sacrifice
//	 it. When you do, you control target opponent during their next
//	 turn. (You see all cards that player could see and make all
//	 decisions for them.)"
//
// Two triggers and a reflexive one. The draw trigger makes the Villain
// and adds the counter in one resolution (they are one printed
// sentence). The payoff is a trigger on the placement that CROSSES
// seven (Midnight Clock's shape: a counter that starts above seven, or
// an eighth, fires nothing). It sacrifices the enchantment and, only if
// that happened, creates the CR 603.12 reflexive "when you do" that
// targets an opponent as it goes on the stack, so the opponent chosen
// can respond and a Cube that left in response (a new object) creates
// no control at all (the Overlook lands' rule, #1432).
//
// The control itself is turn control (CR 723.1, ADR 0167, #3857), the
// plain form with no extra turn afterwards.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "369aa9a6-af8a-432c-adff-ced2bb203dc7",
		Name:         "Construct a Cosmic Cube",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b41DrewYourSecondCardThisTurn(ev, source, g)
			}, "Construct a Cosmic Cube — create a 2/1 Villain with menace and put a plan counter on this enchantment",
				cosmicCubeVillainAndCounter),
			On(game.EventCounterPlaced, cosmicCubeSeventhPlan,
				"Construct a Cosmic Cube — sacrifice it", cosmicCubeSacrifice),
		},
	})
}

// cosmicCubeVillainAndCounter is the draw trigger: the token, then the
// counter. Package-level so the stack item captures nothing.
func cosmicCubeVillainAndCounter(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := (CreateToken{Template: TokenCard("2/1 black Villain with menace"), N: 1}).Apply(ctx); err != nil {
		return err
	}
	return AddCounter{Target: item.SourceCardID, Kind: "plan", N: 1}.Apply(ctx)
}

// cosmicCubeSeventhPlan is "when the seventh plan counter is put on this
// enchantment": a plan-counter placement on the source whose new total
// reaches seven from below.
func cosmicCubeSeventhPlan(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventCounterPlaced || ev.Target != source.InstanceID || ev.Label != "plan" || ev.Amount < 7 {
		return false
	}
	before, _ := b12CounterTotalBefore(ev, g)
	return before < 7
}

// cosmicCubeControlBody is the reflexive half: "you control target
// opponent during their next turn". Registered once (ADR 0041 P9) so a
// table with it waiting is a restore point.
var cosmicCubeControlBody = game.ReflexiveBody("construct-a-cosmic-cube/control",
	simpleBody(controlTargetPlayerNextTurn("Construct a Cosmic Cube")),
	constTargets(func() *game.TargetSpec { return TargetPlayer("target opponent", Opponent()) }))

// cosmicCubeSacrifice sacrifices the Cube and, only if that happened,
// creates the reflexive trigger. A Cube that already left (or came back
// as a new object) sacrifices nothing and controls nobody.
func cosmicCubeSacrifice(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if z := g.FindCardZoneForEffect(item.SourceCardID); z == nil || z.Kind != game.ZoneBattlefield ||
		ctx.isNewSourceObjectAsThis(item.SourceCardID) {
		return nil
	}
	if err := (SacrificePermanent{Target: item.SourceCardID}).Apply(ctx); err != nil {
		return err
	}
	return WhenYouDo("Construct a Cosmic Cube — you control target opponent during their next turn",
		cosmicCubeControlBody).Apply(ctx)
}
