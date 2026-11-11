package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rite of the Raging Storm — Enchantment {3}{R}{R}:
//
//	"Creatures named Lightning Rager can't attack you or planeswalkers you
//	 control.
//	 At the beginning of each player's upkeep, that player creates a 5/1
//	 red Elemental creature token named Lightning Rager. It has trample,
//	 haste, and 'At the beginning of the end step, sacrifice this token.'"
//
// The restriction is written onto every creature with the Rite as its
// source and binds only an attacker named Lightning Rager, judged at
// declaration (cmdr2S34NamedCreaturesCantAttackYouOrYourPlaneswalkers), so
// the Rite's controller's own Rager can't attack them either, and every
// other player's can attack anyone else. The upkeep's player is read off
// the trigger's event, and the token is theirs. The token is a catalog
// template whose trample and haste are printed and whose end-step
// sacrifice is its own trigger.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "984f92dd-cb9b-4421-a8e6-75bfe566be10",
		Name:         "Rite of the Raging Storm",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			cmdr2S34NamedCreaturesCantAttackYouOrYourPlaneswalkers("Creatures named Lightning Rager can't attack you or planeswalkers you control.", "Lightning Rager"),
		},
		Triggered: []game.TriggeredAbility{
			AtEachUpkeep("Rite of the Raging Storm — that player creates a 5/1 Lightning Rager", riteOfTheRagingStormUpkeep),
		},
	})
}

func riteOfTheRagingStormUpkeep(g *game.Game, item *game.StackItem) error {
	player := triggeringActor(item)
	if player == uuid.Nil || g.PlayerByIDForEffect(player) == nil {
		return nil
	}
	return CreateToken{Controller: player, Template: cmdr2S34LightningRagerToken(), N: 1}.Apply(NewContext(g, item))
}
