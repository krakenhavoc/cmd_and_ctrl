package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Eriette's Tempting Apple — Legendary Artifact — Food {4} (Wilds of
// Eldraine, #4471):
//
//	"When Eriette's Tempting Apple enters, gain control of target creature
//	 until end of turn. Untap that creature. It gains haste until end of
//	 turn.
//	 {2}, {T}, Sacrifice Eriette's Tempting Apple: You gain 3 life.
//	 {2}, {T}, Sacrifice Eriette's Tempting Apple: Target opponent loses
//	 3 life."
//
// The entry trigger is Act of Treason's three clauses on a targeted
// trigger. The two Food abilities are separate printed abilities (you
// pick which one to pay for), the second targeting an opponent.
//
// No simplification.
func init() {
	enter := WhenThisEnters("Eriette's Tempting Apple — gain control of target creature until end of turn, untap it, haste",
		func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (GainControl{Target: t.ID, Duration: DurationUntilEndOfTurn(ctx),
					Label: "Eriette's Tempting Apple — gain control until end of turn"}).Apply(ctx); err != nil {
					return err
				}
				if err := (UntapTarget{Target: t.ID}).Apply(ctx); err != nil {
					return err
				}
				return GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"haste"},
					Label: "Eriette's Tempting Apple — haste until end of turn"}.Apply(ctx)
			}
			return nil
		})
	enter.Targets = TargetCreature("target creature")
	Register(Spec{
		OracleID:     "b26a2875-c704-491a-b82d-07c316e53611",
		Name:         "Eriette's Tempting Apple",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{enter},
		Activated: []ActivatedAbility{
			{
				Label:   "{2}, {T}, Sacrifice Eriette's Tempting Apple: You gain 3 life.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    Plus(ManaCost("{2}"), TapCost(), SacrificeThis()),
				Effect:  Do(GainLife{Amount: 3}),
			},
			{
				Label:   "{2}, {T}, Sacrifice Eriette's Tempting Apple: Target opponent loses 3 life.",
				Cost:    Plus(ManaCost("{2}"), TapCost(), SacrificeThis()),
				Targets: TargetPlayer("target opponent", Opponent()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind == game.TargetPlayer {
							return g.ChangePlayerLifeForEffect(item.SourceCardID, t.ID, -3)
						}
					}
					return nil
				},
			},
		},
	})
}
