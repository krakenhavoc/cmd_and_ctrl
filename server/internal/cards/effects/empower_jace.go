package effects

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// empower_jace.go — Reality Fracture's keyword action (CR 701.71,
// ADR 0139, #2796) and the Jace planeswalker token it makes.
//
// CR 701.71a: "To empower Jace N means 'If you don't control a Jace
// planeswalker token, create a blue Jace planeswalker token with 0
// loyalty, "[−1]: Surveil 1," and "[−3]: Draw a card." Choose a Jace
// planeswalker token you control. Put N loyalty counters on it.'"
//
// Three sentences, one instruction. The token is created with no
// loyalty, and it is the same instruction that then puts the counters
// on it, so the CR 704.5i state-based action never sees an empty Jace:
// state-based actions are not checked while a spell or ability is
// resolving (CR 704.3), and a resolution waiting on its own prompt
// holds them (#1289). Writing "create the token" and "put counters on
// it" as two separate effects would get this wrong the moment
// anything ran between them.
//
// What the card files write:
//
//	EmpowerJace{N: 2}.Apply(ctx)                          // Empower Jace 2
//	EmpowerJace{Count: islandsYouControl}.Apply(ctx)      // Empower Jace X, X counted on resolution
//	EmpowerJace{N: 6, Then: drawOne}.Apply(ctx)           // "Empower Jace 6. Draw a card."
//
// The token is a catalog token template (ADR 0083) whose two loyalty
// abilities are ordinary CR 606 activated abilities in the template's
// Activated slot, so the view, the legal-move enumerator, the bots and
// ActivateCatalogAbility all reach them through the same
// ActivatedAbilitiesForCard route a printed planeswalker's take.

// JaceTokenSubtype is the planeswalker type the keyword action looks
// for: "a Jace planeswalker token" is any token that is a planeswalker
// with the Jace subtype — the one this file makes, a token copy of a
// Jace planeswalker card, or anything an effect made a Jace.
const JaceTokenSubtype = "Jace"

// JaceToken is the CR 701.71a token: "a blue Jace planeswalker token
// with 0 loyalty, '[−1]: Surveil 1,' and '[−3]: Draw a card.'"
// (Scryfall tfra, oracle a6a09b69-dd86-4fda-8971-b77a2a4abddb).
//
// Not legendary: two of them may be on the battlefield at once, which
// is why the keyword action asks which one gets the counters.
func JaceToken() game.Card { return tokenFromCatalog(printedJaceToken) }

// printedJaceToken is the Jace token as printed, abilities included.
// Its printed loyalty is 0, so the CR 306.5b entry stamp puts nothing
// on it: the counters come from the instruction that made it.
func printedJaceToken() tokenTemplate {
	return tokenTemplate{
		Slug: "jace",
		Card: game.Card{
			Name:            "Jace",
			TypeLine:        "Token Planeswalker — Jace",
			Colors:          []string{"U"},
			StartingLoyalty: 0,
		},
		Text: "−1: Surveil 1.\n−3: Draw a card.",
		Activated: []game.ActivatedAbilityShape{
			{
				Label: "−1: Surveil 1.",
				Cost:  LoyaltyCost(-1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return Surveil{N: 1}.Apply(NewContext(g, item))
				},
			},
			{
				Label:   "−3: Draw a card.",
				Cost:    LoyaltyCost(-3),
				Purpose: game.Purpose{Draws: 1},
				Effect: func(g *game.Game, item *game.StackItem) error {
					return DrawCards{N: 1}.Apply(NewContext(g, item))
				},
			},
		},
	}
}

// IsJacePlaneswalkerToken reports whether c is "a Jace planeswalker
// token": a token, a planeswalker, and a Jace. All three are read off
// the object's current characteristics, so a token copy of Jace,
// Reality Sculptor counts and a Jace token that stopped being a
// planeswalker would not.
func IsJacePlaneswalkerToken(c game.Card) bool {
	return c.IsToken() && c.IsPlaneswalker() && c.HasSubtype(JaceTokenSubtype)
}

