package effects

import (
	"sort"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// rprB23ExileChosenCreatureAndOthersWithItsName is Sever the Bloodline's
// body: the announced creature, if still legal, and every other creature on
// the battlefield that shares its name (anyone's, tokens included), exiled
// together. The name is read at resolution off the effective
// characteristics, so a copy of the target goes with it, and each creature is
// exiled through the single-target verb.
func rprB23ExileChosenCreatureAndOthersWithItsName(ctx *Context) error {
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	chosen, found := ctx.Game.LookupCardForEffect(id)
	if !found {
		return nil
	}
	name := chosen.Effective().Name
	ids := []uuid.UUID{id}
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.InstanceID != id && c.IsCreature() && c.Effective().Name == name {
			ids = append(ids, c.InstanceID)
		}
	}
	for _, target := range ids {
		if !onBattlefield(ctx.Game, target) {
			continue
		}
		if err := (ExileTarget{Target: target}).Apply(ctx.asGroupMember()); err != nil {
			return err
		}
	}
	return nil
}

// rprB23SabertoothReturn bounces the creature Temur Sabertooth's controller
// picked and, behind the move, gives the Sabertooth indestructible until end
// of turn, only if the creature really reached a hand ("if you do").
var rprB23SabertoothReturn = RegisterPermanentsPick("prompt-run/temur-sabertooth-return", PermanentsDef{
	Then: func(ctx *Context, picked game.PromptedPicks) error {
		for _, id := range picked.Cards() {
			if !onBattlefield(ctx.Game, id) {
				continue
			}
			return BounceToHand{Target: id, ThenTail: ctx.ItemTail(rprB23SabertoothGrant)}.Apply(ctx)
		}
		return nil
	},
})

// rprB23SabertoothGrant is "this creature gains indestructible until end of
// turn", told whether the bounce happened. An on-disk key.
var rprB23SabertoothGrant = CardTail("temur-sabertooth/indestructible", func(ctx *Context, o game.TailOutcome) error {
	if !oneWentThisWay(o) || !onBattlefield(ctx.Game, ctx.Source()) || ctx.Game.AbilitySourceIsNewObjectForEffect(ctx.Item) {
		return nil
	}
	return GrantKeywordUntilEOT{Target: ctx.Source(), Keywords: []string{"indestructible"}, Label: "Temur Sabertooth — indestructible"}.Apply(ctx)
})

// rprB23BogbonderMoves is Slippery Bogbonder's "move any number of counters
// from among creatures you control onto that creature": each of the
// controller's other creatures that carries counters is asked in turn, in
// battlefield order, how many of each kind to move, and the counts move
// onto the target before the next is asked. The target rides the chain as
// extra[0], the creatures still to ask behind it. An on-disk key.
var rprB23BogbonderMoves *counterChoice

func init() {
	rprB23BogbonderMoves = registerCounterChoice("number/slippery-bogbonder-move-counters", nil,
		func(ctx *Context, from uuid.UUID, extra []uuid.UUID, picked map[string]int) error {
			if len(extra) == 0 {
				return nil
			}
			to := extra[0]
			if onBattlefield(ctx.Game, to) {
				kinds := make([]string, 0, len(picked))
				for kind := range picked {
					kinds = append(kinds, kind)
				}
				sort.Strings(kinds)
				for _, kind := range kinds {
					src, ok := ctx.Game.LookupCardForEffect(from)
					if !ok || !onBattlefield(ctx.Game, from) {
						break
					}
					n := min(picked[kind], src.Counters[kind])
					if n <= 0 {
						continue
					}
					if err := (AddCounter{Target: from, Kind: kind, N: -n}).Apply(ctx); err != nil {
						return err
					}
					if err := (AddCounter{Target: to, Kind: kind, N: n}).Apply(ctx); err != nil {
						return err
					}
				}
			}
			return rprB23BogbonderNext(ctx, to, extra[1:])
		})
}

// rprB23BogbonderNext asks about the next creature in line that still
// carries counters.
func rprB23BogbonderNext(ctx *Context, to uuid.UUID, rest []uuid.UUID) error {
	if !onBattlefield(ctx.Game, to) {
		return nil
	}
	for i, id := range rest {
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok || !onBattlefield(ctx.Game, id) {
			continue
		}
		has := false
		for _, n := range c.Counters {
			if n > 0 {
				has = true
			}
		}
		if !has {
			continue
		}
		extra := append([]uuid.UUID{to}, rest[i+1:]...)
		return rprB23BogbonderMoves.choose(ctx, id, extra, 0, "Slippery Bogbonder — move any number of counters from "+c.Name+" onto that creature")
	}
	return nil
}

// rprB23BogbonderResolve is the enters trigger: a hexproof counter on the
// target creature, then the moves.
func rprB23BogbonderResolve(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	to, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	if err := (AddCounter{Target: to, Kind: game.CounterHexproof, N: 1}).Apply(ctx); err != nil {
		return err
	}
	var from []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == ctx.Controller() && c.IsCreature() && c.InstanceID != to {
			from = append(from, c.InstanceID)
		}
	}
	return rprB23BogbonderNext(ctx, to, from)
}
