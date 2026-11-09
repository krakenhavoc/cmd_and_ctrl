package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Weftwalking — Enchantment {4}{U}{U} (Edge of Eternities, #4461):
//
//	"When this enchantment enters, if you cast it, shuffle your hand and
//	 graveyard into your library, then draw seven cards.
//	 The first spell each player casts during each of their turns may be
//	 cast without paying its mana cost."
//
// The enters trigger is Echo of Eons' shuffle-then-draw over just you, with
// an intervening "if you cast it" (CR 603.4): asked as the permanent enters
// (b16EnteredFromStack, which also refuses a spell someone else cast) and
// again as the trigger resolves. A Weftwalking put onto the battlefield any
// other way does nothing.
//
// The static is the engine's granted alternative cost for EVERY player's
// casts (game.GrantedAlternativeCost.EveryPlayer, FirstSpellOnOwnTurn): a
// free cast offered on a player's own turn until they have cast a spell
// that turn, whichever zone it is cast from. It is an alternative cost
// (CR 118.9), so the ruling holds: no other alternative cost goes with it,
// but additional costs, kicker included, are still paid, and a mandatory
// additional cost still must be. A spell that was countered still was cast,
// so it spends the turn's chance; a copy is not cast and does not.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "08af6751-5848-47fb-95f5-57215e899460",
		Name:         "Weftwalking",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 7},
		Triggered: []game.TriggeredAbility{
			InterveningIf(CondBoard{
				Name: "you cast it",
				Test: func(ctx *Context) bool { return b16EnteredFromStack(ctx.Game, ctx.Source()) },
			}, On(game.EventETB, AllOf(Self, weftwalkingYouCastIt),
				"Weftwalking — shuffle your hand and graveyard into your library, then draw seven cards",
				func(g *game.Game, item *game.StackItem) error {
					me := []uuid.UUID{item.Controller}
					return echoOfEonsShuffleEach(g, me, me)
				})),
		},
		GrantedAlternativeCosts: []game.GrantedAlternativeCost{FirstSpellEachPlayerCastsOnTheirTurnFree()},
	})
}

func weftwalkingYouCastIt(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return source != nil && b16EnteredFromStack(g, source.InstanceID)
}
