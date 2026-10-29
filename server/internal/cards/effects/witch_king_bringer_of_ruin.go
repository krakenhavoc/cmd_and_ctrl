package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Witch-king, Bringer of Ruin — Legendary Creature — Wraith Noble
// {4}{B}{B}, 5/3 (The Hobbit, #4454):
//
//	"Flying
//	 Whenever Witch-king attacks, defending player sacrifices a creature
//	 with the least power among creatures they control."
//
// An edict whose choice is restricted to the creatures tied for the least
// power the defending player controls: the sacrifice prompt (the defender
// chooses, nothing targets, so hexproof is no defence) offers only those.
// The defending player is the player attacked, or the controller of the
// planeswalker or battle attacked (CR 506.3d). A defender who has left the
// game sacrifices nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "afcb21bb-aeb3-4024-9ce1-68ba7fb64b2d",
		Name:            "Witch-king, Bringer of Ruin",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Witch-king, Bringer of Ruin — defending player sacrifices a creature with the least power", witchKingBringerOfRuinEdict),
		},
	})
}

// witchKingBringerOfRuinEdict queues the defending player's sacrifice.
func witchKingBringerOfRuinEdict(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	defender := defendingPlayerOfAttack(g, item.Trigger.Event)
	if defender == uuid.Nil || !defendingPlayerStillIn(g, defender) {
		return nil
	}
	g.PlayerSacrificesForEffect(item.SourceCardID, defender,
		sacrificeSpec("a creature with the least power among creatures you control", Creature(), hasTheLeastPowerAmongItsControllersCreatures),
		"Witch-king, Bringer of Ruin — sacrifice a creature with the least power")
	return nil
}

// hasTheLeastPowerAmongItsControllersCreatures is "a creature with the least
// power among creatures they control": no creature its controller controls
// has less power than this one.
func hasTheLeastPowerAmongItsControllersCreatures(g *game.Game, _ uuid.UUID, c game.Card) bool {
	power := c.CurrentPower()
	for _, other := range g.BattlefieldCardsForEffect() {
		if other.Controller == c.Controller && other.IsCreature() && other.CurrentPower() < power {
			return false
		}
	}
	return true
}
