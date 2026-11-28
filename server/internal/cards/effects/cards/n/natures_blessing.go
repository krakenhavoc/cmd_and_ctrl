package cardsn

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nature's Blessing — Enchantment {2}{G}{W} (Alliances, S82 OLD batch 818):
//
//	"{G}{W}, Discard a card: Put a +1/+1 counter on target creature or that
//	 creature gains banding, first strike, or trample. (This effect lasts
//	 indefinitely.)"
//
// The target is chosen with the ability; the choice among the four is made as
// it resolves (Flowstone Sculpture's shape, with the target carried).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ae0f65b4-3814-45fe-a869-8585f6869278",
		Name:         "Nature's Blessing",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{G}{W}, Discard a card: Put a +1/+1 counter on target creature or that creature gains banding, first strike, or trample.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1, Toughness: 1}},
			Cost:    effects.Plus(effects.ManaCost("{G}{W}"), effects.DiscardN(1, "a card")),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.PickOption{
					Question: "Nature's Blessing — put a +1/+1 counter on it, or it gains…",
					Options:  []game.ChoiceOption{{Label: "A +1/+1 counter"}, {Label: "Banding"}, {Label: "First strike"}, {Label: "Trample"}},
					ThenKey:  natureSBlessingChoice,
					Carry:    []uuid.UUID{id},
					WithItem: true,
				}.Apply(ctx)
			},
		}},
	})
}

var natureSBlessingChoice = effects.OptionPickThen("option-pick/s82-old-b818-natures-blessing", func(ctx *effects.Context, r game.OptionPicked) error {
	if len(r.Carry) == 0 || r.Index < 0 || r.Index > 3 || !effects.OnBattlefield(ctx.Game, r.Carry[0]) {
		return nil
	}
	id := r.Carry[0]
	if r.Index == 0 {
		return effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
	}
	keyword := []string{"", "banding", "first strike", "trample"}[r.Index]
	return effects.ScopedEffectFor{
		Target:   id,
		Mods:     []game.Mod{game.AddKeywordsMod(keyword)},
		Duration: game.IndefiniteDuration(),
		Label:    "Nature's Blessing — " + keyword,
	}.Apply(ctx)
})
