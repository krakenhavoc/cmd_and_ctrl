package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Helga, Skittish Seer — Legendary Creature — Frog Druid {G}{W}{U}, 1/3
// (Bloomburrow, #4467):
//
//	"Whenever you cast a creature spell with mana value 4 or greater, you
//	 draw a card, gain 1 life, and put a +1/+1 counter on Helga.
//	 {T}: Add X mana of any one color, where X is Helga's power. Spend
//	 this mana only to cast creature spells with mana value 4 or greater
//	 or creature spells with {X} in their mana costs."
//
// The cast trigger is Jin-Gitaxias's shape over creature spells: the
// mana value is the spell's on the stack, {X} at the value announced
// (CR 202.3e). The draw, the life and the counter come in printed order.
//
// The mana is Woodland Weavemaster's (X of one colour, X the power as
// the ability resolves) with Troyan, Gutsy Explorer's restriction made
// creature-only: one tag that is an OR (CR 106.6) of a creature spell
// with mana value 4 or greater and a creature spell with {X} in its
// printed mana cost. Neither alternative pays for an activated ability.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0a0d25a8-4f28-453b-9270-4d449fe0ca65",
		Name:         "Helga, Skittish Seer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(And(Creature(), ManaValueGE(4)),
				"Helga, Skittish Seer — draw a card, gain 1 life and put a +1/+1 counter on Helga",
				Do(DrawCards{N: 1}, GainLife{Amount: 1}, CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
		ManaAbilities: []ManaAbility{{
			Cost:           ManaAbilityCost{Tap: true},
			ProducedAmount: ProducedOneColor(SourcePower),
			Label:          "{T}: Add X mana of any one color, where X is Helga's power. Spend this mana only to cast creature spells with mana value 4 or greater or creature spells with {X} in their mana costs.",
			Restrictions: []string{game.ManaRestrictAnyOf(
				[]string{game.ManaRestrictCast, game.ManaRestrictType("Creature"), game.ManaRestrictManaValueAtLeast(4)},
				[]string{game.ManaRestrictCast, game.ManaRestrictType("Creature"), game.ManaRestrictSpellXCost},
			)},
		}},
	})
}
