package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Collector Ouphe — Creature — Ouphe {1}{G}, 2/2:
//
//	"Activated abilities of artifacts can't be activated."
//
// The green Null Rod, and the one printed on a body you can attack
// with. Same clause as Cursed Totem with the type swapped, and the
// same missing "unless they're mana abilities": a Sol Ring, a Mana
// Crypt, an Arcane Signet and every Talisman stop producing while
// this is on the battlefield, which is what separates the Ouphe from
// Stony Silence's usual comparison and makes it a real Commander
// card rather than a sideboard one.
//
// (Null Rod and Stony Silence, when they arrive, are the SAME
// constructor with `exemptMana` true — that argument exists so the
// difference is a word in the card file and not a second mechanism.)
//
// Symmetrical: its controller's Sol Ring is off too. "Artifacts" is
// the layered type line (ADR 0039), so an artifact a type-changing
// effect has stopped being one starts producing again.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0c4bc9ea-a5fd-4f44-96a1-5448eee228c4",
		Name:         "Collector Ouphe",
		Completeness: CompletenessFull,
		ActivationRestrictions: []game.ActivationRestriction{
			SourcesCantActivate(
				"Collector Ouphe — activated abilities of artifacts can't be activated.",
				Artifact(), false),
		},
	})
}
