package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b678_helpers.go — shared pieces of the Nemesis batch 678 cards
// (#5618).

// OldB678DestroyEnchantedOnLeave is "When this Aura leaves the battlefield,
// destroy enchanted creature. That creature can't be regenerated" (Parallax
// Dementia): the creature is the one the Aura was attached to as it last
// existed (CR 608.2h). An Aura that left because its host did, or whose host is
// no longer the same object, destroys nothing.
func OldB678DestroyEnchantedOnLeave(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	info, ok := ctx.SourcePermanent()
	if !ok || info.AttachedTo.Kind != game.TargetCard || info.AttachedTo.ID == uuid.Nil {
		return nil
	}
	host := info.AttachedTo.ID
	if info.AttachedObject != nil && g.ObjectEpochForEffect(host) != info.AttachedObject.Epoch {
		return nil
	}
	if !OnBattlefield(g, host) {
		return nil
	}
	return DestroyTarget{Target: host, CantBeRegenerated: true}.Apply(ctx)
}

// OldB678SacrificeUnlessPay is "that player sacrifices a permanent of their
// choice unless they pay {1}" (Aether Barrier): the caster is asked, and the
// sacrifice is theirs.
func OldB678SacrificeUnlessPay(name string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		return PayUnless{
			Chooser:  item.Trigger.Event.Actor,
			Cost:     "{1}",
			Question: name + " — pay {1} or sacrifice a permanent?",
			Else:     oldB678BarrierDecline,
		}.Apply(NewContext(g, item))
	}
}

var oldB678BarrierDecline = PayUnlessElse("pay-unless/s82-old-b678-aether-barrier-sacrifice", func(ctx *Context, r game.PayUnlessAnswer) error {
	B17PlayerSacrificesN(ctx.Game, ctx.Source(), r.Chooser, 1, "Aether Barrier — sacrifice a permanent")
	return nil
})

// OldB678TargetOpponentExilesFromHandLinked is "Target opponent exiles a card
// from their hand" whose card the same source's leave trigger reads back as
// "exiled with it" (Parallax Nexus): the opponent picks (CR 406.3) and the
// exile stamps the link to this object. A Nexus that has left exiles with no
// link.
func OldB678TargetOpponentExilesFromHandLinked(g *game.Game, item *game.StackItem) error {
	id := TargetedPlayerOf(NewContext(g, item))
	p := g.PlayerByIDForEffect(id)
	if id == uuid.Nil || p == nil || p.Hand == nil || p.Hand.Size() == 0 {
		return nil
	}
	cards := make([]uuid.UUID, 0, p.Hand.Size())
	for _, c := range p.Hand.Cards {
		cards = append(cards, c.InstanceID)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    id,
		FromPlayer: id,
		Source:     item.SourceCardID,
		Question:   "Parallax Nexus — exile a card from your hand",
		Cards:      cards,
		Min:        1,
		Max:        1,
		ThenKey:    oldB678ExileLinked,
		LiveZone:   game.ZoneHand,
	})
	return nil
}

// oldB678ExileLinked is the keyed answer to
// OldB678TargetOpponentExilesFromHandLinked. An on-disk key: never renamed,
// never reused.
var oldB678ExileLinked = game.RegisterChooseCardsThen("choose-cards/s82-old-b678-exile-linked", func(g *game.Game, r game.CardsChosen) error {
	ref, linked := g.PermanentRefForEffect(r.Source)
	for _, id := range r.Picked {
		if linked {
			if err := g.ExileLinkedForEffect(ref, id); err != nil {
				return err
			}
			continue
		}
		if err := g.ExileCardForEffect(id); err != nil {
			return err
		}
	}
	return nil
})

// OldB678Rupture is Rupture's body: the controller sacrifices a creature, then
// the spell deals damage equal to that creature's last-known power to each
// creature without flying and each player (Flamebreak's recipients).
func OldB678Rupture(g *game.Game, item *game.StackItem) error {
	return g.PlayerSacrificesThenKeyedForEffect(item.SourceCardID, item.Controller,
		SacrificeSpec("a creature", Creature()),
		"Rupture — sacrifice a creature", 1,
		oldB678RuptureDamage, []uuid.UUID{item.Controller})
}

var oldB678RuptureDamage = game.RegisterPromptRunThen("prompt-run/s82-old-b678-rupture", func(g *game.Game, r game.RunDone) error {
	if len(r.Carry) == 0 {
		return nil
	}
	cards := r.Sacrifices().Cards()
	if len(cards) == 0 {
		return nil
	}
	n := DepartedCreaturePower(g, cards[0])
	if n <= 0 {
		return nil
	}
	ctx := KeyedContext(g, r.Carry[0], r.Source)
	return DamageEachMatchingAndEachPlayerThen(ctx, And(Creature(), WithoutKeyword("flying")), n, RecipientThen{})
})
