package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Demon of Fate's Design — Enchantment Creature — Demon {4}{B}{B}, 6/6
// (Duskmourn Commander):
//
//	"Flying, trample
//	 Once during each of your turns, you may cast an enchantment spell by
//	 paying life equal to its mana value rather than paying its mana cost.
//	 {2}{B}, Sacrifice another enchantment: This creature gets +X/+0 until
//	 end of turn, where X is the sacrificed enchantment's mana value."
//
// The first is Access Maze's granted offer, narrowed to enchantment spells
// (cmdr2S43PayLifeForAnEnchantmentSpellOnceEachOfYourTurns): a cost the caster
// must have the life for, once per turn per Demon, on your turn only, from any
// zone. The pump reads the sacrificed enchantment's mana value as it last
// existed (CR 608.2h), and does nothing if the Demon is gone.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "53716e0b-4d4e-40b0-8370-38e11f73d195",
		Name:            "Demon of Fate's Design",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "trample"},
		GrantedAlternativeCosts: []game.GrantedAlternativeCost{
			cmdr2S43PayLifeForAnEnchantmentSpellOnceEachOfYourTurns(),
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}{B}, Sacrifice another enchantment: This creature gets +X/+0 until end of turn, where X is the sacrificed enchantment's mana value",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    Plus(ManaCost("{2}{B}"), SacrificeAnotherN(1, "another enchantment", Enchantment())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				x := sacrificedManaValue(ctx)
				if x <= 0 || !b15OnBattlefield(g, item.SourceCardID) {
					return nil
				}
				return BoostUntilEOT{Target: item.SourceCardID, Power: x, Label: "Demon of Fate's Design — +X/+0"}.Apply(ctx)
			},
		}},
	})
}
