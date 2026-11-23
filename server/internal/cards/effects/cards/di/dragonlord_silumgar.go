package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dragonlord Silumgar — Legendary Creature — Elder Dragon {4}{U}{B}, 3/5
// (Dragons of Tarkir):
//
//	"Flying, deathtouch
//	 When Dragonlord Silumgar enters, gain control of target creature or
//	 planeswalker for as long as you control Dragonlord Silumgar."
//
// Flying and deathtouch are printed keywords. Mind Flayer's duration, CR
// 611.2b's "for as long as you control this creature": the permanent goes
// home when Silumgar leaves AND when somebody else takes Silumgar. A
// Silumgar that has already left, or changed hands, by the time the trigger
// resolves takes nothing: the effect's condition is false as it would
// begin, so it never begins.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "44581d4f-746b-43b8-a34d-16dc383ab70e",
		Name:            "Dragonlord Silumgar",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "deathtouch"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Dragonlord Silumgar — gain control of target creature or planeswalker for as long as you control Dragonlord Silumgar", silumgarSteal),
				effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker()))),
		},
	})
}

func silumgarSteal(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		d, ok := effects.DurationWhileYouControlSource(ctx, item.SourceCardID, item.Controller)
		if !ok {
			return nil
		}
		return effects.GainControl{
			Target:     t.ID,
			Controller: item.Controller,
			Duration:   d,
			Label:      "Dragonlord Silumgar — control for as long as you control Dragonlord Silumgar",
		}.Apply(ctx)
	}
	return nil
}
