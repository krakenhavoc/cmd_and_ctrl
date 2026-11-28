package cardsx

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Zombie Mob — Creature — Zombie {2}{B}{B}, 2/0 (Mirage, S82 OLD b805):
//
//	"This creature enters with a +1/+1 counter on it for each creature card in
//	 your graveyard.
//	 When this creature enters, exile all creature cards from your graveyard."
//
// Academy Elite's counted entry replacement (CR 614.1c): the graveyard is
// counted as the Mob enters, so the 2/0 is already grown when state-based
// actions look at it. The enters trigger then exiles every creature card
// there, ids snapshotted before the first exile (Bojuka Bog's reading).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c56f0f4c-c85a-4931-b22a-42dc5d1dc7dd",
		Name:         "Zombie Mob",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.B19EntersWithCountersCounted(game.CounterPlusOne, func(g *game.Game, src *game.Card) int {
				n := 0
				if p := g.PlayerByIDForEffect(src.Controller); p != nil && p.Graveyard != nil {
					for _, c := range p.Graveyard.Cards {
						if c.IsCreature() {
							n++
						}
					}
				}
				return n
			}, "Zombie Mob enters with a +1/+1 counter on it for each creature card in your graveyard"),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Zombie Mob — exile all creature cards from your graveyard",
				func(g *game.Game, item *game.StackItem) error {
					p := g.PlayerByIDForEffect(item.Controller)
					if p == nil || p.Graveyard == nil {
						return nil
					}
					var ids []uuid.UUID
					for _, c := range p.Graveyard.Cards {
						if c.IsCreature() {
							ids = append(ids, c.InstanceID)
						}
					}
					ctx := effects.NewContext(g, item)
					for _, id := range ids {
						if err := (effects.ExileTarget{Target: id}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				}),
		},
	})
}
