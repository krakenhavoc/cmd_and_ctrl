package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bishop of the Bloodstained — Creature — Vampire Cleric {3}{B}{B}, 3/3
// (The List):
//
//	"When this creature enters, target opponent loses 1 life for each
//	 Vampire you control."
//
// Shaman of the Pack's drain counting Vampire permanents (Champion of Dusk's
// count): taken as the trigger resolves, the Bishop counts itself, and the
// subtype is read through the source's text (CR 612.2). Life loss, not damage.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e9398f3b-c457-4fff-9b09-377c612b6a48",
		Name:         "Bishop of the Bloodstained",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: b06SelfETB,
			Targets:   TargetPlayer("target opponent", Opponent()),
			Key:       "Bishop of the Bloodstained — target opponent loses 1 life for each Vampire you control",
			Effect: func(g *game.Game, item *game.StackItem) error {
				vampire := g.TextReaderForEffect(item).TextSubtype("Vampire")
				n := 0
				for _, c := range g.BattlefieldCardsForEffect() {
					if c.Controller == item.Controller && c.HasSubtype(vampire) {
						n++
					}
				}
				return b43TargetPlayerLoses(g, item, n)
			},
		}},
	})
}
