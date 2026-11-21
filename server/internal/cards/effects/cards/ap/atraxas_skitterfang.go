package cardsap

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Atraxa's Skitterfang — Artifact Creature — Phyrexian Insect {3}, 2/2
// (Phyrexia: All Will Be One):
//
//	"This creature enters with three oil counters on it.
//	 At the beginning of combat on your turn, you may remove an oil
//	 counter from this creature. When you do, target creature you control
//	 gains your choice of flying, vigilance, deathtouch, or lifelink until
//	 end of turn."
//
// The removal is made as the trigger resolves; only a counter really
// removed makes the reflexive trigger (CR 603.12), which targets a
// creature you control as it goes on the stack. The keyword is chosen as
// the reflexive trigger resolves, and the grant lasts to end of turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8255b408-cc8c-453f-982c-f361b5559cec",
		Name:         "Atraxa's Skitterfang",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.B10EntersWithCounters(effects.S82OneOil, 3, "Atraxa's Skitterfang: enters with three oil counters"),
		},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.AtBeginningOfYourCombat("Atraxa's Skitterfang — you may remove an oil counter from it; when you do, target creature you control gains your choice of flying, vigilance, deathtouch, or lifelink until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					c, ok := g.LookupCardForEffect(item.SourceCardID)
					if !ok || !effects.OnBattlefield(g, item.SourceCardID) || effects.SourceIsNewObject(g, item) || c.Counters[effects.S82OneOil] < 1 {
						return nil
					}
					if err := g.AddCounterForEffect(item.SourceCardID, effects.S82OneOil, -1); err != nil {
						return err
					}
					return effects.WhenYouDo("Atraxa's Skitterfang — target creature you control gains your choice of flying, vigilance, deathtouch, or lifelink until end of turn", skitterfangBody).Apply(effects.NewContext(g, item))
				}),
				"Atraxa's Skitterfang — remove an oil counter from it?"),
		},
	})
}

var skitterfangKeywords = []string{"flying", "vigilance", "deathtouch", "lifelink"}

// skitterfangBody is the reflexive trigger: ask which keyword, then grant
// it. An on-disk key: never renamed, never reused.
var skitterfangBody = game.ReflexiveBody("atraxas-skitterfang/grant", effects.SimpleBody(skitterfangAsk),
	effects.ConstTargets(func() *game.TargetSpec {
		return effects.TargetCreature("target creature you control", effects.YouControl())
	}))

func skitterfangAsk(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	id, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	options := make([]game.ChoiceOption, len(skitterfangKeywords))
	for i, k := range skitterfangKeywords {
		options[i] = game.ChoiceOption{Label: k}
	}
	return effects.PickOption{
		Question: "Atraxa's Skitterfang — target creature gains your choice of keyword until end of turn",
		Options:  options,
		ThenKey:  skitterfangChosen,
		Carry:    []uuid.UUID{id},
	}.Apply(ctx)
}

var skitterfangChosen = effects.OptionPickThen("option-pick/atraxas-skitterfang-keyword", func(ctx *effects.Context, r game.OptionPicked) error {
	if r.Index < 0 || r.Index >= len(skitterfangKeywords) || len(r.Carry) == 0 {
		return nil
	}
	return effects.GrantKeywordUntilEOT{Target: r.Carry[0], Keywords: []string{skitterfangKeywords[r.Index]},
		Label: "Atraxa's Skitterfang — " + skitterfangKeywords[r.Index] + " until end of turn"}.Apply(ctx)
})
