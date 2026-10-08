package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:     "1c50c637-750c-4e31-a69c-915430fc3194",
		Name:         "Plasma Caster",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttached(1, 1)},
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attachedCreatureAttacked(ev, source)
			}, "Plasma Caster — you get {E}{E}", ebYouGetEnergy(2)),
		},
		Activated: []ActivatedAbility{
			{
				Label:   "Pay {E}{E}: Choose target creature that's blocking equipped creature. Flip a coin. If you win the flip, exile the chosen creature. Otherwise, this Equipment deals 1 damage to it.",
				Cost:    PayEnergy(2),
				Targets: BlockingEquipped(TargetCreature("target creature that's blocking equipped creature")),
				Effect:  plasmaCasterFlip,
			},
			EquipAbility("{2}"),
		},
	})
}

// plasmaCasterFlip flips for the chosen blocker: exiled on a win, 1
// damage from the Equipment on a loss. Nothing is flipped when the
// target is no longer legal (CR 608.2b). The continuation is rebuilt
// from values, never from the item it began with.
func plasmaCasterFlip(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	target := FirstLegalBattlefieldTarget(ctx)
	if target == uuid.Nil {
		return nil
	}
	ref, hasRef := ctx.SourceRef()
	controller, source, self := item.Controller, item.SourceCardID, item.ID
	g.FlipCoinForEffect(game.CoinFlipSpec{
		Flipper:  controller,
		Source:   source,
		Question: "Plasma Caster — call the coin flip",
		Then: func(g *game.Game, result game.CoinFlipResult) error {
			c := NewContext(g, &game.StackItem{
				ID: self, Kind: game.StackItemActivated,
				Controller: controller, Owner: controller, SourceCardID: source,
			})
			if !onBattlefield(g, target) {
				return nil
			}
			if len(result.Won) > 0 && result.Won[0] {
				return plasmaCasterExile(c, target)
			}
			if !hasRef {
				return nil
			}
			return DealDamage{SourceObject: &ref, Target: target, Amount: 1}.Apply(c)
		},
	})
	return nil
}

// plasmaCasterExile is the won flip: the chosen creature is exiled.
func plasmaCasterExile(ctx *Context, target uuid.UUID) error {
	return ExileTarget{Target: target}.Apply(ctx)
}
