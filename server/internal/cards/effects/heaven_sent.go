package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Heaven Sent — Enchantment — Saga {U}{R} (Doctor Who):
//
//	"I, II — Investigate.
//	 III — This Saga deals 1 damage to each opponent. Then if an opponent
//	 has 0 or less life, draw seven cards. Otherwise, exile this Saga and
//	 you may cast it this turn."
//
// III deals the damage first, then looks at the opponents' life as it
// stands. With no opponent at 0 or less, the Saga (still on the battlefield
// while its last chapter resolves, CR 714.4) is exiled and a cast permission
// over it lasts until end of turn — the "you may cast it this turn" is the
// permission itself, which the controller can leave unused.
//
// No simplification.
func init() {
	const name = "Heaven Sent"
	investigate := Do(Investigate{N: 1})
	Register(Spec{
		OracleID:     "1863ba60-f786-4e52-90fc-f3135ca0257f",
		Name:         name,
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 7},
		Triggered: []game.TriggeredAbility{
			ChapterTrigger(1, SagaChapterLabel(name, 1, "investigate"), investigate),
			ChapterTrigger(2, SagaChapterLabel(name, 2, "investigate"), investigate),
			ChapterTrigger(3, SagaChapterLabel(name, 3, "1 damage to each opponent; then draw seven cards if an opponent has 0 or less life, otherwise exile this Saga and you may cast it this turn"),
				heavenSentFinale),
		},
	})
}

func heavenSentFinale(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := (DamageEachOpponent{N: 1}).Apply(ctx); err != nil {
		return err
	}
	for _, id := range ctx.Opponents() {
		if p := g.PlayerByIDForEffect(id); p != nil && p.Life <= 0 {
			return DrawCards{N: 7}.Apply(ctx)
		}
	}
	if !onBattlefield(g, item.SourceCardID) {
		return nil
	}
	return ExileTarget{Target: item.SourceCardID, ThenTail: ctx.ItemTail(heavenSentExiled)}.Apply(ctx)
}

// heavenSentExiled grants the cast permission over the exiled Saga for the
// rest of the turn. An on-disk key.
var heavenSentExiled = game.RegisterTail("then/heaven-sent-cast-permission", func(g *game.Game, o game.TailOutcome) error {
	for _, id := range o.IDs {
		g.GrantCastPermissionOverCardForEffect(id, game.CastPermission{
			Player: o.Tail.Player,
			Zone:   game.ZoneExile,
			Source: o.Tail.Source,
			Label:  "Heaven Sent — you may cast this Saga this turn",
		})
	}
	return nil
})
