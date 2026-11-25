package cardsbl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Breath of Fury — Enchantment — Aura {2}{R}{R}:
//
//	"Enchant creature you control
//	 When enchanted creature deals combat damage to a player, sacrifice it and
//	 attach this Aura to a creature you control. If you do, untap all creatures
//	 you control and after this phase, there is an additional combat phase."
//
// The trigger is the Aura's own, matched through the attachment, and fires for
// combat damage to any player, as printed. Its body runs in order (CR 608.2c):
// the enchanted creature is sacrificed; the creature to attach to is chosen as
// it resolves (it doesn't target, Choking Miasma's prompt) among the creatures
// you control; and the Aura is moved onto it (CR 303.4f: an Aura moved this way
// ignores its enchant restriction only where the card says so; here it is
// "a creature you control", the Aura's own restriction). "If you do" is read as
// the attach having happened: if you control no other creature the Aura stays
// attached to nothing and is put into the graveyard by state-based actions, and
// you neither untap nor get the combat. A Breath of Fury that has left the
// battlefield attaches nothing. Untapping and the extra combat phase are
// Bloodthirster's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fdaa0faf-e7ee-48b3-9330-de0fc3792ad3",
		Name:         "Breath of Fury",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(effects.YouControl()),
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Kind != game.EventDealDamage || !ev.Combat || ev.Amount <= 0 || !source.IsAttachedTo(ev.Source) {
					return false
				}
				return g.PlayerByIDForEffect(ev.Target) != nil
			},
			Key: "Breath of Fury — sacrifice it and attach this Aura to a creature you control; if you do, untap all creatures you control and there is an additional combat phase",
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.SacrificePermanent{Target: item.Trigger.Event.Source, ThenTail: ctx.ItemTail(breathOfFurySacrificed)}.Apply(ctx)
			},
		}},
	})
}

// breathOfFurySacrificed chooses the creature to attach to once the enchanted
// creature has been sacrificed. An on-disk key.
var breathOfFurySacrificed = effects.CardTail("breath-of-fury/sacrificed", func(ctx *effects.Context, o game.TailOutcome) error {
	if !effects.OneWentThisWay(o) {
		return nil
	}
	if len(effects.CreaturesControlledByPlayer(ctx.Game, ctx.Controller())) == 0 {
		return nil
	}
	return effects.ChoosePermanents{
		Question:   "Breath of Fury — attach this Aura to a creature you control",
		Candidates: effects.HumiliateCreatures,
		Keyed:      breathOfFuryChosen,
	}.Apply(ctx)
})

// breathOfFuryChosen attaches the Aura to the creature chosen and, when that
// happened, untaps your creatures and adds the combat phase. An on-disk key.
var breathOfFuryChosen = effects.RegisterPermanentsPick("prompt-run/breath-of-fury-attach", effects.PermanentsDef{
	Then: func(ctx *effects.Context, picked game.PromptedPicks) error {
		ids := picked.Cards()
		if len(ids) == 0 || ctx.Item == nil || ctx.IsNewSourceObject(ctx.Source()) {
			return nil
		}
		if err := ctx.Game.AttachSourceForEffect(ctx.Item, game.TargetRef{Kind: game.TargetCard, ID: ids[0]}); err != nil {
			return err
		}
		if !breathOfFuryAttached(ctx.Game, ctx.Source(), ids[0]) {
			return nil
		}
		mine := effects.CreaturesControlledByPlayer(ctx.Game, ctx.Controller())
		if err := effects.UntapEach(ctx, mine); err != nil {
			return err
		}
		return effects.ExtraCombatAfterThisPhase().Apply(ctx)
	},
})

// breathOfFuryAttached reports whether the permanent `aura` is on the battlefield
// attached to `host`.
func breathOfFuryAttached(g *game.Game, aura, host uuid.UUID) bool {
	c, ok := effects.BattlefieldCardForEffect(g, aura)
	return ok && c.IsAttachedTo(host)
}
