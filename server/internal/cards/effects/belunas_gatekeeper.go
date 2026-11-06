package effects

// Beluna's Gatekeeper // Entry Denied — Creature — Giant Soldier {5}{U},
// 6/5 // Sorcery — Adventure {1}{U} (Wilds of Eldraine, #4471):
//
//	Beluna's Gatekeeper — no rules text.
//	Entry Denied — "Return target creature you don't control with mana
//	 value 3 or less to its owner's hand. (Then exile this card. You may
//	 cast the creature later from exile.)"
//
// Both faces register under the Adventure keyspace (the bare oracle ID
// and "#1"): the vanilla creature so the card does not wear the
// "unimplemented" badge in hand, as Besotted Knight's does. Entry Denied
// is Void Squall's bounceTheTarget behind Kobayashi Maru Scenario's
// predicates (a creature someone else controls, mana value 3 or less),
// judged at announce and again at resolution, so a creature that grew
// past the bound or changed hands in response is no longer a legal
// target.
//
// No simplification.
const belunasGatekeeperOracleID = "52fd9309-b849-419a-8d19-5882c00a72b4"

func init() {
	Register(Spec{
		OracleID:     belunasGatekeeperOracleID,
		Name:         "Beluna's Gatekeeper",
		Completeness: CompletenessFull,
	})
	Register(Spec{
		OracleID:     belunasGatekeeperOracleID + "#1",
		Name:         "Entry Denied",
		Completeness: CompletenessFull,
		Targets: TargetCreature("target creature you don't control with mana value 3 or less",
			OpponentControls(), ManaValueLE(3)),
		OnResolve: bounceTheTarget,
	})
}
