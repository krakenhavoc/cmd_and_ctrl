package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Capricious Hellraiser — Creature — Phyrexian Dragon {3}{R}{R}{R}, 4/4:
//
//	"This spell costs {3} less to cast if you have nine or more cards in
//	 your graveyard.
//	 Flying
//	 When this creature enters, exile three cards at random from your
//	 graveyard. Choose a noncreature, nonland card from among them and
//	 copy it. You may cast the copy without paying its mana cost."
//
// The discount is a generic-only SelfCostModifier read at CR 601.2f. The
// enters trigger draws three cards from the keyed random stream (all of
// them when the graveyard has fewer) and exiles them; the pick is over the
// noncreature, nonland cards that really reached exile (a card a
// replacement sent elsewhere is not among them), and none eligible means no
// copy. The copy is a card copy (CR 707.12) offered a free cast during the
// resolution (CR 608.2g); the original stays exiled, and a copy not cast
// ceases to exist.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "eb7b024c-da9b-4544-8648-0ac0395cf7fc",
		Name:            "Capricious Hellraiser",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		SelfCostModifiers: []game.CostModifier{
			CostsLess(3, "This spell costs {3} less to cast if you have nine or more cards in your graveyard.",
				capriciousHellraiserNineInGraveyard),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Capricious Hellraiser — exile three cards at random from your graveyard; copy a noncreature, nonland one and you may cast the copy",
				capriciousHellraiserEnters),
		},
	})
}

// capriciousHellraiserEnters exiles three cards at random from the graveyard.
func capriciousHellraiserEnters(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ids := graveyardIDs(g, ctx.Controller(), nil)
	pick := g.ChooseAtRandomForEffect(randomDraw(ctx), ids, 3)
	if len(pick) == 0 {
		return nil
	}
	return g.ExileCardsKeyedForEffect(pick, ctx.ItemTail(capriciousHellraiserExiled))
}

// capriciousHellraiserExiled is "choose a noncreature, nonland card from
// among them", over the cards that reached exile. An on-disk key.
var capriciousHellraiserExiled = CardTail("capricious-hellraiser/choose-and-copy", func(ctx *Context, o game.TailOutcome) error {
	var eligible []uuid.UUID
	for _, id := range o.IDs {
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok || c.IsCreature() || c.IsLand() {
			continue
		}
		if z := ctx.Game.FindCardZoneForEffect(id); z != nil && z.Kind == game.ZoneExile {
			eligible = append(eligible, id)
		}
	}
	if len(eligible) == 0 {
		return nil
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  ctx.Controller(),
		Source:   ctx.Source(),
		Question: "Capricious Hellraiser — choose a noncreature, nonland card exiled this way to copy",
		Cards:    eligible,
		Min:      1,
		Max:      1,
		LiveZone: game.ZoneExile,
		ThenKey:  capriciousHellraiserPicked,
	})
	return nil
})

// capriciousHellraiserPicked copies the chosen card and offers its free cast.
var capriciousHellraiserPicked = game.RegisterChooseCardsThen("choose-cards/capricious-hellraiser-copy", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Picked) != 1 {
		return nil
	}
	if z := g.FindCardZoneForEffect(r.Picked[0]); z == nil || z.Kind != game.ZoneExile {
		return nil
	}
	return CopyAndCast{Cards: r.Picked, Player: r.Chooser, Free: true}.Apply(keyedContext(g, r.Chooser, r.Source))
})
