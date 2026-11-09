package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// std_b9_helpers.go — shared helpers for Standard batches 9 and 10 (#4463).

// stdB9CountersOnSource is the number of counters of every kind on the
// resolving ability's source, live while it is the same object on the
// battlefield and as it last existed once it has left (CR 608.2h).
func stdB9CountersOnSource(ctx *Context) int {
	info, ok := ctx.SourcePermanent()
	if !ok {
		return 0
	}
	n := 0
	for _, c := range info.Counters {
		n += c
	}
	return n
}

// stdB9PutPlusOneOnThisUnlessDrew is the continuation of "draw a card …
// If you don't draw a card this way, put a +1/+1 counter on this
// creature" (Trade Route Envoy): it runs once the draw instruction has
// finished, however many prompts it paused for. DrawThen.N is the number
// of cards the player had drawn this turn before the instruction; a
// larger count after it means a card was drawn.
var stdB9PutPlusOneOnThisUnlessDrew = game.RegisterDrawThen("trade-route-envoy-counter-unless-drew",
	func(g *game.Game, d game.DrawThen) error {
		if len(g.DrawnThisTurn[d.Player]) > d.N {
			return nil
		}
		if !onBattlefield(g, d.Source) {
			return nil
		}
		return g.AddCounterForEffect(d.Source, game.CounterPlusOne, 1)
	})

// stdB9InLibrary is whether the card is still in a library (CR 608.2b).
func stdB9InLibrary(g *game.Game, id uuid.UUID) bool {
	z := g.FindCardZoneForEffect(id)
	return z != nil && z.Kind == game.ZoneLibrary
}

// stdB9BigCreaturesYouControl is the number of creatures the modifier's
// source's controller controls with power 4 or greater (Temur Battlecrier).
func stdB9BigCreaturesYouControl(q game.CostQuery) int {
	if q.Game == nil {
		return 0
	}
	you := q.Source.Controller
	big := And(Creature(), PowerGE(4))
	n := 0
	for _, c := range q.Game.BattlefieldCardsForEffect() {
		if c.Controller == you && big(q.Game, you, c) {
			n++
		}
	}
	return n
}

// stdB9YouControlA is whether `player` controls a permanent `match` admits.
func stdB9YouControlA(g *game.Game, player uuid.UUID, match CardPredicate) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == player && match(g, player, c) {
			return true
		}
	}
	return false
}

// stdB9LibraryEmpty is whether `player` has no card to draw.
func stdB9LibraryEmpty(g *game.Game, player uuid.UUID) bool {
	p := g.PlayerByIDForEffect(player)
	return p == nil || p.Library == nil || len(p.Library.Cards) == 0
}
