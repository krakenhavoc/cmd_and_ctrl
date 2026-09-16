package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The First Iroan Games — Enchantment — Saga for {2}{G}:
//
//	"I — Create a 1/1 white Human Soldier creature token.
//	 II — Put three +1/+1 counters on target creature you control.
//	 III — If you control a creature with power 4 or greater, draw
//	       two cards.
//	 IV — Create a Gold token."
//
// The only FOUR-chapter Saga in the S27 batch, and the reason it is
// here: the CR 704.5s sacrifice reads the final chapter off the
// declarations, so a card that stops at IV is what proves the engine
// is not hard-coded to III.
//
// Chapter III's "if you control a creature with power 4 or greater"
// is an intervening-if-shaped condition checked at RESOLUTION, not
// at trigger time — the printed wording is a conditional inside the
// ability, not an intervening-if clause (which would read "if you
// control …, draw two cards" ahead of the chapter marker). Reading
// it at resolution is therefore both simpler and right: a creature
// that grew in response counts.
func init() {
	Register(Spec{
		OracleID:     "58934f6d-1aa2-414c-85c6-955a1e26d675",
		Name:         "The First Iroan Games",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The Human Soldier token is created colorless instead of white, so anything that cares about a creature's color doesn't see it."},
		Triggered: []game.TriggeredAbility{
			ChapterTrigger(1, "The First Iroan Games — I: create a 1/1 Human Soldier", iroanSoldier),
			ChapterTriggerTargeting(2, "The First Iroan Games — II: three +1/+1 counters",
				TargetCreature("target creature you control", YouControl()),
				iroanCounters),
			ChapterTrigger(3, "The First Iroan Games — III: draw two if you control a 4-power creature",
				iroanDrawTwo),
			ChapterTrigger(4, "The First Iroan Games — IV: create a Gold token", iroanGold),
		},
	})
}

func iroanSoldier(g *game.Game, item *game.StackItem) error {
	return CreateToken{
		Controller: item.Controller,
		Template:   TokenCard("1/1 colorless Human Soldier"),
		N:          1,
	}.Apply(NewContext(g, item))
}

func iroanCounters(g *game.Game, item *game.StackItem) error {
	if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
		return nil
	}
	return AddCounter{
		Target: item.Targets[0].ID,
		Kind:   game.CounterPlusOne,
		N:      3,
	}.Apply(NewContext(g, item))
}

func iroanDrawTwo(g *game.Game, item *game.StackItem) error {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == item.Controller && c.IsCreature() && c.CurrentPower() >= 4 {
			return DrawCards{Player: item.Controller, N: 2}.Apply(NewContext(g, item))
		}
	}
	return nil
}

func iroanGold(g *game.Game, item *game.StackItem) error {
	return CreateToken{
		Controller: item.Controller,
		Template:   GoldToken(),
		N:          1,
	}.Apply(NewContext(g, item))
}
