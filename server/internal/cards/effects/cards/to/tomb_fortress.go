package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tomb Fortress — Land (Warhammer 40,000 Commander, #5948):
//
//	"This land enters tapped.
//	 {T}: Add {B}.
//	 {2}{B}{B}{B}, {T}, Exile this land: Mill four cards, then return a
//	 creature card from your graveyard to the battlefield. Activate only as
//	 a sorcery."
//
// Port of Karfell's shape with an exile cost, an untapped return and no
// blue. The pick waits for the mill (MillToZone.ThenTail) and offers every
// creature card in the graveyard, not only the milled ones.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "986f510c-e2ec-423e-a443-51a169939558",
		Name:         "Tomb Fortress",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{B}",
			Label:    "Add {B}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:        "{2}{B}{B}{B}, {T}, Exile this land: Mill four cards, then return a creature card from your graveyard to the battlefield. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{2}{B}{B}{B}"), effects.TapCost(), effects.ExileThis()),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.MillToZone{Player: item.Controller, N: 4, ThenTail: ctx.ItemTail(tombFortressMilled)}.Apply(ctx)
			},
		}},
	})
}

// tombFortressMilled is the clause after the mill: "return a creature card
// from your graveyard to the battlefield". An on-disk key.
var tombFortressMilled = effects.CardTail("tomb-fortress/after-mill", func(ctx *effects.Context, _ game.TailOutcome) error {
	item := ctx.Item
	candidates := effects.GraveyardCardIDs(ctx, item.Controller, game.Card.IsCreature)
	if len(candidates) == 0 {
		return nil
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    item.Controller,
		FromPlayer: item.Controller,
		Source:     item.SourceCardID,
		Question:   "Tomb Fortress — return a creature card from your graveyard to the battlefield",
		Cards:      candidates,
		Min:        1,
		Max:        1,
		ThenKey:    tombFortressReturn,
	})
	return nil
})

// tombFortressReturn is the keyed answer to "which creature card". An
// on-disk key: never renamed, never reused.
var tombFortressReturn = game.RegisterChooseCardsThen("choose-cards/tomb-fortress-return", func(g *game.Game, r game.CardsChosen) error {
	for _, id := range r.Picked {
		z := g.FindCardZoneForEffect(id)
		if z == nil || z.Kind != game.ZoneGraveyard || z.Owner != r.Chooser {
			continue
		}
		if c, ok := g.LookupCardForEffect(id); !ok || !c.IsCreature() {
			continue
		}
		if err := effects.NothingIfGone(g.ReturnFromGraveyardTappedForEffect(id, game.ZoneBattlefield, r.Chooser, false)); err != nil {
			return err
		}
	}
	return nil
})
