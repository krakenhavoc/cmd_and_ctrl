package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Atlas, Sizable Stooge — Legendary Creature — Human Villain
// {2}{B/G}{B/G}, 5/4 (Marvel Super Heroes Commander, #4455):
//
//	"Reach
//	 Whenever Atlas attacks or blocks, you gain 1 life for each creature
//	 you control with power 4 or greater."
//
// The count is taken as the trigger resolves, Atlas itself included when
// its power is 4 or more. Getaway Car's attacks-or-blocks shape.
func init() {
	Register(Spec{
		OracleID:        "efd5f51f-931d-46a9-9773-f74cfc318c23",
		Name:            "Atlas, Sizable Stooge",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventAttack, game.EventBlock},
				func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return attackDeclared(ev, source) || selfBlocksOnce(ev, source)
				},
				"Atlas — you gain 1 life for each creature you control with power 4 or greater",
				func(g *game.Game, item *game.StackItem) error {
					n := 0
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller == item.Controller && c.IsCreature() && c.CurrentPower() >= 4 {
							n++
						}
					}
					return GainLife{Amount: n}.Apply(NewContext(g, item))
				}),
		},
	})
}
