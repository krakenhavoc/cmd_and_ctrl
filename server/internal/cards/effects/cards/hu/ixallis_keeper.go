package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ixalli's Keeper — Creature — Human Shaman {1}{G}, 2/2 (Ixalan):
//
//	"{7}{G}, {T}, Sacrifice this creature: Target creature gets +5/+5 and
//	 gains trample until end of turn."
//
// Priest of the Haunted Edge's cost shape (the tap and the sacrifice are
// costs, so it can't be activated the turn it arrives, CR 302.6, and is gone
// whatever happens to the target). The pump and the trample are one record
// on the target (CR 611.2c); a target that has left does nothing.
//
// No simplification.
func init() {
	const label = "Ixalli's Keeper — target creature gets +5/+5 and gains trample until end of turn"
	effects.Register(effects.Spec{
		OracleID:     "183e2796-2fc1-4f0e-b1be-95285d2ea046",
		Name:         "Ixalli's Keeper",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{7}{G}, {T}, Sacrifice this creature: Target creature gets +5/+5 and gains trample until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 5, 5, "trample")),
			Cost:    effects.Plus(effects.ManaCost("{7}{G}"), effects.TapCost(), effects.SacrificeThis()),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.UntilEndOfTurn(ctx, id, nil, label, game.ModifyPTMod(5, 5), game.AddKeywordsMod("trample"))
			},
		}},
	})
}
