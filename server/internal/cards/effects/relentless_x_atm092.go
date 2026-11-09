package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Relentless X-ATM092 — Artifact Creature — Robot Spider {6}, 6/5 (Final
// Fantasy, #4462):
//
//	"This creature can't be blocked except by three or more creatures.
//	 {8}: Return this card from your graveyard to the battlefield tapped
//	 with a finality counter on it."
//
// Rampaging Ceratops' minimum-blockers rule, and Gastal Thrillroller's
// graveyard return with the tapped entry: the card comes back tapped with
// the counter as it enters (CR 122.1h).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "50994df7-300d-4e30-897a-0974dc0820ee",
		Name:         "Relentless X-ATM092",
		Completeness: CompletenessFull,
		BlockRules: []game.BlockRule{
			MinBlockers(OnSelf(), 3),
		},
		Activated: []ActivatedAbility{{
			Label:  "{8}: Return this card from your graveyard to the battlefield tapped with a finality counter on it.",
			Cost:   ManaCost("{8}"),
			Zones:  []game.ZoneKind{game.ZoneGraveyard},
			Effect: returnThisFromGraveyardTappedWithAFinalityCounter,
		}},
	})
}
