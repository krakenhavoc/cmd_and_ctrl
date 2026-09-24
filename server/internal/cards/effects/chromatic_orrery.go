package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Chromatic Orrery — Legendary Artifact {7}:
//
//	"You may spend mana as though it were mana of any color.
//	 {T}: Add {C}{C}{C}{C}{C}.
//	 {5}, {T}: Draw a card for each color among permanents you
//	 control."
//
// The "spend as any color" static is the same clause Mycosynth
// Lattice already declares unimplemented (its own caveat: "you cannot
// spend mana as any color") — no engine reader exists for it yet, so
// this card ships the same gap rather than a new one. The two tap
// abilities are ordinary.
//
// Caveat: "you may spend mana as though it were mana of any color"
// isn't implemented. The two tap abilities work.
func init() {
	Register(Spec{
		OracleID:     "95c3976c-33f3-490b-bfd3-7f1af2fe0416",
		Name:         "Chromatic Orrery",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Spending mana as though it were mana of any color isn't implemented — only the two tap abilities work."},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}{C}{C}{C}{C}",
			Label:    "Add {C}{C}{C}{C}{C}",
		}},
		Activated: []ActivatedAbility{{
			Label:  "{5}, {T}: Draw a card for each color among permanents you control",
			Cost:   Plus(ManaCost("{5}"), TapCost()),
			Effect: chromaticOrreryDraw,
		}},
	})
}

// chromaticOrreryDraw counts the distinct colours among the
// controller's permanents and draws that many cards.
func chromaticOrreryDraw(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
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
	return DrawCards{Player: controller, N: len(colors)}.Apply(ctx)
}
