package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// CantAttackUnlessYouPayMana is "This creature can't attack unless you pay
// <mana>" (Qal Sisma Behemoth, CR 508.1h): paid by the attacking creature's
// controller whatever they attack, and bound to the permanent that prints it.
func CantAttackUnlessYouPayMana(mana, label string) game.AttackTax {
	return game.AttackTax{
		Label:      label,
		OnAttacker: true,
		AppliesTo:  SelfAttacker,
		ManaCost:   func(game.AttackTaxQuery) string { return mana },
	}
}

// CantBlockUnlessYouPayMana is "This creature can't block unless you pay
// <mana>" (CR 509.1d), the block half of the same clause.
func CantBlockUnlessYouPayMana(mana, label string) game.BlockTax {
	t := BlockTax(mana, label)
	t.AppliesTo = game.BlockedBySelf
	return t
}

// AttachThisToTheEnteringCreature is Stormrider Rig's "Whenever a creature you
// control enters, you may attach this Equipment to it": the optional trigger
// the entering creature's controller answers, then the attach (CR 301.5c,
// 701.3b) if both are still on the battlefield.
func AttachThisToTheEnteringCreature(name string) game.TriggeredAbility {
	return game.TriggeredAbility{
		Watches: []game.EventKind{game.EventETB},
		AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
			c, ok := EnteredUnderYourControl(ev, source, g, false)
			return ok && c.IsCreature() && ev.CardID != source.InstanceID
		},
		Key: name + " — you may attach this Equipment to the creature that entered",
		Effect: func(g *game.Game, item *game.StackItem) error {
			return attachSourceToEntered(g, item.SourceCardID, item.Trigger.Event.CardID)
		},
		OptionalPrompt: &game.TriggerOptionalPrompt{
			Question: name + " — attach this Equipment to the creature that entered?",
		},
	}
}

func attachSourceToEntered(g *game.Game, equipment, entered uuid.UUID) error {
	if !OnBattlefield(g, equipment) || !OnBattlefield(g, entered) {
		return nil
	}
	return g.AttachForEffect(equipment, game.TargetRef{Kind: game.TargetCard, ID: entered})
}
