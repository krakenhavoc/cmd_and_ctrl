package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Heartless Conscription — Sorcery {6}{B}{B} (Outlaws of Thunder Junction
// Commander):
//
//	"Exile all creatures. For each card exiled this way, you may play that
//	 card for as long as it remains exiled, and mana of any type can be
//	 spent to cast that spell. Exile Heartless Conscription."
//
// Sunfall's sweep (ExileAllMatching) with a continuation that runs once the
// batch has settled, told the cards that actually reached exile: each gets
// the spell's controller's any-type play grant for as long as it stays
// exiled. Tokens cease to exist and a commander the owner sent to the
// command zone is not in exile, so neither is granted. The grant is made
// for the cards face up (they are public), then the spell exiles itself,
// last, unless it is a copy (a copy is not a card).
func init() {
	Register(Spec{
		OracleID:     "cb9b317b-bb31-4c19-8b78-0fb9a61324b3",
		Name:         "Heartless Conscription",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepExile}},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return ExileAllMatching{
				Match:    Creature(),
				ThenTail: withN(ctx.ItemTail(heartlessConscriptionRest), copyFlag(item)),
			}.Apply(ctx)
		},
	})
}

// heartlessConscriptionRest is "for each card exiled this way, you may play
// it…; exile Heartless Conscription" as a keyed tail. The whether-copy bit
// rides as N. An on-disk key: never renamed, never reused.
var heartlessConscriptionRest = CardTail("heartless-conscription/play-exiled-then-exile-this",
	func(ctx *Context, o game.TailOutcome) error {
		var firstErr error
		for _, id := range o.IDs {
			if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneExile {
				continue
			}
			if err := ctx.Game.ExileCardWithPermissionForEffect(id, game.CastPermission{
				Player:     ctx.Controller(),
				AnyColor:   true,
				AnyType:    true,
				Duration:   game.WhileInZoneDuration(),
				Source:     ctx.Item.SourceCardID,
				SourceName: "Heartless Conscription",
				Label:      "You may play that card for as long as it remains exiled; mana of any type can be spent to cast that spell",
			}); err != nil && firstErr == nil {
				firstErr = err
			}
		}
		if o.Tail.N == 1 {
			return firstErr
		}
		if err := ctx.Game.ExileCardForEffect(ctx.Item.SourceCardID); err != nil && firstErr == nil {
			firstErr = err
		}
		return firstErr
	})
