package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Little Bear — Creature — Bear {2}{G} (The Hobbit), 3/2:
//
//	"Flash
//	 When this creature enters, untap another target creature you control.
//	 If that creature is a Bear, put a +1/+1 counter on it."
//
// #4454, S76. Charmed Clothier's "another target creature you control"
// enters trigger. The untap is the sentence's first clause and the counter
// its second, read once the creature has been untapped: "Bear" is a word of
// this card's text (CR 612.2), so a text change moves it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "53c6c6bf-0256-424d-b937-1cc7d9e2d574",
		Name:            "Little Bear",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Little Bear — untap another target creature you control; +1/+1 counter if it is a Bear", littleBearUntap),
				effects.Another(effects.TargetCreature("another target creature you control", effects.YouControl()))),
		},
	})
}

func littleBearUntap(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	id, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	if err := (effects.UntapTarget{Target: id}).Apply(ctx); err != nil {
		return err
	}
	c, found := g.LookupCardForEffect(id)
	if !found || !effects.OnBattlefield(g, id) || !c.HasSubtype(ctx.TextSubtype("Bear")) {
		return nil
	}
	return effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
}
