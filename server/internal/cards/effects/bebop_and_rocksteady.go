package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bebop & Rocksteady — Legendary Creature — Boar Rhino Mutant
// {1}{B/G}{B/G}, 7/5 (Teenage Mutant Ninja Turtles, #4457):
//
//	"Whenever Bebop & Rocksteady attack or block, sacrifice a permanent
//	 unless you discard a card."
//
// Painful Quandary's shape pointed at its own controller: the trigger
// asks one question with the card's two consequences as its answers
// (MayChoice), "Discard a card" or "Sacrifice a permanent". The attack
// half is Liberty Prime's attackDeclared, and the block half fires once
// per block declaration (selfBlocksOnce), however many attackers the
// creature blocks.
//
// An empty hand cannot pay the "unless", so nothing is asked and the
// sacrifice is chosen at once (CR 608.2). The sacrifice is any
// permanent you control, lands and Bebop & Rocksteady included, chosen
// as it happens; with none, nothing is sacrificed (CR 701.21a). A hand
// that emptied between the question and the answer degrades to the
// sacrifice.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a08b0a07-6da2-4ff6-88e5-88355dd84589",
		Name:         "Bebop & Rocksteady",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventAttack, game.EventBlock}, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attackDeclared(ev, source) || selfBlocksOnce(ev, source)
			}, "Bebop & Rocksteady — sacrifice a permanent unless you discard a card", bebopAndRocksteadyAsk),
		},
	})
}

// bebopAndRocksteadyAsk is the trigger's body: ask the controller, or
// sacrifice at once when there is no card to discard.
func bebopAndRocksteadyAsk(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	p := g.PlayerByIDForEffect(item.Controller)
	if p == nil || p.Eliminated {
		return nil
	}
	if p.Hand == nil || p.Hand.Size() == 0 {
		return bebopAndRocksteadySacrifice(ctx)
	}
	return MayChoice{
		Player:   item.Controller,
		Question: "Bebop & Rocksteady — discard a card, or sacrifice a permanent?",
		YesLabel: "Discard a card",
		NoLabel:  "Sacrifice a permanent",
		Then:     bebopAndRocksteadyAnswer,
	}.Apply(ctx)
}

// bebopAndRocksteadySacrifice is "sacrifice a permanent": any permanent
// the controller controls, chosen now.
func bebopAndRocksteadySacrifice(ctx *Context) error {
	return SacrificeChoice{
		Player:     ctx.Controller(),
		Candidates: PermanentsControlledBy(ctx.Game, ctx.Controller()),
		Question:   "Bebop & Rocksteady — sacrifice a permanent",
	}.Apply(ctx)
}

// bebopAndRocksteadyAnswer is the controller's answer. Keyed (#4291).
var bebopAndRocksteadyAnswer = MayYesNo("confirm/bebop-and-rocksteady-discard-or-sacrifice",
	func(ctx *Context, r game.ConfirmAnswer) error {
		p := ctx.Game.PlayerByIDForEffect(r.Chooser)
		if p == nil || p.Hand == nil || p.Hand.Size() == 0 {
			return bebopAndRocksteadySacrifice(ctx)
		}
		ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{
			Player:   r.Chooser,
			Source:   ctx.Source(),
			N:        1,
			Question: "Bebop & Rocksteady — discard a card",
		})
		return nil
	},
	func(ctx *Context, _ game.ConfirmAnswer) error { return bebopAndRocksteadySacrifice(ctx) })
