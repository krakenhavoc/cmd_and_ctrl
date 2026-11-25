package cardss

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sequestered Stash — Land (Kaladesh, S82):
//
//	"{T}: Add {C}.
//	 {4}, {T}, Sacrifice this land: Mill five cards. Then you may put an
//	 artifact card from your graveyard on top of your library."
//
// The mill is one keyed step; its tail then offers every artifact card in
// the graveyard, not just the milled ones, so an artifact that was already
// there is a legal pick (CR 400.7 for the live re-check). The put goes
// through the shared tuck exit, so a commander gets the CR 903.9 offer.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b6fe779f-b20d-49cc-96dd-54f1ffb312e1",
		Name:         "Sequestered Stash",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}, {T}, Sacrifice this land: Mill five cards. Then you may put an artifact card from your graveyard on top of your library.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{4}"), effects.TapCost(), effects.SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return g.MillToZoneKeyedForEffect(ctx.Controller(), 5, game.ZoneGraveyard, game.MillUntil{},
					ctx.ItemTail(sequesteredStashMilled))
			},
		}},
	})
}

// sequesteredStashMilled is the mill's keyed "then": an optional pick among
// the artifact cards in the graveyard. An on-disk key: never renamed, never
// reused.
var sequesteredStashMilled = effects.CardTail("mill-then/sequestered-stash", func(ctx *effects.Context, _ game.TailOutcome) error {
	p := ctx.Game.PlayerByIDForEffect(ctx.Controller())
	if p == nil || p.Graveyard == nil {
		return nil
	}
	var cands []uuid.UUID
	for _, c := range p.Graveyard.Cards {
		if c.IsArtifact() {
			cands = append(cands, c.InstanceID)
		}
	}
	if len(cands) == 0 {
		return nil
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  ctx.Controller(),
		Source:   ctx.Source(),
		Question: "Sequestered Stash — you may put an artifact card from your graveyard on top of your library",
		Cards:    cands,
		Min:      0,
		Max:      1,
		LiveZone: game.ZoneGraveyard,
		ThenKey:  sequesteredStashPick,
		Asked:    game.NewAskedItem(ctx.Item),
	})
	return nil
})

var sequesteredStashPick = effects.AskedPick("choose-cards/sequestered-stash", func(ctx *effects.Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	return ctx.Game.TuckToLibraryForEffect(r.Picked[0], false)
})
