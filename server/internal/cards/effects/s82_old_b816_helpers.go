package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b816_helpers.go — shared pieces of the S82 OLD batch 816 cards
// (Alliances, Auras, #5634).

// OldB816OtherThanEnchantedCreature narrows a clause to "other than enchanted
// creature" (Kjeldoran Pride's move, Veteran's Voice's second creature): any
// candidate but the permanent the ability's Aura is attached to
// (TargetSpec.NotTheSourcesHost, judged in the one shared walk). Wrap the
// whole clause, like Another.
func OldB816OtherThanEnchantedCreature(spec *game.TargetSpec) *game.TargetSpec {
	spec.NotTheSourcesHost = true
	return spec
}

// OldB816TapEnchantedCreature is "Tap enchanted creature" as an Aura's
// activation cost: one untapped creature, the one the Aura is attached to
// (TargetSpec.IsTheSourcesHost). Like every tap-others cost it is not the {T}
// symbol (CR 302.6), so a creature that arrived this turn may pay it, and it
// does not target.
func OldB816TapEnchantedCreature() game.AbilityCost {
	return game.AbilityCost{TapOthers: &game.TapOthersCost{
		Count:  1,
		Filter: oldB816EnchantedFilter(),
		Label:  "enchanted creature",
	}}
}

func oldB816EnchantedFilter() *game.TargetSpec {
	spec := TargetPermanent("enchanted creature", Creature())
	spec.IsTheSourcesHost = true
	return spec
}

// OldB816DiscardOneOfThem is the pick of "draw three cards, then discard one of
// them" (Casting of Bones): the chosen card is discarded through the one
// discard path. An on-disk key.
var oldB816DiscardOneOfThem = game.RegisterChooseCardsThen("choose-cards/old-b816-discard-one-of-them", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	return g.DiscardCardsKeyedForEffect(r.Chooser, r.Picked, r.Source, game.KeyedTail{})
})

// OldB816DrawThreeThenDiscardOneOfThem is "draw three cards, then discard one
// of them": the discard is a pick over the cards just drawn (the hand's new
// ones), not a card that was already there. A draw cut short by an empty
// library leaves fewer to choose from.
func OldB816DrawThreeThenDiscardOneOfThem(g *game.Game, item *game.StackItem) error {
	player := item.Controller
	p := g.PlayerByIDForEffect(player)
	if p == nil || p.Hand == nil {
		return nil
	}
	before := map[uuid.UUID]bool{}
	for _, c := range p.Hand.Cards {
		before[c.InstanceID] = true
	}
	if err := g.DrawNForEffect(player, 3); err != nil {
		return err
	}
	var drawn []uuid.UUID
	for _, c := range p.Hand.Cards {
		if !before[c.InstanceID] {
			drawn = append(drawn, c.InstanceID)
		}
	}
	if len(drawn) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  player,
		Source:   item.SourceCardID,
		Question: "Casting of Bones — discard one of the cards you drew",
		Cards:    drawn,
		Min:      1,
		Max:      1,
		LiveZone: game.ZoneHand,
		ThenKey:  oldB816DiscardOneOfThem,
	})
	return nil
}

// OldB816EnchantedCreatureIs reports whether the Aura `source`'s enchanted
// creature has the colour `letter` as the Aura's text reads it ("white" is a
// word of the Aura, CR 612.2). False when the Aura is attached to nothing.
func OldB816EnchantedCreatureIs(g *game.Game, source uuid.UUID, letter string) bool {
	aura, ok := g.LookupCardForEffect(source)
	if !ok || aura.AttachedTo.Kind != game.TargetCard {
		return false
	}
	host, ok := g.LookupCardForEffect(aura.AttachedTo.ID)
	return ok && host.HasColor(aura.TextColor(letter))
}

// OldB816EnchantedCreatureIsUntapped reports whether the Aura's enchanted
// creature is untapped.
func OldB816EnchantedCreatureIsUntapped(g *game.Game, source uuid.UUID) bool {
	aura, ok := g.LookupCardForEffect(source)
	if !ok || aura.AttachedTo.Kind != game.TargetCard {
		return false
	}
	host := oldB816HostCard(g, aura.AttachedTo.ID)
	return host != nil && !host.Tapped
}

func oldB816HostCard(g *game.Game, id uuid.UUID) *game.Card {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.InstanceID == id {
			cp := c
			return &cp
		}
	}
	return nil
}

// OldB816UntapEnchantedCreature is "Untap enchanted creature" as an Aura's
// ability body.
func OldB816UntapEnchantedCreature(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	host, ok := EnchantedCreatureAsItResolves(ctx)
	if !ok {
		return nil
	}
	return UntapTarget{Target: host}.Apply(ctx)
}
