package effects

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Volatile Stormdrake — Creature — Drake {1}{U}, 3/2 (Modern Horizons 3):
//
//	"Flying, hexproof from activated and triggered abilities
//	 When this creature enters, exchange control of this creature and
//	 target creature an opponent controls. If you do, you get {E}{E}{E}{E}
//	 (four energy counters), then sacrifice that creature unless you pay an
//	 amount of {E} equal to its mana value."
//
// The hexproof is Shanna's restriction (rprB03AbilitiesFromOpponentsCantTargetThis):
// it refuses abilities and not spells, and only an opponent's (CR 702.11b),
// asked at announcement and again on resolution; a creature that is the
// Stormdrake's opponent's now is judged as it is then, so after the
// exchange it is hexproof against its new opponent's abilities too.
//
// The exchange is Gilded Drake's (ExchangeControl; CR 701.12a: it happens
// only if both are still on the battlefield, and "if you do" gates the rest:
// no exchange, no energy and no sacrifice). The energy is gained first, then
// the question "pay {E} equal to its mana value" is asked through
// PayEnergyUnless (ADR 0129 §3), and the decline (or a payer who cannot
// fund it) sacrifices the creature the Stormdrake traded for, which the
// payer controls by then. A mana value of zero costs nothing to pay.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "08d64b16-9221-4532-890c-d5cdbf181e28",
		Name:            "Volatile Stormdrake",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Purpose:         game.Purpose{Energy: 4},
		TargetingRestrictions: []game.TargetingRestriction{
			rprB03AbilitiesFromOpponentsCantTargetThis("Hexproof from activated and triggered abilities."),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Volatile Stormdrake — exchange control of this creature and target creature an opponent controls; if you do, you get {E}{E}{E}{E}, then sacrifice that creature unless you pay {E} equal to its mana value",
				volatileStormdrakeEnters),
				TargetCreature("target creature an opponent controls", OpponentControls())),
		},
	})
}

func volatileStormdrakeEnters(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	target, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	exchanged, err := ExchangeControl{A: item.SourceCardID, B: target, Label: "Volatile Stormdrake — exchange control"}.ApplyAndReport(ctx)
	if err != nil || !exchanged {
		return err
	}
	if err := (GetEnergy{N: 4}).Apply(ctx); err != nil {
		return err
	}
	card, ok := g.LookupCardForEffect(target)
	if !ok {
		return nil
	}
	mv, _ := g.ManaValueForEffect(card)
	if mv <= 0 {
		return nil
	}
	return PayEnergyUnless{
		N:        mv,
		Question: fmt.Sprintf("Volatile Stormdrake — pay %s, or sacrifice that creature?", EnergySymbols(mv)),
		Else:     volatileStormdrakeSacrifice,
	}.Apply(ctx)
}

// volatileStormdrakeSacrifice is the decline half: "sacrifice that
// creature", the target of the trigger that asked. An on-disk key.
var volatileStormdrakeSacrifice = energyDeclined("pay-unless/gap-b54-volatile-stormdrake-sacrifice",
	func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		id, ok := b16FirstLegalTargetCard(ctx)
		if !ok && len(item.Targets) > 0 && item.Targets[0].Kind == game.TargetCard {
			id, ok = item.Targets[0].ID, true
		}
		if !ok {
			return nil
		}
		return SacrificePermanent{Target: id}.Apply(ctx)
	})
