package effects

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bebop, Skull & Crossbones — Legendary Creature — Boar Mutant {1}{B},
// 2/1 (Teenage Mutant Ninja Turtles Commander, #5011):
//
//	"Partner with Rocksteady, Mutant Marauder (When this creature
//	 enters, target player may put Rocksteady into their hand from their
//	 library, then shuffle.)
//	 Deathtouch
//	 Whenever Bebop deals combat damage to a player, you may draw X
//	 cards, where X is the number of counters on Bebop. If you do, you
//	 lose X life."
//
// Partner with is the shared entry trigger (PartnerWith); the
// deck-construction half is read off the oracle text. The damage
// trigger fixes X once, as it resolves (the ruling, CR 608.2h): every
// counter of every kind on Bebop, read from the permanent as it last
// existed on the battlefield if it has left (SourcePermanent). Then the
// controller is asked; on a yes they draw X cards and lose X life. With
// no counters X is zero, and drawing and losing nothing is no choice to
// ask about. A draw limit that forbids drawing X makes the answer a
// "no" without asking (MayChoice.Draws, CR 121.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "59a405d7-2f28-41b8-a5bb-011f45a5fdf2",
		Name:            "Bebop, Skull & Crossbones",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			PartnerWith("Bebop, Skull & Crossbones", "Rocksteady, Mutant Marauder"),
			WheneverThisDealsCombatDamageToAPlayer(
				"Bebop — you may draw X cards and lose X life, where X is the number of counters on Bebop",
				bebopMayDraw),
		},
	})
}

// bebopMayDraw counts Bebop's counters and asks.
func bebopMayDraw(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	info, ok := ctx.SourcePermanent()
	if !ok {
		return nil
	}
	x := 0
	for _, n := range info.Counters {
		if n > 0 {
			x += n
		}
	}
	if x <= 0 {
		return nil
	}
	word := strconv.Itoa(x)
	return MayChoice{
		Question: "Bebop, Skull & Crossbones — draw " + word + " cards and lose " + word + " life?",
		YesLabel: "Draw " + word,
		NoLabel:  "Don't draw",
		Draws:    x,
		Then:     bebopDrawAndLose,
		Word:     word,
	}.Apply(ctx)
}

// bebopDrawAndLose is the yes: draw X, then lose X life.
var bebopDrawAndLose = MayYes("confirm/bebop-skull-and-crossbones-draw-lose", func(ctx *Context, r game.ConfirmAnswer) error {
	x := wordInt(r)
	if err := (DrawCards{Player: r.Chooser, N: x}).Apply(ctx); err != nil {
		return err
	}
	return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), r.Chooser, -x)
})
