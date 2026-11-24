package cardsto

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Toralf, God of Fury // Toralf's Hammer — Kaldheim modal double-faced card
// (#5505), the front face under the bare oracle ID and the back face under
// "<oracle_id>#1":
//
//	Toralf, God of Fury {2}{R}{R}, Legendary Creature — God, 5/4
//	 "Trample
//	  Whenever a creature or planeswalker an opponent controls is dealt
//	  excess noncombat damage, Toralf deals damage equal to the excess to
//	  any target other than that permanent."
//
//	Toralf's Hammer {1}{R}, Legendary Artifact — Equipment
//	 "Equipped creature has "{1}{R}, {T}, Unattach Toralf's Hammer: It deals
//	  3 damage to any target. Return Toralf's Hammer to its owner's hand."
//	  Equipped creature gets +3/+0 as long as it's legendary.
//	  Equip {1}{R}"
//
// The excess is the engine's (Event.ExcessDamage, CR 120.10, a planeswalker's
// loyalty before the damage included), the extra target is chosen as the
// trigger goes on the stack and excludes the damaged permanent. The hammer
// throw is a granted ability whose cost unattaches its GRANTOR
// (UnattachTheGrantor), so "Toralf's Hammer" is the very Equipment that
// gave it, and the effect returns that object (ctx.Grantor) to its owner's
// hand. The creature deals the 3 damage.
const toralfsHammerGrant = "toralfs-hammer/throw"

func init() {
	excess := effects.On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		return effects.ExcessNoncombatDamageToOpponentPermanent(ev, source, g) > 0
	}, "Toralf — deals damage equal to the excess to any target other than that permanent", func(g *game.Game, item *game.StackItem) error {
		if item.Trigger == nil {
			return nil
		}
		ctx := effects.NewContext(g, item)
		for _, t := range ctx.LegalTargets() {
			return effects.DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: item.Trigger.Event.ExcessDamage}.Apply(ctx)
		}
		return nil
	})
	excess = effects.WithTargetsFrom(excess, func(tc game.TriggerContext, _ *game.Card, _ *game.Game) *game.TargetSpec {
		return effects.AnyTargetOtherThan(tc.Event.Target)
	})
	effects.Register(effects.Spec{
		OracleID:        "fffcb71f-e802-436e-917d-eaa6607ab74f",
		Name:            "Toralf, God of Fury",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered:       []game.TriggeredAbility{excess},
	})
	legendary := effects.PumpAttached(3, 0)
	legendary.Label = "Equipped creature gets +3/+0 as long as it's legendary."
	legendary.AppliesTo = func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return effects.AttachedToSource(target, nil, source) && target.HasSupertype("Legendary")
	}
	effects.Register(effects.Spec{
		OracleID:     "fffcb71f-e802-436e-917d-eaa6607ab74f#1",
		Name:         "Toralf's Hammer",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: toralfsHammerGrant,
			Activated: []effects.ActivatedAbility{{
				Label:   "{1}{R}, {T}, Unattach Toralf's Hammer: It deals 3 damage to any target. Return Toralf's Hammer to its owner's hand.",
				Cost:    effects.Plus(effects.ManaCost("{1}{R}"), effects.TapCost(), effects.UnattachTheGrantor()),
				Targets: effects.TargetAny(),
				Effect:  toralfsHammerThrow,
			}},
			Text: "{1}{R}, {T}, Unattach Toralf's Hammer: It deals 3 damage to any target. Return Toralf's Hammer to its owner's hand.",
		}},
		Static: []game.StaticAbility{
			legendary,
			effects.GrantAbilitiesToAttached(toralfsHammerGrant),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{1}{R}")},
	})
}

func toralfsHammerThrow(g *game.Game, item *game.StackItem) error {
	if err := effects.DealDamageToTheTarget(3)(g, item); err != nil {
		return err
	}
	ctx := effects.NewContext(g, item)
	hammer := ctx.Grantor()
	if hammer == uuid.Nil || !effects.OnBattlefield(g, hammer) {
		return nil
	}
	return effects.BounceToHand{Target: hammer}.Apply(ctx)
}
