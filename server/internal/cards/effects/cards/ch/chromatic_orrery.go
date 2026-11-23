package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Chromatic Orrery — Legendary Artifact {7}:
//
//	"You may spend mana as though it were mana of any color.
//	 {T}: Add {C}{C}{C}{C}{C}.
//	 {5}, {T}: Draw a card for each color among permanents you
//	 control."
//
// The first line is a player static (CR 609.4b, #1600): the engine
// reads it at every payment its controller makes — spells, activated
// and mana abilities, special actions, attack taxes and "unless pays"
// costs — and lets any mana, the Orrery's own colourless included, pay
// a coloured symbol. It changes how a cost is paid, not the cost: the
// prices on the client stay printed, a {C} still needs colourless mana
// (CR 106.1b), and "spend this mana only on …" still binds. See
// game/spend_any_color.go.
//
// The draw counts colours among the controller's permanents as the
// ability resolves; colourless is not a colour, so an artifact board
// draws nothing.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "95c3976c-33f3-490b-bfd3-7f1af2fe0416",
		Name:          "Chromatic Orrery",
		Completeness:  effects.CompletenessFull,
		AnyColorSpend: effects.YouMaySpendManaAsAnyColor(),
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}{C}{C}{C}{C}",
			Label:    "Add {C}{C}{C}{C}{C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{5}, {T}: Draw a card for each color among permanents you control",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{5}"), effects.TapCost()),
			Effect:  ChromaticOrreryDraw,
		}},
	})
}

// chromaticOrreryDraw counts the distinct colours among the
// controller's permanents and draws that many cards.
func ChromaticOrreryDraw(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	controller := ctx.Controller()
	colors := map[string]bool{}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != controller {
			continue
		}
		for _, col := range c.EffectiveColors() {
			colors[col] = true
		}
	}
	if len(colors) == 0 {
		return nil
	}
	return effects.DrawCards{Player: controller, N: len(colors)}.Apply(ctx)
}
