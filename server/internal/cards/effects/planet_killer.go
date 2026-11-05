package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Planet Killer — Artifact Creature — Construct {8}, 9/9 (Star Trek
// Commander, #4974):
//
//	"Flying, indestructible
//	 This creature attacks each combat if able.
//	 Whenever this creature deals combat damage to a player, put a +1/+1
//	 counter on it for each land that player controls."
//
// The attack requirement is AttacksEachCombat (CR 508.1d). "That player"
// is the player the damage event names; the lands are counted as the
// trigger resolves (CR 608.2h), so a land sacrificed in response is not
// counted. The counters go on Planet Killer only while it is still the
// object that dealt the damage.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4c2c8cc6-bb90-468a-867b-db56f4eb7e59",
		Name:            "Planet Killer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "indestructible"},
		Static:          []game.StaticAbility{AttacksEachCombat()},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Planet Killer — put a +1/+1 counter on it for each land that player controls", planetKillerGrow),
		},
	})
}

// planetKillerGrow counts the damaged player's lands now and puts that
// many +1/+1 counters on this creature.
func planetKillerGrow(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil || !sourceIsStillThisPermanent(g, item) {
		return nil
	}
	lands := len(landsControlledByPlayer(g, item.Trigger.Event.Target))
	if lands == 0 {
		return nil
	}
	return AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: lands}.Apply(NewContext(g, item))
}
