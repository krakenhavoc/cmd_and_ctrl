package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fallen Ideal — Enchantment — Aura {2}{B} (Secrets of Strixhaven
// Commander, #4834):
//
//	"Enchant creature
//	 Enchanted creature has flying and "Sacrifice a creature: This
//	 creature gets +2/+1 until end of turn."
//	 When this Aura is put into a graveyard from the battlefield, return
//	 it to its owner's hand."
//
// Flying is a layer 6 grant; the sacrifice ability is an ADR 0093
// bundle the Aura gives its creature, so it is the creature's own
// ability: the creature's controller activates it, not the Aura's (the
// ruling), and "this creature" is the creature. The creature may
// sacrifice itself; the pump then finds nothing on the battlefield.
//
// The return is Eye of Nidhogg's and Rancor's: any trip from the
// battlefield to a graveyard, moving the card only while it is still the
// object that went there.
//
// No simplification.
const fallenIdealGrant = "fallen-ideal/sacrifice-pump" // an on-disk identity: never renamed

func init() {
	Register(Spec{
		OracleID:     "4b8183e6-4ff2-4254-9ee0-0ebbf8cef4a3",
		Name:         "Fallen Ideal",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Grants: []AbilityGrant{{
			Key: fallenIdealGrant,
			Activated: []ActivatedAbility{{
				Label:   "Sacrifice a creature: This creature gets +2/+1 until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerSacOutlet},
				Cost:    SacrificeACreature(),
				Effect:  thisGetsUntilEndOfTurn(2, 1, "Fallen Ideal — +2/+1 until end of turn"),
			}},
			Text: "Sacrifice a creature: This creature gets +2/+1 until end of turn.",
		}},
		Static: []game.StaticAbility{
			GrantToAttached("flying"),
			GrantAbilitiesToAttached(fallenIdealGrant),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, ThisDied, "Fallen Ideal — return it to its owner's hand", returnSelfFromGraveyardToHand),
		},
	})
}
