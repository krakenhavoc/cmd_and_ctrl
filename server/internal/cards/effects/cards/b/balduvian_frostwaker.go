package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Balduvian Frostwaker — Creature — Human Wizard {2}{U}, 1/1 (Coldsnap):
//
//	"{U}, {T}: Target snow land becomes a 2/2 blue Elemental creature with
//	 flying. It's still a land."
//
// No duration is printed, so the land stays a 2/2 blue Elemental creature
// with flying for as long as it is on the battlefield (CR 611.2a, CR 611.2c:
// it follows that one object, not a later one that left and came back,
// CR 400.7). The land keeps its land type, subtypes and mana ability
// (CR 205.1b); the engine refuses to tap it for mana while it is a
// summoning-sick creature (CR 302.6). The target must be a snow land as the
// ability is activated and again as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2cff1f9e-11a6-460c-8f5f-4c02424357e3",
		Name:         "Balduvian Frostwaker",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}, {T}: Target snow land becomes a 2/2 blue Elemental creature with flying. It's still a land.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    effects.Plus(effects.ManaCost("{U}"), effects.TapCost()),
			Targets: effects.TargetPermanent("target snow land", effects.Land(), effects.Snow()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.S82OldB561BecomesElemental(ctx, id)
			},
		}},
	})
}
