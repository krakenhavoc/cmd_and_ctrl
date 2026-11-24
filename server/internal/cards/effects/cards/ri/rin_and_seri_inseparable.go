package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rin and Seri, Inseparable — Legendary Creature — Dog Cat {1}{R}{G}{W}, 4/4
// (Secret Lair Drop):
//
//	"Whenever you cast a Dog spell, create a 1/1 green Cat creature token.
//	 Whenever you cast a Cat spell, create a 1/1 white Dog creature token.
//	 {R}{G}{W}, {T}: Rin and Seri deals damage to any target equal to the
//	 number of Dogs you control. You gain life equal to the number of Cats you
//	 control."
//
// The two cast triggers read the spell's creature types (a Dog Cat spell
// triggers both). The activated ability counts as it resolves; Rin and Seri
// is a Dog and a Cat herself. The life is gained even if the target became
// illegal only when the whole ability fizzles (CR 608.2b), as for any
// targeted ability.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "28c46eb7-bdfc-44b9-b7e9-b073260851f0",
		Name:         "Rin and Seri, Inseparable",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.OfCreatureType("Dog"), "Rin and Seri, Inseparable — create a 1/1 green Cat creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 green Cat"), N: 1})),
			effects.WheneverYouCast(effects.OfCreatureType("Cat"), "Rin and Seri, Inseparable — create a 1/1 white Dog creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Dog"), N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}{G}{W}, {T}: Rin and Seri deals damage to any target equal to the number of Dogs you control. You gain life equal to the number of Cats you control.",
			Cost:    effects.Plus(effects.ManaCost("{R}{G}{W}"), effects.TapCost()),
			Targets: effects.TargetAny(),
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if ts := ctx.LegalTargets(); len(ts) > 0 {
					if err := (effects.DealDamage{Source: item.SourceCardID, Target: ts[0].ID, Amount: effects.OldB10CountSubtype(g, item.Controller, "Dog")}).Apply(ctx); err != nil {
						return err
					}
				}
				return effects.GainLife{Player: item.Controller, Amount: effects.OldB10CountSubtype(g, item.Controller, "Cat")}.Apply(ctx)
			},
		}},
	})
}
