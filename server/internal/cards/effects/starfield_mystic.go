package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Starfield Mystic — Creature — Human Cleric {1}{W}, 2/2 (Secrets of
// Strixhaven Commander, #4834):
//
//	"Enchantment spells you cast cost {1} less to cast.
//	 Whenever an enchantment you control is put into a graveyard from the
//	 battlefield, put a +1/+1 counter on this creature."
//
// The discount is Etherium Sculptor's shape over enchantment spells
// (CostsLess, CR 601.2f); it reduces generic mana only. The trigger reads
// the enchantment as it last existed on the battlefield and under whose
// control (CR 603.10a), so an Aura that falls off with its creature counts,
// a token enchantment counts, and an enchantment an opponent controlled
// does not. The Mystic itself is not an enchantment, so it is never its own
// trigger. The counter goes on the Mystic only while it is still on the
// battlefield (CounterOnThis): when it is the enchantment's death that
// removed it too, nothing happens.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2975590d-6a5b-41e5-8a3b-e9325e71226a",
		Name:         "Starfield Mystic",
		Completeness: CompletenessFull,
		CostModifiers: []game.CostModifier{
			CostsLess(1, "Enchantment spells you cast cost {1} less to cast.", YourSpell(), socB13EnchantmentSpell()),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, socB13EnchantmentYouControlWasPutIntoAGraveyard,
				"Starfield Mystic — put a +1/+1 counter on this creature",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
