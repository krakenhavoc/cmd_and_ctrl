package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Markov Warlord — Creature — Vampire Warrior {5}{R}, 4/4:
//
//	"Haste
//	 When this creature enters, up to two target creatures can't block this
//	 turn."
//
// Beat a Path's body: each chosen creature still a legal target at
// resolution gains the can't-block restriction until end of turn
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2279005d-5434-49ce-b16a-212759c5775a",
		Name:            "Markov Warlord",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Markov Warlord — up to two target creatures can't block this turn",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						for _, t := range ctx.LegalTargets() {
							if t.Kind != game.TargetCard {
								continue
							}
							if err := (effects.RestrictUntilEOT{Target: t.ID, Restrictions: game.CantBlock, Label: "Markov Warlord — can't block"}).Apply(ctx); err != nil {
								return err
							}
						}
						return nil
					}),
				effects.TargetCreature("up to two target creatures").WithCount(0, 2),
			),
		},
	})
}
