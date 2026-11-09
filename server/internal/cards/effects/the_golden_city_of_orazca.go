package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Golden City of Orazca — Plane — Ixalan (Commander Masters, #4870):
//
//	"Ascend (If you control ten or more permanents, you get the city's
//	 blessing for the rest of the game.)
//	 Whenever one or more creatures you control deal combat damage to a
//	 player, create a Treasure token. Then draw a card if you have the
//	 city's blessing.
//	 Whenever chaos ensues, you may put a permanent card from your hand
//	 onto the battlefield tapped."
//
// A plane (ADR 0216 §6). Ascend on the plane is the ability that functions
// from the command zone (CR 901.7): while it is face up the planar controller
// gets the city's blessing as soon as they control ten permanents (the
// engine's earned-designation sweep reads the face-up planes), and the
// ruling holds, so a player who controls ten permanents before the plane is
// found does not have it until then. The combat damage ability triggers once
// for each player the planar controller's creatures hit in a combat damage
// step (CR 603.2c); the card is drawn after the Treasure only with the
// blessing, read as it resolves. The chaos ability (CR 311.7) is a "you
// may" putting any permanent card from the hand onto the battlefield tapped.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3c200cc0-02b1-4a32-a910-5cd3c4d716b7",
		Name:            "The Golden City of Orazca",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordAscend},
		Triggered: []game.TriggeredAbility{
			WheneverOneOrMoreCreaturesYouControlDealCombatDamageToAPlayer(nil,
				"The Golden City of Orazca — create a Treasure token, then draw a card if you have the city's blessing",
				Do(CreateToken{Template: TreasureToken(), N: 1},
					If{Cond: CondCitysBlessing{}, Then: []Applier{DrawCards{N: 1}}})),
			WheneverChaosEnsues("The Golden City of Orazca — you may put a permanent card from your hand onto the battlefield tapped",
				Do(PutFromHandOntoBattlefield{
					Tapped:   true,
					Optional: true,
					Label:    "The Golden City of Orazca — put a permanent card from your hand onto the battlefield tapped",
				})),
		},
	})
}
