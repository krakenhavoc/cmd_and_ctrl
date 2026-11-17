package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kemba, Kha Enduring — Legendary Creature — Cat Cleric {1}{W}, 2/2 (March of
// the Machine: Phyrexia: All Will Be One Commander):
//
//	"Whenever Kemba or another Cat you control enters, attach up to one target
//	 Equipment you control to that creature.
//	 Equipped creatures you control get +1/+1.
//	 {3}{W}{W}: Create a 2/2 white Cat creature token."
//
// The trigger fires for Kemba herself or any Cat entering under her
// controller's control (the subtype a word of her text, CR 612.2), with an
// "up to one" Equipment target: no Equipment is a legal choice and then
// nothing moves. At resolution the Equipment, if still legal (CR 608.2b), is
// attached to the creature that entered; an entering creature that has since
// left is not there to be attached to (CR 701.3b). The static is a layer 7c
// +1/+1 over creatures its controller controls that have an Equipment
// attached (isEquippedCreature, read on every recompute, CR 301.5).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "93c97e1a-bd7e-4307-8ada-58f53a887cf8",
		Name:         "Kemba, Kha Enduring",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label:    "Equipped creatures you control get +1/+1.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.Controller == source.Controller && isEquippedCreature(g, *target)
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power++
				c.Toughness++
			},
		}},
		Triggered: []game.TriggeredAbility{
			Targeting(On(game.EventETB, kembaOrAnotherCatEntered,
				"Kemba, Kha Enduring — attach up to one target Equipment you control to that creature",
				kembaAttachToEntered),
				TargetPermanent("up to one target Equipment you control", HasSubtype("Equipment"), YouControl()).WithCount(0, 1)),
		},
		Activated: []ActivatedAbility{{
			Label:   "{3}{W}{W}: Create a 2/2 white Cat creature token.",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Cost:    ManaCost("{3}{W}{W}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return CreateToken{Template: TokenCard("2/2 white Cat"), N: 1}.Apply(NewContext(g, item))
			},
		}},
	})
}

// kembaOrAnotherCatEntered is the trigger condition: Kemba, or another Cat
// her controller controls, entered.
func kembaOrAnotherCatEntered(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.CardID == source.InstanceID {
		return true
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.Controller == source.Controller && c.HasSubtype(source.TextSubtype("Cat"))
}

// kembaAttachToEntered attaches each still-legal Equipment target to the
// creature that caused the trigger.
func kembaAttachToEntered(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	host := item.Trigger.Event.CardID
	if z := g.FindCardZoneForEffect(host); z == nil || z.Kind != game.ZoneBattlefield {
		return nil
	}
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := g.AttachForEffect(t.ID, game.TargetRef{Kind: game.TargetCard, ID: host}); err != nil {
			return err
		}
	}
	return nil
}
