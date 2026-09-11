package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Eldest Reborn — Enchantment — Saga for {4}{B}:
//
//	"I — Each opponent sacrifices a creature or planeswalker of
//	     their choice.
//	 II — Each opponent discards a card.
//	 III — Put target creature or planeswalker card from a graveyard
//	       onto the battlefield under your control."
//
// Chapter I is EachPlayerSacrifices, not a loop: "of their choice"
// is the rules content, and the primitive fans one prompt out per
// opponent so nobody picks for anyone else. It does not target, so
// hexproof and protection are irrelevant to it.
//
// Chapter III says "a graveyard", not "your graveyard" — the target
// clause is deliberately unscoped by owner, and the reanimated
// permanent lands under the SAGA's controller, which is what
// ReturnFromGraveyard's Controller field is for. Leaving it zero
// would hand an opponent's creature straight back to them.
func init() {
	Register(Spec{
		OracleID: "44ed4c0c-a012-4895-a547-b04150553bba",
		Name:     "The Eldest Reborn",
		Triggered: []game.TriggeredAbility{
			ChapterTrigger(1, "The Eldest Reborn — I: each opponent sacrifices a creature or planeswalker",
				eldestRebornSacrifice),
			ChapterTrigger(2, "The Eldest Reborn — II: each opponent discards a card",
				eldestRebornDiscard),
			ChapterTriggerTargeting(3, "The Eldest Reborn — III: reanimate under your control",
				TargetCardInGraveyard("target creature or planeswalker card in a graveyard",
					Or(Creature(), Planeswalker())),
				eldestRebornReanimate),
		},
	})
}

func eldestRebornSacrifice(g *game.Game, item *game.StackItem) error {
	return EachPlayerSacrifices{
		ExceptController: true,
		Match:            Or(Creature(), Planeswalker()),
		Label:            "a creature or planeswalker",
	}.Apply(NewContext(g, item))
}

func eldestRebornDiscard(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, opp := range ctx.Opponents() {
		g.DiscardChoiceForEffect(opp, 1)
	}
	return nil
}

func eldestRebornReanimate(g *game.Game, item *game.StackItem) error {
	if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
		return nil
	}
	return ReturnFromGraveyard{
		Target:     item.Targets[0].ID,
		Dest:       game.ZoneBattlefield,
		Controller: item.Controller,
	}.Apply(NewContext(g, item))
}
