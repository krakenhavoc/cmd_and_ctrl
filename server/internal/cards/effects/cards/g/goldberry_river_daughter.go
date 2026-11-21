package cardsg

import (
	"sort"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Goldberry, River-Daughter — Legendary Creature — Nymph {1}{U}, 1/3 (The Lord
// of the Rings: Tales of Middle-earth Commander):
//
//	"{T}: Move a counter of each kind not on Goldberry from another target
//	 permanent you control onto Goldberry.
//	 {U}, {T}: Move one or more counters from Goldberry onto another target
//	 permanent you control. If you do, draw a card."
//
// Both are counter moves through the counter primitive (a removal, then a
// placement, so a doubler sees the placement). The first takes one counter
// of every kind the other permanent has and Goldberry does not (CR 122.5),
// fixed when it resolves. The second is Resourceful Defense's "choose how
// many of each kind" (RegisterCounterChoice), where choosing none moves
// nothing and draws nothing; a card is drawn only when at least one counter
// actually moved. Both halves need Goldberry and the other permanent on the
// battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fca3545a-2956-49f9-8465-1c8937deb294",
		Name:         "Goldberry, River-Daughter",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{T}: Move a counter of each kind not on Goldberry from another target permanent you control onto Goldberry.",
				Cost:    effects.TapCost(),
				Targets: effects.Another(effects.TargetPermanent("another target permanent you control", effects.YouControl())),
				Effect:  goldberryGather,
			},
			{
				Label:   "{U}, {T}: Move one or more counters from Goldberry onto another target permanent you control. If you do, draw a card.",
				Purpose: game.Purpose{Draws: 1},
				Cost:    effects.Plus(effects.ManaCost("{U}"), effects.TapCost()),
				Targets: effects.Another(effects.TargetPermanent("another target permanent you control", effects.YouControl())),
				Effect:  goldberryGive,
			},
		},
	})
}

func goldberryGather(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	from, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok || !effects.SourceIsStillThisPermanent(g, item) || !effects.OnBattlefield(g, from) {
		return nil
	}
	src, ok := g.LookupCardForEffect(from)
	self, ok2 := g.LookupCardForEffect(item.SourceCardID)
	if !ok || !ok2 {
		return nil
	}
	var kinds []string
	for kind, n := range src.Counters {
		if n > 0 && self.Counters[kind] <= 0 {
			kinds = append(kinds, kind)
		}
	}
	sort.Strings(kinds)
	for _, kind := range kinds {
		if err := (effects.AddCounter{Target: from, Kind: kind, N: -1}).Apply(ctx); err != nil {
			return err
		}
		if err := (effects.AddCounter{Target: item.SourceCardID, Kind: kind, N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

func goldberryGive(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	to, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok || !effects.SourceIsStillThisPermanent(g, item) || !effects.OnBattlefield(g, to) {
		return nil
	}
	return goldberryMoves.Choose(ctx, item.SourceCardID, []uuid.UUID{to}, 0, "Goldberry, River-Daughter — move one or more counters")
}

// goldberryMoves is Goldberry's clause: the counts chosen on Goldberry move
// onto the target, and a card is drawn if any moved.
var goldberryMoves = effects.RegisterCounterChoice("number/goldberry-river-daughter-move-counters", nil,
	func(ctx *effects.Context, from uuid.UUID, extra []uuid.UUID, picked map[string]int) error {
		if len(extra) == 0 || !effects.OnBattlefield(ctx.Game, extra[0]) || !effects.OnBattlefield(ctx.Game, from) {
			return nil
		}
		to := extra[0]
		kinds := make([]string, 0, len(picked))
		for kind := range picked {
			kinds = append(kinds, kind)
		}
		sort.Strings(kinds)
		moved := 0
		for _, kind := range kinds {
			n := picked[kind]
			src, ok := ctx.Game.LookupCardForEffect(from)
			if !ok {
				return nil
			}
			if have := src.Counters[kind]; have < n {
				n = have
			}
			if n <= 0 {
				continue
			}
			if err := (effects.AddCounter{Target: from, Kind: kind, N: -n}).Apply(ctx); err != nil {
				return err
			}
			if err := (effects.AddCounter{Target: to, Kind: kind, N: n}).Apply(ctx); err != nil {
				return err
			}
			moved += n
		}
		if moved == 0 {
			return nil
		}
		return effects.DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
	})
