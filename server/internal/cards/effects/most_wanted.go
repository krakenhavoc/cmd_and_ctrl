package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Most Wanted — Enchantment — Aura {2}{G} (Streets of New Capenna):
//
//	"Flash
//	 Enchant creature
//	 Enchanted creature gets +2/+1.
//	 When enchanted creature dies, create two Treasure tokens."
//
// The Treasures are the Aura's controller's, not the dead creature's
// controller's. The trigger is a leaves-the-battlefield ability: it
// looks back at what the Aura was attached to as the creature died
// (CR 603.10a), so it still fires when the Aura goes to the graveyard
// with it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "140d38cc-0ae6-4e9b-a60f-a51d2947e07b",
		Name:            "Most Wanted",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Purpose:         game.Purpose{Tokens: 2},
		Targets:         EnchantCreature(),
		Static: []game.StaticAbility{
			PumpAttached(2, 1),
		},
		Triggered: []game.TriggeredAbility{
			WhenEnchantedCreatureDies("Most Wanted — create two Treasure tokens",
				Do(CreateToken{Template: TreasureToken(), N: 2})),
		},
	})
}
