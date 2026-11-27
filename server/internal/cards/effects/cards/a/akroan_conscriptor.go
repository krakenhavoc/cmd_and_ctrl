package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Akroan Conscriptor — {4}{R} Creature — Human Shaman 3/2 (Born of the Gods):
//
//	"Heroic — Whenever you cast a spell that targets this creature, gain control of another
//	 target creature until end of turn. Untap that creature. It gains haste until end of
//	 turn."
//
// Act of Treason's three steps on the heroic trigger's own target (chosen as the trigger
// goes on the stack; with no other creature the trigger is removed, CR 603.3d).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "54216953-c575-4e08-9bee-384721c0bc4d",
		Name:         "Akroan Conscriptor",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.Heroic("Akroan Conscriptor — gain control of another target creature until end of turn, untap it, it gains haste (heroic)",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					target, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					if err := (effects.GainControl{Target: target, Duration: effects.DurationUntilEndOfTurn(ctx),
						Label: "Akroan Conscriptor — gain control until end of turn"}).Apply(ctx); err != nil {
						return err
					}
					if err := (effects.UntapTarget{Target: target}).Apply(ctx); err != nil {
						return err
					}
					return effects.GrantKeywordUntilEOT{Target: target, Keywords: []string{"haste"},
						Label: "Akroan Conscriptor — haste until end of turn"}.Apply(ctx)
				}),
				effects.Another(effects.TargetCreature("another target creature"))),
		},
	})
}
