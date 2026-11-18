package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Brokers Ascendancy — Enchantment {G}{W}{U} (Streets of New Capenna):
//
//	"At the beginning of your end step, put a +1/+1 counter on each
//	 creature you control and a loyalty counter on each planeswalker you
//	 control."
//
// Ajani Steadfast's two sets without his exclusion: both are read once as
// the ability resolves, a creature planeswalker is in both, and each
// placement is its own so counter replacements apply per permanent.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "809edfed-89a9-4327-b7db-1bcec63f3e66",
		Name:         "Brokers Ascendancy",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourEndStep("Brokers Ascendancy — put a +1/+1 counter on each creature you control and a loyalty counter on each planeswalker you control",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item).asGroupMember()
					var creatures, walkers []uuid.UUID
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller != item.Controller {
							continue
						}
						if c.IsCreature() {
							creatures = append(creatures, c.InstanceID)
						}
						if c.IsPlaneswalker() {
							walkers = append(walkers, c.InstanceID)
						}
					}
					for _, id := range creatures {
						if err := (AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
							return err
						}
					}
					for _, id := range walkers {
						if err := (AddCounter{Target: id, Kind: game.CounterLoyalty, N: 1}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				}),
		},
	})
}
