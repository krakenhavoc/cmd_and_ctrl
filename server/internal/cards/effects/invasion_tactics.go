package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Invasion Tactics — Enchantment {4}{G}:
//
//	"When this enchantment enters, creatures you control get +2/+2 until
//	 end of turn.
//	 Whenever one or more Allies you control deal combat damage to a
//	 player, draw a card."
//
// The enters pump covers the creatures its controller controls as it
// resolves (CR 611.2c): one that enters later doesn't get it. The draw is
// Captain Janeway, Endgame's batch trigger with Ally in place of Officer:
// one draw per player the Allies connect with in a combat damage step
// (CR 603.2c), the Ally type read post-layer off the damage source.
//
// No simplification. Avatar: The Last Airbender (#4459).
func init() {
	Register(Spec{
		OracleID:     "eef63cb4-96de-4c9d-84aa-62a3b52a3c7f",
		Name:         "Invasion Tactics",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Invasion Tactics — creatures you control get +2/+2 until end of turn",
				Do(BoostUntilEOT{Match: And(Creature(), YouControl()), Power: 2, Toughness: 2,
					Label: "Invasion Tactics — creatures you control get +2/+2"})),
			WheneverOneOrMoreCreaturesYouControlDealCombatDamageToAPlayer(HasSubtype("Ally"),
				"Invasion Tactics — draw a card (Allies you control dealt combat damage to a player)",
				Do(DrawCards{N: 1})),
		},
	})
}
