package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
//     exiled. A Cerberus commander may take CR 903.9's offer instead of
//     exile, which pauses the trigger, so the return runs in the
//     exile's continuation.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a36c2b76-d595-42ba-b8c2-bb8f02639981",
		Name:         "Underworld Cerberus",
		Completeness: CompletenessFull,
		BlockRules: []game.BlockRule{
			MinBlockers(OnSelf(), 3),
		},
		TargetingRestrictions: []game.TargetingRestriction{
			CardsInGraveyardsCantBeTargeted("Cards in graveyards can't be the targets of spells or abilities."),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Underworld Cerberus — exile it; each player returns all creature cards from their graveyard to their hand",
				underworldCerberusDies),
		},
	})
}

// underworldCerberusDies exiles the Cerberus from the graveyard, then
// returns every creature card in every graveyard to its owner's hand.
func underworldCerberusDies(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if z := g.FindCardZoneForEffect(item.SourceCardID); z == nil || z.Kind != game.ZoneGraveyard {
		return eachPlayerReturnsCreatureCardsToHand(ctx)
	}
	return ExileTarget{
		Target: item.SourceCardID,
		Then: func(ctx *Context, _ bool) error {
			return eachPlayerReturnsCreatureCardsToHand(ctx)
		},
	}.Apply(ctx)
}

// eachPlayerReturnsCreatureCardsToHand is "each player returns all
// creature cards from their graveyard to their hand": every creature
// card in every graveyard, read once, to its owner's hand as one move.
func eachPlayerReturnsCreatureCardsToHand(ctx *Context) error {
	ids := allGraveyardsCardIDs(ctx, game.Card.IsCreature)
	if len(ids) == 0 {
		return nil
	}
	ctx.Game.BounceCardsToHandForEffect(ids)
	return nil
}
