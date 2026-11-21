package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

const lunarchMantleGrant = "lunarch-mantle/flying"

// Lunarch Mantle — Enchantment — Aura {1}{W} (Innistrad Remastered):
//
//	"Enchant creature
//	 Enchanted creature gets +2/+2 and has "{1}, Sacrifice a permanent: This
//	 creature gains flying until end of turn.""
//
// The granted ability is the creature's, so "this creature" is the host and
// its controller pays, whoever controls the Aura.
func init() {
	Register(Spec{
		OracleID:     "0da104ae-c46c-4106-919f-d975721d42d9",
		Name:         "Lunarch Mantle",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Grants: []AbilityGrant{{
			Key: lunarchMantleGrant,
			Activated: []ActivatedAbility{{
				Label:   "{1}, Sacrifice a permanent: This creature gains flying until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerCombatGrant | game.AnswerSacOutlet},
				Cost:    Plus(ManaCost("{1}"), SacrificeAPermanent()),
				Effect:  thisGainsKeywordUntilEndOfTurn("flying", "Lunarch Mantle — flying until end of turn"),
			}},
			Text: "{1}, Sacrifice a permanent: This creature gains flying until end of turn.",
		}},
		Static: []game.StaticAbility{PumpAttached(2, 2), GrantAbilitiesToAttached(lunarchMantleGrant)},
	})
}
