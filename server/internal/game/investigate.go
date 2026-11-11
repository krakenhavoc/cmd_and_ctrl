package game

import (
	"sync"

	"github.com/google/uuid"
)

// investigate.go — the CR 701.16 keyword action "investigate" (#5425).
//
//	701.16a "Investigate" means "Create a Clue token." See rule 111.10f.
//
// The action is a token creation and nothing else, so it runs through the
// creation's own CR 614 window (CreateTokensKeyedForEffect): a doubled or
// replaced Clue is still one investigate. What the engine adds is that the
// player INVESTIGATED: EventInvestigate, once per time the action is
// taken, emitted when that creation has settled, whatever it ended up
// making. "Create a Clue token" printed without the word is not an
// investigate and emits nothing, which is why this is an entry point of
// its own and not a watch on Clue creation.
//
// "Investigate twice" takes the action twice: two creations, the second
// behind the first, and two events. The count rides the creation's tail
// as data, so a table waiting on a CR 616 prompt inside one is a restore
// point.
//
// The Clue is the catalog's row (effects.ClueToken, CR 111.10f): this
// package owns no token templates, so the catalog registers its
// constructor at init (RegisterInvestigateToken).

// EventInvestigate records that a player investigated (CR 701.16a). Actor
// is the player, Source the card whose effect told them to. One event per
// time the action is taken: "investigate twice" is two.
const EventInvestigate EventKind = "investigate"

var investigateToken = struct {
	sync.RWMutex
	build func() Card
}{}

// RegisterInvestigateToken registers the constructor of the Clue token an
// investigate creates. Call it once, from the catalog's init.
func RegisterInvestigateToken(build func() Card) {
	investigateToken.Lock()
	defer investigateToken.Unlock()
	investigateToken.build = build
}

// InvestigateForEffect has `player` investigate n times (CR 701.16a) on
// behalf of `source`: each time creates a Clue token under their control
// and then emits EventInvestigate. n of 0 or less does nothing.
//
// It can pause, as any token creation can (a CR 616 ordering prompt
// between two token replacements); the event and the remaining
// investigates follow the answer.
//
// Caller must hold g.mu.
func (g *Game) InvestigateForEffect(player, source uuid.UUID, n int) error {
	if n <= 0 {
		return nil
	}
	investigateToken.RLock()
	build := investigateToken.build
	investigateToken.RUnlock()
	then := KeyedTail{Ref: investigatedTail, Player: player, Source: source, N: n - 1}
	if build == nil {
		// No catalog in this binary (an engine test): the player still
		// investigated, with no Clue to show for it.
		return then.Run(g, nil)
	}
	return g.CreateTokensKeyedForEffect(TokenCreation{
		Controller: player,
		Source:     source,
		Groups:     []TokenGroup{{Template: build(), Count: 1}},
	}, then)
}

// investigatedTail is the end of one investigate: the player has
// investigated, and the next one starts while any are owed (the tail's N).
// An on-disk key: never renamed, never reused.
var investigatedTail = RegisterTail("engine/investigated", func(g *Game, o TailOutcome) error {
	g.EmitEvent(Event{
		Kind:   EventInvestigate,
		Actor:  o.Tail.Player,
		Source: o.Tail.Source,
	})
	return investigateAgain(g, o.Tail.Player, o.Tail.Source, o.Tail.N)
})

// investigateAgain is InvestigateForEffect, reached through a variable so
// the package's initialisation order has no cycle.
var investigateAgain func(g *Game, player, source uuid.UUID, n int) error

func init() {
	investigateAgain = func(g *Game, player, source uuid.UUID, n int) error {
		return g.InvestigateForEffect(player, source, n)
	}
}

// InvestigatedThisTurn is how many times `player` has investigated this
// turn. The tally's listener runs ahead of the trigger harvester, so a
// "whenever you investigate" ability asked about an investigate reads a
// count that already includes it: the turn's first is the one that reads 1.
//
// Caller must hold g.mu.
func (g *Game) InvestigatedThisTurn(player uuid.UUID) int {
	return g.TurnTallyFor(player).Investigates
}
