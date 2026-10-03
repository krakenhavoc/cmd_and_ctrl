package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rowan's Talent — Enchantment — Aura {2}{R}{R}:
//
//	"Enchant planeswalker
//	 Enchanted planeswalker has "[+1]: Up to one target creature gets
//	 +2/+0 and gains first strike and trample until end of turn."
//	 Whenever you activate a loyalty ability of enchanted planeswalker,
//	 copy that ability. You may choose new targets for the copy."
//
// The +1 is a loyalty ability granted to the enchanted planeswalker
// (ADR 0109 §2), sharing the walker's one loyalty activation a turn
// (CR 606.3). "Up to one": with no target chosen it resolves and
// does nothing. The +2/+0 and the two keywords are one effect.
//
// The trigger is Rings of Brighthearth's copy without the {2}: the
// ability to copy is the triggering event's StackItemID, read at
// resolution, while the ability is still on the stack beneath it
// (CR 603.3b). The copy is created, not activated (CR 707.10), so
// it pays no loyalty, does not count toward CR 606.3, and does not
// trigger this again. A copy keeps the original's grantor, so a
// copied Teferi's Talent −12 is still that Talent's emblem.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0af6f3ce-59e8-4797-aa1b-dbdb5288fe3d",
		Name:         "Rowan's Talent",
		Completeness: CompletenessFull,
		Targets:      EnchantPlaneswalker(),
		Grants: []AbilityGrant{{
			Key: rowansTalentGrant,
			Activated: []ActivatedAbility{{
				Label:   "+1: Up to one target creature gets +2/+0 and gains first strike and trample until end of turn.",
				Cost:    LoyaltyCost(1),
				Targets: TargetCreature("up to one target creature").WithCount(0, 1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind == game.TargetCard {
							return untilEndOfTurn(ctx, t.ID, nil,
								"Rowan's Talent — +2/+0, first strike and trample",
								game.ModifyPTMod(2, 0), game.AddKeywordsMod("first strike", "trample"))
						}
					}
					return nil
				},
			}},
			Text: "[+1]: Up to one target creature gets +2/+0 and gains first strike and trample until end of turn.",
		}},
		Static: []game.StaticAbility{GrantAbilitiesToAttached(rowansTalentGrant)},
		Triggered: []game.TriggeredAbility{
			WheneverYouActivateALoyaltyAbilityOfEnchanted(
				"Rowan's Talent — copy that ability",
				rowansTalentCopy),
		},
	})
}

const rowansTalentGrant = "rowans-talent/pump"

// rowansTalentCopy copies the loyalty ability the trigger fired on.
func rowansTalentCopy(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ability := ctx.Trigger().Event.StackItemID
	if ability == uuid.Nil {
		return nil
	}
	return CopyAbility{ItemID: ability, ChooseNewTargets: true}.Apply(ctx)
}
