package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Zealous Conscripts — Creature — Human Warrior {4}{R}, 3/3:
//
//	"Haste
//	 When this creature enters, gain control of target permanent until end
//	 of turn. Untap that permanent. It gains haste until end of turn."
//
// The same three primitives as Act of Treason, in printed order, over any
// permanent: a layer-2 until-end-of-turn control effect, an untap and a
// layer-6 haste grant.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1dae6f39-2cbd-485c-b190-017a26401fd4",
		Name:            "Zealous Conscripts",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Zealous Conscripts — gain control of target permanent until end of turn, untap it, it gains haste",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard {
							continue
						}
						if err := (effects.GainControl{
							Target:   t.ID,
							Duration: effects.DurationUntilEndOfTurn(ctx),
							Label:    "Zealous Conscripts — gain control until end of turn",
						}).Apply(ctx); err != nil {
							return err
						}
						if err := (effects.UntapTarget{Target: t.ID}).Apply(ctx); err != nil {
							return err
						}
						return effects.GrantKeywordUntilEOT{
							Target:   t.ID,
							Keywords: []string{"haste"},
							Label:    "Zealous Conscripts — haste until end of turn",
						}.Apply(ctx)
					}
					return nil
				}),
				effects.TargetPermanent("target permanent")),
		},
	})
}
