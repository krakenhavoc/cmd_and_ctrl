package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Slumbering Keepguard — Creature — Human Knight {W}, 1/1 (Wilds of
// Eldraine, #4471):
//
//	"Whenever an enchantment you control enters, scry 1.
//	 {2}{W}: This creature gets +1/+1 until end of turn for each
//	 enchantment you control."
//
// The trigger is Killian's EnchantmentEnteredUnderYourControl. The pump
// counts the enchantments its controller controls as it resolves
// (b29EnchantmentsControlled, effective types), and a Keepguard that has
// left the battlefield gets nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "89078485-c438-4035-95dc-dbe37fb13e50",
		Name:         "Slumbering Keepguard",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Scry: 1},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, EnchantmentEnteredUnderYourControl,
				"Slumbering Keepguard — scry 1", Do(Scry{N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}{W}: This creature gets +1/+1 until end of turn for each enchantment you control.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{2}{W}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !onBattlefield(g, item.SourceCardID) || sourceIsNewObject(g, item) {
					return nil
				}
				n := b29EnchantmentsControlled(g, item.Controller)
				return BoostUntilEOT{Target: item.SourceCardID, Power: n, Toughness: n,
					Label: "Slumbering Keepguard — +1/+1 for each enchantment you control"}.Apply(NewContext(g, item))
			},
		}},
	})
}
