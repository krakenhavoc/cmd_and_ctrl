package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Emeritus of Truce // Swords to Plowshares — Creature — Cat Cleric
// {1}{W}{W}, 3/3 // Instant {W} (Secrets of Strixhaven, #4456; preparation
// card, CR 722):
//
//	"When this creature enters, target player creates a 1/1 white and
//	 black Inkling creature token with flying. Then if an opponent
//	 controls more creatures than you, this creature becomes prepared.
//	 (While it's prepared, you may cast a copy of its spell. Doing so
//	 unprepares it.)"
//
//	Swords to Plowshares — "Exile target creature. Its controller gains
//	 life equal to its power."
//
// It does not enter prepared. The enters trigger targets a player, who
// gets the Inkling (an opponent, to make the "then" more likely, or you).
// "Then" is load-bearing: the count is taken once the token has landed,
// through the keyed creation's tail, so an Inkling given to an opponent
// counts for them and one kept counts for you, and a CR 616 ordering
// prompt between two token replacements holds the count until it is
// answered. "An opponent controls more creatures than you" is one
// opponent against you, not the opponents added together (Keeper of the
// Accord's comparison). It is not an intervening if: the trigger always
// goes on the stack and makes its token. The prepare spell is Swords to
// Plowshares word for word, so it shares that card's body
// (swords_to_plowshares.go).
//
// No simplification.
const emeritusOfTruceOracleID = "e6c41473-2398-4888-9f0c-467dce7c39a4"

func init() {
	Register(Spec{
		OracleID:     emeritusOfTruceOracleID,
		Name:         "Emeritus of Truce",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Emeritus of Truce — target player creates a 1/1 Inkling with flying, then maybe it becomes prepared",
				emeritusOfTruceEnters), TargetPlayer("target player")),
		},
	})
	Register(Spec{
		OracleID:     emeritusOfTruceOracleID + "#1",
		Name:         "Swords to Plowshares",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      ForTargets(RemovalReturning(0, game.TargetReturn{LifeEqualToPower: true})),
		OnResolve:    swordsToPlowsharesOnResolve,
	})
}

// emeritusOfTruceEnters makes the target player's Inkling, then asks the
// "then" once it has landed.
func emeritusOfTruceEnters(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetPlayer {
			continue
		}
		return g.CreateTokensKeyedForEffect(game.TokenCreation{
			Controller: t.ID,
			Source:     item.SourceCardID,
			Groups: []game.TokenGroup{{
				Template: ctx.TextToken(TokenCard("1/1 white and black Inkling with flying")),
				Count:    1,
			}},
		}, ctx.ItemTail(emeritusOfTrucePrepares))
	}
	return nil
}

// emeritusOfTrucePrepares is "then if an opponent controls more creatures
// than you, this creature becomes prepared". An on-disk key.
var emeritusOfTrucePrepares = CardTail("emeritus-of-truce/then-prepare", func(ctx *Context, o game.TailOutcome) error {
	you := o.Tail.Player
	isCreature := func(c game.Card) bool { return c.IsCreature() }
	if !eachOpponent(ctx.Game, you, func(opp uuid.UUID) bool {
		return b11OpponentControlsMoreThanYou(ctx.Game, you, opp, isCreature)
	}) {
		return nil
	}
	return BecomePrepared{Target: o.Tail.Source}.Apply(ctx)
})
