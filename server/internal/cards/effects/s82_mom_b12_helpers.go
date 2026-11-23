package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_mom_b12_helpers.go — S82, March of the Machine batch 12 (#5477):
// the battle-matters cards. Targets that say "creature, planeswalker, or
// battle", the battle protector reading, and excess damage on all three.

// s82MomB12Damageable is "creature, planeswalker, or battle", read through
// the layered view.
func s82MomB12Damageable(c game.Card) bool {
	return c.IsCreature() || c.IsPlaneswalker() || c.IsBattle()
}

// s82MomB12CreaturePlaneswalkerOrBattle is the target clause "target
// creature, planeswalker, or battle".
func S82MomB12CreaturePlaneswalkerOrBattle(label string) *game.TargetSpec {
	return BattlefieldSpec("permanent", label, func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
		return s82MomB12Damageable(c)
	})
}

// s82MomB12TargetBattle is "target battle".
func S82MomB12TargetBattle() *game.TargetSpec {
	return BattlefieldSpec("permanent", "target battle", func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
		return c.IsBattle()
	})
}

// s82MomB12PlayerOrBattle is "target player or battle".
func S82MomB12PlayerOrBattle() *game.TargetSpec {
	return &game.TargetSpec{
		Mode:    "any",
		Label:   "target player or battle",
		Players: true,
		Zones:   []game.ZoneKind{game.ZoneBattlefield},
		CardOK: func(_ *game.Game, _ uuid.UUID, c game.Card, _ game.ZoneKind) bool {
			return c.IsBattle()
		},
		Min: 1, Max: 1,
	}
}

// s82MomB12AnOpponentProtects is "an opponent protects it" (CR 310.9): the
// battle's protector is a player other than `controller`. A battle with no
// protector is not one an opponent protects.
func s82MomB12AnOpponentProtects(c game.Card, controller uuid.UUID) bool {
	return c.IsBattle() && c.ProtectorPlayerID != uuid.Nil && c.ProtectorPlayerID != controller
}

// s82MomB12DefenseEither is "choose target battle. If an opponent protects
// it, remove N defense counters from it. Otherwise, put N defense counters
// on it." (Etched Host Doombringer, Portent Tracker). The protector is read
// as the instruction resolves; a battle that is gone does nothing (CR 608.2b).
func S82MomB12DefenseEither(ctx *Context, battle uuid.UUID, n int) error {
	c, ok := ctx.Game.LookupCardForEffect(battle)
	if !ok || !c.IsBattle() || !OnBattlefield(ctx.Game, battle) {
		return nil
	}
	if s82MomB12AnOpponentProtects(c, ctx.Controller()) {
		have := c.Counters[game.CounterDefense]
		if n > have {
			n = have
		}
		if n <= 0 {
			return nil
		}
		return AddCounter{Target: battle, Kind: game.CounterDefense, N: -n}.Apply(ctx)
	}
	return AddCounter{Target: battle, Kind: game.CounterDefense, N: n}.Apply(ctx)
}

// s82MomB12DealDamageWithExcess is `amount` damage from the resolving
// spell's source to the creature, planeswalker or battle `target`,
// returning the EXCESS damage (CR 120.10): what landed beyond lethal damage
// for a creature, beyond loyalty for a planeswalker, beyond defense for a
// battle, the greatest of those for a permanent of several types. Read
// after the damage lands, so a prevention shield lowers it.
func S82MomB12DealDamageWithExcess(ctx *Context, target uuid.UUID, amount int) (int, error) {
	c, ok := ctx.Game.LookupCardForEffect(target)
	if !ok || amount <= 0 {
		return 0, nil
	}
	// need is the smallest damage that is no longer "excess" for each type.
	needs := []int{}
	if c.IsCreature() {
		needs = append(needs, max(c.CurrentToughness()-c.DamageMarked, 0))
	}
	if c.IsPlaneswalker() {
		needs = append(needs, max(c.Counters[game.CounterLoyalty], 0))
	}
	if c.IsBattle() {
		needs = append(needs, max(c.Counters[game.CounterDefense], 0))
	}
	cursor := B25LastEventSeq(ctx.Game)
	if err := (DealDamage{Source: ctx.Source(), Target: target, Amount: amount}).Apply(ctx); err != nil {
		return 0, err
	}
	dealt := B27DamageDealtToAfter(ctx.Game, ctx.Source(), target, cursor)
	best := 0
	for _, need := range needs {
		best = max(best, dealt-need)
	}
	return best, nil
}
