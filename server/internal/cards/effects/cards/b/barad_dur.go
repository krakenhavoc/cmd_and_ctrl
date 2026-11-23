package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "88159872-d37d-4847-b048-e4a9af6437bd",
		Name:         "Barad-dûr",
		XMatters:     true,
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTappedUnless(effects.B08ControlsLegendaryCreature)},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{B}",
			Label:    "Add {B}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:     "{X}{X}{B}, {T}: Amass Orcs X. Activate only if a creature died this turn.",
			Purpose:   game.Purpose{Answers: game.AnswerPump | game.AnswerMakesBlocker},
			Cost:      effects.Plus(effects.ManaCost("{X}{X}{B}"), effects.TapCost()),
			Condition: effects.ACreatureDiedThisTurn(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.Amass{Subtype: "Orc", N: ctx.X()}.Apply(ctx)
			},
		}},
	})
}
