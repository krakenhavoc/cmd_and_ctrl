package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// clash.go — the card-facing half of the CR 701.30 keyword action
// "clash" (#5424). The engine half is game/clash.go.
//
//	"clash with defending player"                        // Marvo, Deep Operative
//	"clash with an opponent. If you win, …"              // Hoarder's Greed
//	"Whenever you win a clash, …"                        // Marvo, Deep Operative
//
// Each clashing player reveals the top card of their library at the same
// time, then each decides in APNAP order whether to put that card on the
// bottom (CR 701.30a, 701.30c). The decisions are prompts, so anything
// printed after the clash goes in Then, not after Apply returns.

// Clash is "clash with <Opponent>" for the effect's controller. Then, when
// set, is the rest of the card: a registered tail (CardTail) whose
// TailOutcome.Amount is 1 when the controller won the clash (CR 701.30d)
// and 0 otherwise.
type Clash struct {
	Opponent uuid.UUID
	Then     game.KeyedTail
}

func (c Clash) Apply(ctx *Context) error {
	return ctx.Game.ClashForEffect(ctx.Source(), ctx.Controller(), c.Opponent, c.Then)
}

// ClashWithAnOpponent is "clash with an opponent" (CR 701.30b): the
// controller chooses an opponent as the effect resolves, then the two
// clash. Then is Clash.Then, and is told "did not win" when there was no
// opponent to choose.
type ClashWithAnOpponent struct {
	Then game.KeyedTail
}

func (c ClashWithAnOpponent) Apply(ctx *Context) error {
	return ctx.Game.ClashWithAnOpponentForEffect(ctx.Source(), ctx.Controller(), c.Then)
}

// wonTheClash reports whether a Clash's Then was told its controller won.
func wonTheClash(o game.TailOutcome) bool { return o.Amount == 1 }

// youWonAClash is "whenever you win a clash": game.EventClashWon names the
// winner as its Actor (CR 701.30d), whoever's effect started the clash.
func youWonAClash(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.Kind == game.EventClashWon && ev.Actor == source.Controller
}
