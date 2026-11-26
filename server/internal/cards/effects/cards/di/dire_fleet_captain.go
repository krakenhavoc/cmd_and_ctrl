package cardsdi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dire Fleet Captain — Creature — Orc Pirate {B}{R}, 2/2 (Ixalan):
//
//	"Whenever this creature attacks, it gets +1/+1 until end of turn for each
//	 other attacking Pirate."
//
// Shared Animosity's count, read as the trigger resolves (CR 608.2h): the
// other attacking creatures, whoever controls them, that are Pirates through
// the layers (a changeling counts). The Captain itself is never counted, and
// one that has left combat gets nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "68f1da0d-1143-457a-b709-1da3c060a45d",
		Name:         "Dire Fleet Captain",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Dire Fleet Captain — it gets +1/+1 for each other attacking Pirate",
				func(g *game.Game, item *game.StackItem) error {
					self, ok := g.LookupCardForEffect(item.SourceCardID)
					if !ok || !effects.SourceIsStillThisPermanent(g, item) || self.AttackingTarget == uuid.Nil {
						return nil
					}
					n := 0
					pirate := self.TextSubtype("Pirate")
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.InstanceID != self.InstanceID && c.AttackingTarget != uuid.Nil && c.HasSubtypeWord(pirate) {
							n++
						}
					}
					return effects.BoostUntilEOT{
						Target: self.InstanceID, Power: n, Toughness: n,
						Label: "Dire Fleet Captain — +1/+1 for each other attacking Pirate",
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