// JaceTokensControlledBy lists the Jace planeswalker tokens `player`
// controls, in battlefield order. ONE walk for the keyword action's
// "if you don't control one" and its "choose one", so the two cannot
// disagree.
//
// Caller must hold g.mu.
func JaceTokensControlledBy(g *game.Game, player uuid.UUID) []uuid.UUID {
	g.RecomputeLayersIfStaleLocked()
	var out []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller == player && IsJacePlaneswalkerToken(c) {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// EmpowerJace is the keyword action (CR 701.71a). See the file comment.
type EmpowerJace struct {
	// N is the number of loyalty counters for a fixed "Empower Jace N".
	N int

	// Count, when set, is "Empower Jace X, where X is …": it is
	// evaluated once, as the action begins, against the live game
	// (Jace, Reality Sculptor's Islands; Repurposed Enforcer's
	// creatures). It replaces N. A negative count is zero.
	Count func(ctx *Context) int

	// Then is the rest of the effect, for a card that prints more after
	// "Empower Jace N." (Protege's Awakening's "Draw a card."). The
	// action may pause — on a CR 616 ordering prompt for the token's
	// creation or the counters, or on the question of which Jace gets
	// them — so anything after it goes here, never on the next line.
	// It runs once the counters are on, whether or not any were placed.
	Then func(ctx *Context) error
}

// Apply runs the action for the resolving item's controller: the
// player who "empowers Jace" is always the one who controls the
// instruction (CR 701.71a's "you").
func (e EmpowerJace) Apply(ctx *Context) error {
	player := ctx.Controller()
	n := e.N
	if e.Count != nil {
		n = e.Count(ctx)
	}
	if n < 0 {
		n = 0
	}
	item := ctx.Item
	then := e.Then
	finish := func(g *game.Game) error { return resumeClause(g, item, then) }
	place := func(g *game.Game) error {
		return empowerChosenJaceLocked(g, item, player, n, finish)
	}
	if len(JaceTokensControlledBy(ctx.Game, player)) > 0 {
		return place(ctx.Game)
	}
	// "If you don't control a Jace planeswalker token, create" one.
	// Through the one creation path, so the CR 701.7b window applies:
	// a Doubling Season makes two Jaces, and the "choose" below picks
	// which of them gets the counters (the other has 0 loyalty and
	// goes at the next state-based action check, CR 704.5i).
	return ctx.Game.CreateTokensThenForEffect(game.TokenCreation{
		Controller: player,
		Source:     ctx.Source(),
		Groups:     []game.TokenGroup{{Template: JaceToken(), Count: 1}},
	}, func(g *game.Game, _ []uuid.UUID) error {
		return place(g)
	})
}

// empowerChosenJaceLocked is the "choose a Jace planeswalker token you
// control. Put N loyalty counters on it." half. One candidate is not a
// choice and asks nothing; two or more ask the controller (an
// own-permanents pick, which the enumerator and the bots already
// answer). None — the creation was replaced away — places nothing and
// still finishes the instruction.
//
// The counters go through the CR 614 counter pipeline with the
// controller as their placer (CR 701.71a: "you" put them), so a
// Doubling Season doubles them and "whenever you put one or more
// loyalty counters" sees who did it. This is an effect, not a loyalty
// cost (ADR 0032 §8's distinction), so the replacements apply.
//
// Caller must hold g.mu.
func empowerChosenJaceLocked(g *game.Game, item *game.StackItem, player uuid.UUID, n int, finish func(*game.Game) error) error {
	putOn := func(g *game.Game, jace uuid.UUID) error {
		return g.AddCounterByThenForEffect(player, jace, game.CounterLoyalty, n, func(g *game.Game, _ int) error {
			return finish(g)
		})
	}
	jaces := JaceTokensControlledBy(g, player)
	switch len(jaces) {
	case 0:
		return finish(g)
	case 1:
		return putOn(g, jaces[0])
	}
	return ChoosePermanents{
		Player:   player,
		Question: fmt.Sprintf("Empower Jace %d — choose the Jace token to put the loyalty counters on", n),
		Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
			return JaceTokensControlledBy(g, of), 1, 1
		},
		Then: func(ctx *Context, picked game.PromptedPicks) error {
			chosen := picked.Cards()
			if len(chosen) == 0 {
				// Every Jace on offer left before the answer:
				// nothing to put the counters on.
				return finish(ctx.Game)
			}
			return putOn(ctx.Game, chosen[0])
		},
	}.Apply(NewContext(g, item))
}

// JaceLoyaltyAmong is "the number of loyalty counters among Jaces you
// control" (Jace, Reality Sculptor): every permanent `player` controls
// with the Jace subtype, token or card. Caller must hold g.mu.
func JaceLoyaltyAmong(g *game.Game, player uuid.UUID) int {
	g.RecomputeLayersIfStaleLocked()
	total := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == player && c.HasSubtype(JaceTokenSubtype) {
			total += c.Counters[game.CounterLoyalty]
		}
	}
	return total
}
