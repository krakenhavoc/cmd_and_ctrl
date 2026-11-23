package cardsp

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Plasma Caster — Artifact — Equipment {1}{R}:
//
//	"Equipped creature gets +1/+1.
//	 Whenever equipped creature attacks, you get {E}{E} (two energy counters).
//	 Pay {E}{E}: Choose target creature that's blocking equipped creature. Flip a coin. If you win the flip, exile the chosen creature. Otherwise, this Equipment deals 1 damage to it.
//	 Equip {2}"
//
// #1863 and ADR 0129 PR 1 (#1995): "blocking equipped creature" is the
// combat relation read off the creature the Equipment is attached to
// (BlockingEquipped), judged at announce and again at resolution (CR
// 608.2b). The energy price is paid when the ability is activated; the
// flip is the ordinary call-and-win prompt (CR 705.2), and the chosen
// creature is re-checked when the coin lands, so one that has left
// is simply not exiled. The damage comes from the Equipment, as it
// last existed if it has gone (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1c50c637-750c-4e31-a69c-915430fc3194",
		Name:         "Plasma Caster",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(1, 1)},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.AttachedCreatureAttacked(ev, source)
			}, "Plasma Caster — you get {E}{E}", effects.EbYouGetEnergy(2)),
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "Pay {E}{E}: Choose target creature that's blocking equipped creature. Flip a coin. If you win the flip, exile the chosen creature. Otherwise, this Equipment deals 1 damage to it.",
				Cost:    effects.PayEnergy(2),
				Targets: effects.BlockingEquipped(effects.TargetCreature("target creature that's blocking equipped creature")),
				Effect:  plasmaCasterFlip,
			},
			effects.EquipAbility("{2}"),
		},
	})
}

// plasmaCasterFlip flips for the chosen blocker: exiled on a win, 1
// damage from the Equipment on a loss. Nothing is flipped when the
// target is no longer legal (CR 608.2b). The continuation is rebuilt
// from values, never from the item it began with.
func plasmaCasterFlip(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	target := effects.FirstLegalBattlefieldTarget(ctx)
	if target == uuid.Nil {
		return nil
	}
	ref, hasRef := ctx.SourceRef()
	controller, source, self := item.Controller, item.SourceCardID, item.ID
	refNum := 0
	if hasRef {
		refNum = 1
	}
	g.FlipCoinForEffect(game.CoinFlipSpec{
		Flipper:  controller,
		Source:   source,
		Question: "Plasma Caster — call the coin flip",
		// Keyed (#4253 slice 10b): the blocker, the ability and the
		// Equipment as it was pinned now (CR 400.7) ride the call.
		ThenKey: plasmaCasterFlipped,
		Carry:   []uuid.UUID{target, self, ref.ID},
		Nums:    []int{refNum, ref.Epoch},
	})
	return nil
}

// plasmaCasterFlipped is the settled call. An on-disk key: never renamed,
// never reused.
var plasmaCasterFlipped = game.RegisterCoinFlipThen("coin-flip/plasma-caster", func(g *game.Game, r game.CoinFlipped) error {
	target := effects.CarriedID(r.Carry, 0)
	c := effects.NewContext(g, &game.StackItem{
		ID: effects.CarriedID(r.Carry, 1), Kind: game.StackItemActivated,
		Controller: r.Flipper, Owner: r.Flipper, SourceCardID: r.Source,
	})
	if !effects.OnBattlefield(g, target) {
		return nil
	}
	if effects.WonFirstFlip(r) {
		return plasmaCasterExile(c, target)
	}
	if effects.CarriedNum(r.Nums, 0) == 0 {
		return nil
	}
	ref := game.ObjectRef{ID: effects.CarriedID(r.Carry, 2), Epoch: effects.CarriedNum(r.Nums, 1)}
	return effects.DealDamage{SourceObject: &ref, Target: target, Amount: 1}.Apply(c)
})

// plasmaCasterExile is the won flip: the chosen creature is exiled.
func plasmaCasterExile(ctx *effects.Context, target uuid.UUID) error {
	return effects.ExileTarget{Target: target}.Apply(ctx)
}
