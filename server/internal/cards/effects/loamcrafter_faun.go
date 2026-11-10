package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Loamcrafter Faun — Creature — Satyr Druid {2}{G}, 3/3 (Foundations
// Commander):
//
//	"When this creature enters, you may discard one or more land cards.
//	 When you do, return up to that many target nonland permanent cards
//	 from your graveyard to your hand."
//
// Awaken the Honored Dead's optional discard and reflexive trigger, with
// the count: the discard is "up to" every land card in hand with a set rule
// that refuses any card that is not a land (cmdrB9LandsOnly), and the
// reflexive trigger carries the number really discarded as Params.Amount,
// which is its "up to X targets" (Campsite Cuisine's shape). Its targets
// are chosen as it goes on the stack, as the ruling says. Declining, or
// holding no land card, discards nothing and makes no trigger.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2fbfdbcc-5f89-47b2-bd8d-49c797e0624f",
		Name:         "Loamcrafter Faun",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Loamcrafter Faun — you may discard one or more land cards; when you do, return up to that many target nonland permanent cards from your graveyard to your hand",
				loamcrafterFaunEnters),
		},
	})
}

func loamcrafterFaunEnters(g *game.Game, item *game.StackItem) error {
	p := g.PlayerByIDForEffect(item.Controller)
	if p == nil {
		return nil
	}
	lands := 0
	for _, c := range p.Hand.Cards {
		if c.IsLand() {
			lands++
		}
	}
	if lands == 0 {
		return nil
	}
	return g.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
		Player:   item.Controller,
		Source:   item.SourceCardID,
		N:        lands,
		UpTo:     true,
		Question: "Loamcrafter Faun — you may discard one or more land cards",
	}, loamcrafterFaunDiscarded, nil)
}

// loamcrafterFaunDiscarded makes the reflexive trigger for the land cards
// really discarded. An on-disk key.
var loamcrafterFaunDiscarded = game.RegisterDiscardThenRule("choose-cards/loamcrafter-faun-discard", cmdrB9LandsOnly,
	func(g *game.Game, r game.DiscardsDone) error {
		n := len(r.Discarded.Cards())
		if n == 0 {
			return nil
		}
		t := WhenYouDo("Loamcrafter Faun — return up to that many target nonland permanent cards from your graveyard to your hand", loamcrafterFaunReturnBody)
		t.Params = game.EffectParams{Amount: n}
		return t.Apply(keyedContext(g, r.Player, r.Source))
	})

// loamcrafterFaunReturnBody returns each still-legal target to the hand.
// Its clause is "up to X target nonland permanent cards from your
// graveyard", X being Params.Amount.
var loamcrafterFaunReturnBody = game.ReflexiveBody("loamcrafter-faun/return-to-hand", simpleBody(func(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := (ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneHand}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}), func(_ uuid.UUID, p game.EffectParams) *game.TargetSpec {
	return TargetCardInGraveyard("up to X target nonland permanent cards from your graveyard", YouOwn(), Nonland(), cmdrB9PermanentCard()).WithCount(0, p.Amount)
})
