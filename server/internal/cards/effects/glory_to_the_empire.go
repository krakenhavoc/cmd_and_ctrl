package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Glory to the Empire! — Enchantment {5}{R} (Star Trek Commander, #4974):
//
//	"When this enchantment enters, create a number of 2/1 red Klingon
//	 Warrior creature tokens with haste equal to the number of opponents
//	 you have.
//	 Creatures you control have melee. (Whenever a creature with melee
//	 attacks, it gets +1/+1 until end of turn for each opponent you
//	 attacked this combat.)"
//
// The tokens are Cha'DIch Investigator's printed Klingon Warrior, one
// per opponent still in the game as the trigger resolves (Goblin
// Goliath's count). Melee is the engine's keyword (CR 702.121a), granted
// in layer 6 to every creature you control the way Titania, Proud
// Pummeler grants it: a creature that prints melee too has two instances
// and each triggers (CR 702.121b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "22c53662-51b9-4a8c-8ec7-e8b456a268b7",
		Name:         "Glory to the Empire!",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Glory to the Empire! — create a 2/1 Klingon Warrior with haste for each opponent", func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				n := len(ctx.Opponents())
				if n == 0 {
					return nil
				}
				return CreateToken{Controller: ctx.Controller(), Template: TokenCard("2/1 red Klingon Warrior with haste"), N: n}.Apply(ctx)
			}),
		},
		Static: []game.StaticAbility{
			TribalKeywordGrant(TribeFilter{YoursOnly: true}, game.KeywordMelee),
		},
	})
}
