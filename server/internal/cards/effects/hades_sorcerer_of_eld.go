package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hades, Sorcerer of Eld — the back face of Emet-Selch, Unsundered (Final
// Fantasy, #4462): Legendary Creature — Avatar, 6/6.
//
//	"Vigilance
//	 Echo of the Lost — During your turn, you may play cards from your
//	 graveyard.
//	 If a card or token would be put into your graveyard from anywhere,
//	 exile it instead."
//
// Echo of the Lost is a standing play permission over the controller's
// graveyard, gated to their own turn (Azula's DuringYourTurn): lands and
// spells both, at their usual cost and timing. The replacement is Forbidden
// Crypt's graveyard-to-exile over the controller's graveyard (the owner's,
// CR 404.1), cards AND tokens, so a dying token is exiled instead and
// "dies" abilities of permanents the controller owns don't trigger (CR
// 700.4); a discarded card is exiled, not put into the graveyard.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        emetSelchUnsunderedOracleID + "#1",
		Name:            "Hades, Sorcerer of Eld",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		GatedCastPermissions: []game.CastPermissionGate{{
			Condition: DuringYourTurn(),
			Permission: game.CastPermission{
				Zone:  game.ZoneGraveyard,
				Label: "Echo of the Lost — during your turn, play cards from your graveyard (Hades, Sorcerer of Eld)",
			},
		}},
		Replacements: []game.ReplacementEffect{
			GraveyardBecomesExile{
				YoursOnly: true,
				Label:     "Hades, Sorcerer of Eld: exile instead of your graveyard",
			}.Build(),
		},
	})
}
