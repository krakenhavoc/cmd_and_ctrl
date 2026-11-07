package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Queen's Bay Paladin — Creature — Vampire Knight {3}{B}{B}, 5/4 (The
// Lost Caverns of Ixalan, #4470):
//
//	"Whenever this creature enters or attacks, return up to one target
//	 Vampire card from your graveyard to the battlefield with a finality
//	 counter on it. You lose life equal to its mana value. (If a creature
//	 with a finality counter on it would die, exile it instead.)"
//
// One ability with two trigger conditions (WhenThisEntersOrAttacks, Sun
// Titan's reading). The return is Excava's: the finality counter is an
// entry counter (CR 614.1c), and the finality exile is the engine's
// (CR 122.1h). The life loss is the rest of the effect, a registered
// tail that runs once the entry is complete, so it follows the return
// in printed order even when an entering card asks a question first. The
// mana value is read as the card is in the graveyard, before it moves,
// so an {X} counts as 0 (CR 202.3e). With no target chosen, or a target
// that has left the graveyard by resolution (CR 608.2b), nothing returns
// and no life is lost. "Vampire" is a word of the card's text
// (CR 612.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "60b18d94-8ae5-4508-b4db-18f92c2d39fd",
		Name:         "Queen's Bay Paladin",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEntersOrAttacks("Queen's Bay Paladin — return up to one Vampire card with a finality counter; you lose life equal to its mana value", queensBayPaladinReturn),
				TargetCardInGraveyard("up to one target Vampire card from your graveyard", YouOwn(), Subtype("Vampire")).WithCount(0, 1)),
		},
	})
}

// queensBayPaladinReturn is the trigger's resolution: the return, then
// the life loss as its tail, carrying the mana value.
func queensBayPaladinReturn(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var id uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			id = t.ID
			break
		}
	}
	if id == uuid.Nil {
		return nil
	}
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		return nil
	}
	mv, _ := g.ManaValueForEffect(c)
	return putCardsFromZoneTogetherWith(ctx, game.ZoneGraveyard, []uuid.UUID{id}, game.ZoneEntryOptions{
		Controller:         item.Controller,
		EntersWithCounters: map[string]int{game.CounterFinality: 1},
	}, queensBayPaladinLoseLife, mv)
}

// queensBayPaladinLoseLife is "You lose life equal to its mana value",
// after the return. N is the mana value. An on-disk key: never renamed,
// never reused.
var queensBayPaladinLoseLife = TogetherTail("queens-bay-paladin/lose-life", func(ctx *Context, _ []uuid.UUID, n int) error {
	if n <= 0 {
		return nil
	}
	return GainLife{Amount: -n}.Apply(ctx)
})
