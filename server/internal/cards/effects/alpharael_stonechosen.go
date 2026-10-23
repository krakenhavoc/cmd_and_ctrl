package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Alpharael, Stonechosen — Legendary Creature — Human Cleric {3}{B}{B},
// 3/3 (Edge of Eternities):
//
//	"Ward—Discard a card at random.
//	 Void — Whenever Alpharael attacks, if a nonland permanent left the
//	 battlefield this turn or a spell was warped this turn, defending
//	 player loses half their life, rounded up."
//
// The ward is Ward(WardDiscardAtRandom(1)) (ward.go): the payer decides
// only whether to pay, and the discarded card is picked at random from
// their hand (CR 701.9b); an empty hand cannot pay, so the spell or
// ability is countered. Void is an intervening if (void.go). "Defending
// player" is the player behind what Alpharael attacked, a planeswalker's
// or battle's protector included, and half is of the life total as the
// ability resolves, rounded up (CR 107.1a); a player at 0 or less loses
// nothing more.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ed43f1f9-ef96-47bc-8a54-e59154319f9d",
		Name:         "Alpharael, Stonechosen",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Ward(WardDiscardAtRandom(1), "Alpharael, Stonechosen — ward, discard a card at random"),
			WheneverThisAttacksIfVoid("Alpharael, Stonechosen — defending player loses half their life, rounded up",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					return playerLosesHalfTheirLife(g, item, g.DefendingPlayerForAttackForEffect(item.Trigger.Event.Target))
				}),
		},
	})
}
