package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Port of Karfell — Land:
//
//	"This land enters tapped.
//	 {T}: Add {U}.
//	 {3}{U}{B}{B}, {T}, Sacrifice this land: Mill four cards, then return
//	 a creature card from your graveyard to the battlefield tapped. (To
//	 mill a card, put the top card of your library into your graveyard.)"
//
// The mill resolves first and the creature is chosen AFTER it, as the
// ability resolves (CR 608.2c), so a creature the mill just put into the
// graveyard is a legal pick. The pick is queued from MillToZone.Then, not
// the next line, so a replacement that stops the mill to ask something
// (a commander's command-zone offer) finishes it first. The pick is mandatory ("return a creature
// card"): with no creature card in the graveyard after the mill the
// ability does nothing more and no prompt is asked. The land is
// sacrificed as a COST, so the ability does not depend on it.
//
// The pick is a KEYED choose-cards prompt (portOfKarfellReturn), so a
// table waiting on it is still a restore point (ADR 0041). The creature
// enters tapped through the CR 614 entry, not a tap after it. The
// continuation re-checks the pick is still a creature card in the
// controller's graveyard, since it runs after the prompt was asked.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f500a8a4-6135-448c-9116-bd2695a72229",
		Name:         "Port of Karfell",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{U}",
			Label:    "Add {U}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{3}{U}{B}{B}, {T}, Sacrifice this land: Mill four cards, then return a creature card from your graveyard to the battlefield tapped.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{3}{U}{B}{B}"), TapCost(), SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				// The pick waits for the mill (MillToZone.Then): a replacement
				// that stops to ask during it must not let the pick be built
				// from a graveyard the mill has not finished filling.
				return MillToZone{Player: item.Controller, N: 4, Then: portOfKarfellChoose}.Apply(NewContext(g, item))
			},
		}},
	})
}

// portOfKarfellChoose is "then return a creature card from your
// graveyard to the battlefield tapped", asked once the mill is done. The
// candidates are every creature card in the graveyard, not only the
// milled ones: the clause says "your graveyard".
func portOfKarfellChoose(ctx *Context, _ []uuid.UUID) error {
	item := ctx.Item
	candidates := graveyardCardIDs(ctx, item.Controller, game.Card.IsCreature)
	if len(candidates) == 0 {
		return nil
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    item.Controller,
		FromPlayer: item.Controller,
		Source:     item.SourceCardID,
		Question:   "Port of Karfell — return a creature card from your graveyard to the battlefield tapped",
		Cards:      candidates,
		Min:        1,
		Max:        1,
		ThenKey:    portOfKarfellReturn,
	})
	return nil
}

// portOfKarfellReturn is the keyed answer to "which creature card". Its
// key is an on-disk identity: never renamed, never reused.
var portOfKarfellReturn = game.RegisterChooseCardsThen("choose-cards/port-of-karfell-return", func(g *game.Game, r game.CardsChosen) error {
	for _, id := range r.Picked {
		z := g.FindCardZoneForEffect(id)
		if z == nil || z.Kind != game.ZoneGraveyard || z.Owner != r.Chooser {
			continue
		}
		if c, ok := g.LookupCardForEffect(id); !ok || !c.IsCreature() {
			continue
		}
		if err := nothingIfGone(g.ReturnFromGraveyardTappedForEffect(id, game.ZoneBattlefield, r.Chooser, true)); err != nil {
			return err
		}
	}
	return nil
})
