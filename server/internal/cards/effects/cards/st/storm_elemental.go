package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Storm Elemental — Creature — Elemental {5}{U}, 3/4:
//
//	"Flying
//	 {U}, Exile the top card of your library: Tap target creature with
//	 flying.
//	 {U}, Exile the top card of your library: If the exiled card is a
//	 snow land, this creature gets +1/+1 until end of turn."
//
// ADR 0109 §7 (#1902): "Exile the top N cards of your library" is a
// cost with nothing to choose. A library of fewer than N cards can't pay
// it (CR 118.3), and it is paid after every other cost (CR 601.2h).
// "The exiled card" is read off the payment record (Context.Exiled,
// CR 400.7j) as the second ability resolves: a land with the snow
// supertype (CR 205.4g) pumps the Elemental, anything else does nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f1fee486-660e-4356-896e-671e8b675ad8",
		Name:            "Storm Elemental",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{U}, Exile the top card of your library: Tap target creature with flying.",
				Cost:    effects.Plus(effects.ManaCost("{U}"), effects.ExileTopOfLibrary(1)),
				Targets: effects.TargetCreature("target creature with flying", effects.HasKeyword("flying")),
				Effect:  effects.TapTheFirstLegalTarget,
			},
			{
				Label:   "{U}, Exile the top card of your library: If the exiled card is a snow land, this creature gets +1/+1 until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    effects.Plus(effects.ManaCost("{U}"), effects.ExileTopOfLibrary(1)),
				Effect: func(g *game.Game, item *game.StackItem) error {
					for _, id := range effects.NewContext(g, item).Exiled() {
						if c, ok := g.LookupCardForEffect(id); ok && c.IsLand() && c.HasSupertype("snow") {
							return effects.ThisGetsUntilEndOfTurn(1, 1, "Storm Elemental — +1/+1")(g, item)
						}
					}
					return nil
				},
			},
		},
	})
}
