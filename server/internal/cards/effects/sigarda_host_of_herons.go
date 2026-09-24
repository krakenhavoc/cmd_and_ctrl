package effects

// Sigarda, Host of Herons — Legendary Creature — Angel, {2}{G}{W}{W}, 5/5:
//
//	"Flying, hexproof"
//	"Spells and abilities your opponents control can't cause you to
//	 sacrifice permanents."
//
// A five-mana 5/5 flier that cannot be targeted — the archetypal
// voltron commander, and the one whose body is closest to fully
// modelled today. Five evasive power is five turns to the 21-damage
// clock (CR 903.10a), or three with any two pump spells from this
// sprint's protection suite.
//
// Both printed keywords are real: `hexproof` has been enforced at
// the targeting choke point since S23, and flying has been enforced
// in block validation since S18. Sigarda genuinely cannot be Doom
// Bladed and genuinely cannot be blocked by a ground creature.
//
// # DECLARED SIMPLIFICATION — the sacrifice-protection clause
//
// "Spells and abilities your opponents control can't cause you to
// sacrifice permanents" is not enforced. It is a continuous effect
// that modifies what an OPPONENT'S effect is allowed to do to a
// third object, which is a shape the engine has nowhere to put: the
// sacrifice path (`sacrificePermanentLocked`,
// `EachPlayerSacrifices`) takes a player and a permanent and asks
// nobody's permission, and there is no "can't cause" hook to hang
// this on. Modelling it properly means a prohibition layer the
// engine does not have and that only a handful of cards want.
//
// So an opponent's Butcher of Malakir trigger still makes you
// sacrifice. That is WEAKER than printed, which is the acceptable
// direction, and it is declared here rather than left implicit
// because #350 is the ticket about notes outliving their blockers.
// Two of Sigarda's three lines are fully live.
func init() {
	Register(Spec{
		OracleID:        "e55104e2-4900-48de-b288-d3e6abd5e09e",
		Name:            "Sigarda, Host of Herons",
		PrintedKeywords: []string{"flying", "hexproof"},
	})
}
