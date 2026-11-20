package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Underworld Cerberus — Creature — Dog {3}{B}{R}, 6/6:
//
//	"This creature can't be blocked except by three or more creatures.
//	 Cards in graveyards can't be the targets of spells or abilities.
//	 When this creature dies, exile it and each player returns all
//	 creature cards from their graveyard to their hand."
//
// Three clauses, one shape each:
//
//   - THE BLOCK RULE is Rampaging Ceratops's MinBlockers(OnSelf(), 3)
//     (CR 509.1b).
//   - THE STATIC is ADR 0109 §6's TargetingRestrictions (#1885), read at
//     the engine's two targeting choke points: no spell or ability may
//     target a card in any graveyard while the Cerberus is on the
//     battlefield (CR 601.2c, 608.2b).
//   - THE DIES TRIGGER exiles the Cerberus first, if it is still in the
//     graveyard (a Cerberus that has left it is a new object, CR 400.7,
//     and stays where it went), so it is not among the creature cards
//     returned. Then every creature card in every graveyard goes to its
//     owner's hand as one simultaneous move. The return does not depend
//     on the exile: it happens even when the Cerberus could not be
//     exiled. A Cerberus commander dies first and its owner is offered
//     the command zone afterwards (CR 903.9a, ADR 0115); a "yes" leaves
//     the trigger nothing to exile and the return still runs.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a36c2b76-d595-42ba-b8c2-bb8f02639981",
		Name:         "Underworld Cerberus",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{
			effects.MinBlockers(effects.OnSelf(), 3),
		},
		TargetingRestrictions: []game.TargetingRestriction{
			effects.CardsInGraveyardsCantBeTargeted("Cards in graveyards can't be the targets of spells or abilities."),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Underworld Cerberus — exile it; each player returns all creature cards from their graveyard to their hand",
				underworldCerberusDies),
		},
	})
}

// underworldCerberusDies exiles the Cerberus from the graveyard, then
// returns every creature card in every graveyard to its owner's hand.
func underworldCerberusDies(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if z := g.FindCardZoneForEffect(item.SourceCardID); z == nil || z.Kind != game.ZoneGraveyard {
		return eachPlayerReturnsCreatureCardsToHand(ctx)
	}
	return effects.ExileTarget{
		Target:   item.SourceCardID,
		ThenTail: ctx.ItemTail(underworldCerberusExiled),
	}.Apply(ctx)
}

// eachPlayerReturnsCreatureCardsToHand is "each player returns all
// creature cards from their graveyard to their hand": every creature
// card in every graveyard, read once, to its owner's hand as one move.
func eachPlayerReturnsCreatureCardsToHand(ctx *effects.Context) error {
	ids := effects.AllGraveyardsCardIDs(ctx, game.Card.IsCreature)
	if len(ids) == 0 {
		return nil
	}
	ctx.Game.BounceCardsToHandForEffect(ids)
	return nil
}

// underworldCerberusExiled is the clause after the exile (#4253 slice 9).
var underworldCerberusExiled = effects.CardTail("underworld-cerberus/after-exile", func(ctx *effects.Context, o game.TailOutcome) error {

	return eachPlayerReturnsCreatureCardsToHand(ctx)
})
