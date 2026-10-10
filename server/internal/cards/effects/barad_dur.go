package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Barad-dûr — Legendary Land:
//
//	"Barad-dûr enters tapped unless you control a legendary creature.
//	 {T}: Add {B}.
//	 {X}{X}{B}, {T}: Amass Orcs X. Activate only if a creature died
//	 this turn."
//
// The Shire's and Rivendell's legendary-creature entry and a plain {B}.
// The activation is Treasure Vault's two X slots (so X=2 costs five
// mana) with the X read back off the stack item, feeding the amass; it
// carries an activation condition on the table-wide creatures-died tally,
// checked at announcement. An X of 0 still amasses 0 and so
// makes the 0/0 Army, which dies at once, as the rules have it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "88159872-d37d-4847-b048-e4a9af6437bd",
		Name:         "Barad-dûr",
		XMatters:     true,
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTappedUnless(b08ControlsLegendaryCreature)},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{B}",
			Label:    "Add {B}",
		}},
		Activated: []ActivatedAbility{{
			Label:     "{X}{X}{B}, {T}: Amass Orcs X. Activate only if a creature died this turn.",
			Purpose:   game.Purpose{Answers: game.AnswerPump | game.AnswerMakesBlocker},
			Cost:      Plus(ManaCost("{X}{X}{B}"), TapCost()),
			Condition: ACreatureDiedThisTurn(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return Amass{Subtype: "Orc", N: ctx.X()}.Apply(ctx)
			},
		}},
	})
}
