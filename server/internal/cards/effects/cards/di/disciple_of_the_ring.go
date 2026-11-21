package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Disciple of the Ring — Creature — Human Wizard {3}{U}{U}, 3/4 (S82 GAP, #5948):
//
//	"{1}, Exile an instant or sorcery card from your graveyard: Choose one —
//	 • Counter target noncreature spell unless its controller pays {2}.
//	 • This creature gets +1/+1 until end of turn.
//	 • Tap target creature.
//	 • Untap target creature."
//
// Izzet Charm's counter-unless-paid and Kargan Intimidator's modal activated
// ability. The mode is chosen on activation and only the chosen mode's target
// is chosen (CR 700.2b).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f06209e6-475e-47e0-b84c-3ec1215633a9",
		Name:         "Disciple of the Ring",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, Exile an instant or sorcery card from your graveyard: Choose one — counter target noncreature spell unless its controller pays {2}; this creature gets +1/+1 until end of turn; tap target creature; untap target creature.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost: effects.Plus(effects.ManaCost("{1}"),
				effects.ExileFromGraveyard(1, "an instant or sorcery card", func(c game.Card) bool {
					return c.IsInstant() || c.IsSorcery()
				})),
			Modes: effects.ChooseOne(
				effects.Mode("Counter target noncreature spell unless its controller pays {2}.",
					effects.TargetSpell("target noncreature spell", effects.Noncreature())),
				effects.Mode("This creature gets +1/+1 until end of turn."),
				effects.Mode("Tap target creature.", effects.TargetCreature("target creature")),
				effects.Mode("Untap target creature.", effects.TargetCreature("target creature")),
			),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				first := func() (game.TargetRef, bool) {
					if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard || !ctx.IsTargetLegal(item.Targets[0]) {
						return game.TargetRef{}, false
					}
					return item.Targets[0], true
				}
				switch {
				case ctx.HasMode(0):
					t, ok := first()
					if !ok {
						return nil
					}
					return effects.CounterUnlessPaid{StackID: t.ID, Cost: "{2}",
						Question: "Disciple of the Ring — pay {2} or your spell is countered?"}.Apply(ctx)
				case ctx.HasMode(1):
					if effects.SourceIsNewObject(g, item) {
						return nil
					}
					return effects.BoostUntilEOT{Target: item.SourceCardID, Power: 1, Toughness: 1,
						Label: "Disciple of the Ring — +1/+1 until end of turn"}.Apply(ctx)
				case ctx.HasMode(2):
					t, ok := first()
					if !ok {
						return nil
					}
					return effects.TapTarget{Target: t.ID}.Apply(ctx)
				case ctx.HasMode(3):
					t, ok := first()
					if !ok {
						return nil
					}
					return effects.UntapTarget{Target: t.ID}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
