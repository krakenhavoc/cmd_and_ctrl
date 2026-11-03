package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Brazen Borrower // Petty Theft — an adventure card (Secrets of
// Strixhaven Commander, #4834):
//
//	Brazen Borrower {1}{U}{U}, Creature — Faerie Rogue 3/1:
//	"Flash
//	 Flying
//	 This creature can block only creatures with flying."
//
//	Petty Theft {1}{U}, Instant — Adventure:
//	"Return target nonland permanent an opponent controls to its owner's
//	 hand."
//
// The adventure layout is Foulmire Knight's: the creature half under the
// bare oracle ID, the Adventure half under "<oracle_id>#1", and the
// engine exiles the resolved Adventure with permission to cast the
// creature (CR 715.3d). "Can block only creatures with flying" is
// Stormbound Geist's block rule, a restriction checked at the
// declaration (CR 509.1b).
func init() {
	const oracleID = "c7b044c3-3cfa-407e-bf20-2875e8e04b7b"

	// Face 0 — the creature.
	Register(Spec{
		OracleID:        oracleID,
		Name:            "Brazen Borrower",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		BlockRules: []game.BlockRule{
			CantBlockAttackers(OnSelf(), OnMatching(WithoutKeyword("flying")), "it can block only creatures with flying"),
		},
	})

	// Face 1 — Petty Theft, an Adventure instant (CR 715.3a).
	Register(Spec{
		OracleID:     oracleID + "#1",
		Name:         "Petty Theft",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target nonland permanent an opponent controls", Nonland(), OpponentControls()),
		OnResolve:    bounceTheTarget,
	})
}
