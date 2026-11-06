package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// avatar_b4_helpers.go — shared helpers for Avatar: The Last Airbender,
// batch 4 (#4459).

// avB4LessonsInGraveyard counts the Lesson cards in a player's
// graveyard. Lesson is a spell subtype on the printed type line
// ("Sorcery — Lesson"), which is what isLessonCard reads (learn.go).
func avB4LessonsInGraveyard(g *game.Game, player uuid.UUID) int {
	p := g.PlayerByIDForEffect(player)
	if p == nil || p.Graveyard == nil {
		return 0
	}
	n := 0
	for _, c := range p.Graveyard.Cards {
		if isLessonCard(c) {
			n++
		}
	}
	return n
}

// avB4LessonSelfPT is "This creature gets +P/+T as long as there's a
// Lesson card in your graveyard" (First-Time Flyer): a layer 7c modify
// on the source alone, read live on every layer pass.
func avB4LessonSelfPT(power, toughness int) game.StaticAbility {
	return game.StaticAbility{
		Label:    "Gets " + signedPT(power, toughness) + " as long as there's a Lesson card in your graveyard.",
		Layer:    game.Layer7PT,
		SubLayer: game.SubLayer7C_Modify,
		AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
			return target.InstanceID == source.InstanceID && avB4LessonsInGraveyard(g, source.Controller) >= 1
		},
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.Power += power
			c.Toughness += toughness
		},
	}
}

// avB4ThreeLessonsInYourGraveyard is a cost predicate for "as long as
// there are three or more Lesson cards in your graveyard" (Gran-Gran),
// where "your" is the controller of the permanent granting the
// reduction, read as the spell is priced.
func avB4ThreeLessonsInYourGraveyard() CostPredicate {
	return func(q game.CostQuery) bool {
		return avB4LessonsInGraveyard(q.Game, q.Source.Controller) >= 3
	}
}

// avB4SetAttachedName is "enchanted creature is … named <name>" from a
// static ability (Honest Work's Humble Merchant): a layer 3 effect
// (CR 612.8, "loses any names it had and has only the specified name"),
// not a copy effect, so a copy of the creature still copies its printed
// name (CR 707.2).
func avB4SetAttachedName(name string) game.StaticAbility {
	return game.StaticAbility{
		Label:     "Enchanted creature is named " + name + ".",
		Layer:     game.Layer3Text,
		AppliesTo: AttachedToSource,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.Name = name
		},
	}
}

// tapEnchantedAndRemoveAllCounters is "When this Aura enters, tap
// enchanted creature and remove all counters from it" (Enchanted
// River's Grasp, Honest Work). The Aura's attachment is read as the
// trigger resolves, so an Aura that has left in response still taps the
// creature it last enchanted (CR 608.2h).
func tapEnchantedAndRemoveAllCounters(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	aura, ok := ctx.TriggeringPermanent()
	if !ok || aura.AttachedTo.Kind != game.TargetCard {
		return nil
	}
	if err := (TapTarget{Target: aura.AttachedTo.ID}).Apply(ctx); err != nil {
		return err
	}
	return removeEveryCounterFrom(ctx, aura.AttachedTo.ID)
}
