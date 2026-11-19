package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rapacious One — Creature — Eldrazi Drone {5}{R}, 5/4 (Rise of the Eldrazi,
// #5581):
//
//	"Trample
//	 Whenever this creature deals combat damage to a player, create that
//	 many 0/1 colorless Eldrazi Spawn creature tokens. They have "Sacrifice
//	 this token: Add {C}.""
//
// "That many" is the damage the event carries, after prevention and
// trample's assignment, as Ant-Man, Elusive Avenger counts its Treasures.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3af03c99-d215-477b-ba49-51f4e87907ff",
		Name:            "Rapacious One",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Rapacious One — create that many Eldrazi Spawn",
				Do(CreateTokensForEach{
					Template: EldraziSpawnToken(),
					Name:     "the combat damage dealt",
					Count:    func(ctx *Context) int { return ctx.Trigger().Event.Amount },
				})),
		},
	})
}
