package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Claim the Kingdom — {1}{G} Enchantment — Plan (Marvel Super Heroes,
// #4455):
//
//	"Landfall — Whenever a land you control enters, put a +1/+1 counter on
//	 target creature you control and a plan counter on this enchantment.
//	 When the fourth plan counter is put on this enchantment, sacrifice it.
//	 When you do, put an indestructible counter on target creature you
//	 control."
//
// Construct a Cosmic Cube's plan shape. The landfall trigger targets a
// creature you control as it goes on the stack and, as it resolves, puts the
// +1/+1 counter on it and the plan counter on the enchantment (a target gone
// by then fizzles the whole ability, CR 608.2b). The fourth counter is a
// second trigger watching the plan counter total reach four from below; it
// sacrifices the enchantment and only then, as the reflexive "when you do"
// (CR 603.12), puts the indestructible counter trigger on the stack with its
// own target, chosen after the sacrifice.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3cd15225-e9b6-4998-b545-76536c72e948",
		Name:         "Claim the Kingdom",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				Landfall("Claim the Kingdom — landfall: a +1/+1 counter on target creature you control and a plan counter on this", claimTheKingdomLandfall),
				TargetCreature("target creature you control", YouControl())),
			On(game.EventCounterPlaced, mshB2PlanCounterReached(4),
				"Claim the Kingdom — sacrifice it; when you do, put an indestructible counter on target creature you control",
				mshB2SacrificeThisThenWhenYouDo("Claim the Kingdom — put an indestructible counter on target creature you control", claimTheKingdomIndestructibleBody)),
		},
	})
}

// claimTheKingdomLandfall is the landfall body: the counter on the target,
// then the plan counter on the enchantment.
func claimTheKingdomLandfall(g *game.Game, item *game.StackItem) error {
	if err := PlusOneCounterOnTheTarget(g, item); err != nil {
		return err
	}
	return AddCounter{Target: item.SourceCardID, Kind: "plan", N: 1}.Apply(NewContext(g, item))
}

// claimTheKingdomIndestructibleBody is the reflexive half: an indestructible
// counter (CR 122.1b) on the chosen creature. An on-disk key.
var claimTheKingdomIndestructibleBody = game.ReflexiveBody("claim-the-kingdom/indestructible",
	simpleBody(func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		for _, t := range ctx.LegalTargets() {
			if t.Kind == game.TargetCard {
				return AddCounter{Target: t.ID, Kind: game.CounterIndestructible, N: 1}.Apply(ctx)
			}
		}
		return nil
	}),
	constTargets(func() *game.TargetSpec { return TargetCreature("target creature you control", YouControl()) }))
