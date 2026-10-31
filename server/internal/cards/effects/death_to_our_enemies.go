package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Death to Our Enemies — Enchantment — Plan {2}{R} (Marvel Super Heroes,
// #4455):
//
//	"Whenever you cast a noncreature spell, create a tapped Treasure token
//	 and put a plan counter on this enchantment.
//	 When the fourth plan counter is put on this enchantment, sacrifice it.
//	 When you do, it deals 7 damage divided as you choose among one or two
//	 targets."
//
// Construct a Cosmic Cube's plan counters with Riveteers Overlook's
// "sacrifice it, and when you do". The cast trigger is a do of two
// footprinted steps (a tapped Treasure, then the counter, in printed
// order). The payoff watches the placement that CROSSES four, so a fifth
// counter or an enchantment that starts above four fires nothing. The
// sacrifice must really happen (an enchantment that left, or came back as a
// new object, sacrifices nothing and so deals nothing, CR 603.12); the
// reflexive trigger then picks one or two targets and the split as it goes
// on the stack (CR 601.2d through CR 603.3d), every target taking at least
// 1 of the 7, and the damage comes from the sacrificed enchantment as it
// last existed (CR 608.2h), Blaster Hulk's divided damage.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c2b2b743-dfc7-439e-8563-1ccce54422bd",
		Name:         "Death to Our Enemies",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Noncreature(), "Death to Our Enemies — create a tapped Treasure and put a plan counter on this enchantment",
				Do(CreateTokenAdvanced{Spec: Token(TreasureToken()).EntersTapped(), N: 1},
					CounterOnThis{Kind: mshA3PlanCounter, N: 1})),
			On(game.EventCounterPlaced, mshA3PlanCounterReaches(4),
				"Death to Our Enemies — sacrifice it; when you do, it deals 7 damage divided among one or two targets",
				b08OverlookSacrifice("Death to Our Enemies — 7 damage divided as you choose among one or two targets", deathToOurEnemiesDamageBody)),
		},
	})
}

// deathToOurEnemiesDamageBody is the reflexive half: one or two targets of
// any kind, the 7 divided among them as they are announced.
var deathToOurEnemiesDamageBody = game.ReflexiveBody("death-to-our-enemies/divided-damage", simpleBody(blasterHulkDamage),
	constTargets(func() *game.TargetSpec { return TargetAny().WithCount(1, 2).Dividing(Divide(7)) }))
