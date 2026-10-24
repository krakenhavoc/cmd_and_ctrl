package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nightkin Ambusher — Creature — Mutant Warrior {2}{U}{B}, 4/4:
//
//	"Ward {2}
//	 When this creature enters, target player gets four rad counters.
//	 This creature can't be blocked as long as defending player has a
//	 rad counter."
//
// #2809. Ward is the shared declaration. The enter trigger targets a
// player, any player (the Ambusher's own controller included), and
// gives the counters as its controller.
//
// The evasion is a block rule on the attacker (CantBeBlockedWhile
// scoped to itself) whose condition is read each time a block is
// checked: the defending player is the one behind the attacker's
// attack target (CR 508.5), and that player's rad counters are counted
// as the blockers are declared, so a counter given or removed after
// attackers were declared changes the answer. A creature that is not
// attacking has no defending player, so the clause is false for it.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "2808d652-907d-4071-b386-814507a1549e",
		Name:         "Nightkin Ambusher",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{2}"), "Nightkin Ambusher — ward {2}"),
			Targeting(WhenThisEnters("Nightkin Ambusher — target player gets four rad counters",
				targetPlayerGetsRadCounters(4)), TargetPlayer("target player")),
		},
		BlockRules: []game.BlockRule{
			CantBeBlockedWhile(OnSelf(), DefendingPlayerHasARadCounter),
		},
	})
}
