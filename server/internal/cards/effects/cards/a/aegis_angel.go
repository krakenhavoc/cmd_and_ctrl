package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aegis Angel — Creature — Angel {4}{W}{W}, 5/5 (Magic Origins, S82 older
// sets):
//
//	"Flying
//	 When this creature enters, another target permanent gains indestructible
//	 for as long as you control this creature."
//
// A layer 6 keyword grant on the chosen permanent under CR 611.2b's "for as long
// as you control ~" duration (DurationWhileYouControlSource, Mind Flayer's): it
// ends the moment the Angel leaves or changes hands. An Angel that has already
// left, or changed hands, by the time the trigger resolves grants nothing. The
// target is "another" by object identity, so the Angel can't target itself, and
// it is judged again as the trigger resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5902e074-7154-4fc7-8776-b87e8b1925e7",
		Name:            "Aegis Angel",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Aegis Angel — another target permanent gains indestructible for as long as you control this creature", aegisAngelGrant),
				effects.Another(effects.TargetPermanent("another target permanent"))),
		},
	})
}

func aegisAngelGrant(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		d, ok := effects.DurationWhileYouControlSource(ctx, item.SourceCardID, item.Controller)
		if !ok {
			return nil
		}
		return effects.ScopedEffectFor{
			Target:   t.ID,
			Mods:     []game.Mod{game.AddKeywordsMod("indestructible")},
			Duration: d,
			Label:    "Aegis Angel — indestructible for as long as you control Aegis Angel",
		}.Apply(ctx)
	}
	return nil
}
