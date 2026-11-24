package cardsv

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vaevictis Asmadi, the Dire — Legendary Creature — Elder Dragon {3}{B}{R}{G},
// 6/6 (Secret Lair Drop, #5462):
//
//	"Flying
//	 Whenever Vaevictis Asmadi attacks, for each player, choose target
//	 permanent that player controls. Those players sacrifice those
//	 permanents. Each player who sacrificed a permanent this way reveals the
//	 top card of their library, then puts it onto the battlefield if it's a
//	 permanent card."
//
// One clause per player, as the trigger goes on the stack: its controller's
// own required pick, and one required pick for each seat after theirs in turn
// order (Blatant Thievery's per-opponent clauses, a seat that is gone not
// required), each judged again as the ability resolves (CR 608.2b). The chosen
// permanents are sacrificed together by their controllers (CR 701.21; one that
// can't be sacrificed is left, and so is one that left already), not chosen
// by those players. Each player who really sacrificed one, read from the
// controller the permanent had on the battlefield (CR 608.2h), then reveals
// their top card and puts it onto the battlefield under their control if it is
// a permanent card; a library with no cards reveals nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c42ce2e9-4dfe-468a-8f41-49e187cb91d4",
		Name:            "Vaevictis Asmadi, the Dire",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WheneverThisAttacks("Vaevictis Asmadi — for each player, choose target permanent that player controls; those players sacrifice those permanents, then each reveals the top card of their library and puts it onto the battlefield if it's a permanent card",
					vaevictisAttack),
				effects.Clauses(
					effects.TargetPermanent("target permanent you control", effects.YouControl()),
					effects.ForNthOpponent(1, effects.TargetPermanent("target permanent controlled by the first opponent after you in turn order")),
					effects.ForNthOpponent(2, effects.TargetPermanent("target permanent controlled by the second opponent after you in turn order")),
					effects.ForNthOpponent(3, effects.TargetPermanent("target permanent controlled by the third opponent after you in turn order")),
				)),
		},
	})
}

func vaevictisAttack(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	var ids []uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard && effects.OnBattlefield(g, t.ID) {
			ids = append(ids, t.ID)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	return g.SacrificeAllKeyedForEffect(item.SourceCardID, ids, ctx.ItemTail(vaevictisSacrificed))
}

// vaevictisSacrificed reveals, for each player who sacrificed one, their top
// card and puts it onto the battlefield if it is a permanent card. An on-disk
// key: never renamed, never reused.
var vaevictisSacrificed = effects.CardTail("vaevictis-asmadi/sacrificed", func(ctx *effects.Context, o game.TailOutcome) error {
	g := ctx.Game
	var players []uuid.UUID
	seen := map[uuid.UUID]bool{}
	for _, id := range o.IDs {
		info, ok := g.LastKnownPermanentForEffect(id)
		if !ok || seen[info.Controller] {
			continue
		}
		seen[info.Controller] = true
		players = append(players, info.Controller)
	}
	for _, player := range players {
		p := g.PlayerByIDForEffect(player)
		if p == nil || p.Eliminated {
			continue
		}
		for _, id := range g.RevealTopOfLibraryForEffect(player, ctx.Source(), 1, "Vaevictis Asmadi — reveal the top card of their library") {
			c, ok := g.LookupCardForEffect(id)
			if !ok || !c.IsPermanent() {
				continue
			}
			if err := effects.PutCardsFromZoneTogether(ctx, game.ZoneLibrary, []uuid.UUID{id}, player, false); err != nil {
				return err
			}
		}
	}
	return nil
})
