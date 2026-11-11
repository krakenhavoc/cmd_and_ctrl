package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Deathgorge Scavenger — Creature — Dinosaur {2}{G}, 3/2 (Lost Caverns of
// Ixalan Commander):
//
//	"Whenever this creature enters or attacks, you may exile target card
//	 from a graveyard. If a creature card is exiled this way, you gain 2
//	 life. If a noncreature card is exiled this way, this creature gets
//	 +1/+1 until end of turn."
//
// One trigger on two events. The "if" riders read what actually reached
// exile (a keyed tail told the exiled card, CR 608.2c), not what was
// targeted, so a card a replacement sent elsewhere pays nothing. A land
// or any other noncreature card pumps; an artifact creature card is a
// creature card, and gains life only.
func init() {
	const label = "Deathgorge Scavenger — exile target card from a graveyard"
	Register(Spec{
		OracleID:     "17da3654-7517-473b-89e7-0303f6f030cb",
		Name:         "Deathgorge Scavenger",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(Optional(WhenThisEntersOrAttacks(label, deathgorgeScavengerExile),
				"Deathgorge Scavenger — exile target card from a graveyard?"),
				TargetCardInGraveyard("target card from a graveyard")),
		},
	})
}

func deathgorgeScavengerExile(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		return ExileTarget{Target: t.ID, ThenTail: withCarry(ctx.ItemTail(deathgorgeScavengerExiled), t.ID)}.Apply(ctx)
	}
	return nil
}

var deathgorgeScavengerExiled = CardTail("deathgorge-scavenger/exiled-card-riders", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) != 1 {
		return nil
	}
	id := tailCarry0(o)
	if id == uuid.Nil {
		return nil
	}
	card, ok := ctx.Game.LookupCardForEffect(id)
	if !ok {
		return nil
	}
	if card.IsCreature() {
		return GainLife{Player: ctx.Controller(), Amount: 2}.Apply(ctx)
	}
	if onBattlefield(ctx.Game, ctx.Source()) {
		return BoostUntilEOT{Target: ctx.Source(), Power: 1, Toughness: 1, Label: "Deathgorge Scavenger — +1/+1"}.Apply(ctx)
	}
	return nil
})
