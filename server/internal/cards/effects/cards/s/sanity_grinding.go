package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sanity Grinding — {U}{U}{U} Sorcery:
//
//	"Chroma — Reveal the top ten cards of your library. For each blue mana
//	 symbol in the mana costs of the revealed cards, target opponent mills a
//	 card. Then put the cards you revealed this way on the bottom of your
//	 library in any order."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bfa0ec94-8851-4518-92e7-ab9f1f90a051",
		Name:         "Sanity Grinding",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			g := ctx.Game
			you := ctx.Controller()
			revealed := g.RevealTopOfLibraryForEffect(you, ctx.Source(), 10, "Sanity Grinding — reveal the top ten cards")
			color := ctx.TextColor("U")
			n := 0
			for _, id := range revealed {
				if c, ok := g.LookupCardForEffect(id); ok {
					n += effects.S82OldB521Symbols(c.ManaCost, color)
				}
			}
			if victim, ok := effects.FirstLegalPlayerTarget(ctx); ok && n > 0 {
				if err := (effects.MillCards{Player: victim, N: n}).Apply(ctx); err != nil {
					return err
				}
			}
			return g.PutInLibraryInChosenOrderThenForEffect(game.PutInLibrarySpec{
				Chooser:   you,
				Source:    ctx.Source(),
				Cards:     revealed,
				From:      game.ZoneLibrary,
				Placement: game.LibraryPlaceBottom,
				Reason:    "Sanity Grinding — put the revealed cards on the bottom of your library in any order",
			})
		},
	})
}
