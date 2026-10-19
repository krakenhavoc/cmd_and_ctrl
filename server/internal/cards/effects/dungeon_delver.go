package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dungeon Delver — Legendary Enchantment — Background {1}{U}:
//
//	"Commander creatures you own have "Room abilities of dungeons you own
//	 trigger an additional time.""
//
// An ADR 0093 grant to each commander creature you own
// (choose_a_background.go), carrying a CR 603.2d trigger doubler
// (#4137, ADR 0018's trigger doubling addendum, amended 2026-10-19). Each
// commander creature is its own doubler, so two of them (partners) are two
// additional triggers; the doubler is gone from a commander that loses its
// abilities, and a commander an opponent has stolen doubles the thief's
// dungeons ("you" is its controller), as Hama Pashar's would.
//
// No simplification.
const dungeonDelverGrant = "dungeon-delver/room-doubler"

func init() {
	Register(Spec{
		OracleID:     "0073d731-e60d-433a-ac35-a22af855da20",
		Name:         "Dungeon Delver",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key:             dungeonDelverGrant,
			TriggerDoublers: []game.TriggerDoubler{roomAbilitiesOfDungeonsYouOwnDoubler("Dungeon Delver — room abilities trigger an additional time")},
			Text:            "Room abilities of dungeons you own trigger an additional time.",
		}},
		Static: []game.StaticAbility{grantToCommanderCreaturesYouOwn(dungeonDelverGrant)},
	})
}
