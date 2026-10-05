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
// The sacrifice-protection clause is `OpponentEffectProtections`
// (#2178), read off the battlefield whenever an opponent's resolving
// spell or ability asks this player to sacrifice: Butcher of Malakir's
// trigger, a Fleshbag-style edict, annihilator and sacrifice-all
// sweeps all skip its controller. A sacrifice they choose to pay as a
// cost, and their own effects, are untouched. No simplification.
func init() {
	Register(Spec{
		OracleID:                  "e55104e2-4900-48de-b288-d3e6abd5e09e",
		Name:                      "Sigarda, Host of Herons",
		Completeness:              CompletenessFull,
		PrintedKeywords:           []string{"flying", "hexproof"},
		OpponentEffectProtections: CantBeMadeToSacrifice(),
	})
}
