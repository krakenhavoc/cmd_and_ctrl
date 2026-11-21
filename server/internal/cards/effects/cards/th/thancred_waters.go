package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thancred Waters — Legendary Creature — Human Warrior {4}{W}, 3/5:
//
//	"Flash
//	 Royal Guard — When Thancred Waters enters, another target
//	 legendary permanent you control gains indestructible for as long as
//	 you control Thancred Waters.
//	 Whenever you cast a noncreature spell, Thancred Waters gains
//	 indestructible until end of turn."
//
// The ETB grant lasts "for as long as you control" the Thancred that
// entered (CR 611.2b, DurationWhileYouControlSource): a Thancred that
// left and returned before the trigger resolved is not the object the
// ability names, and nothing is granted. The target excludes Thancred
// himself, by instance, so a copy of him is still a legal target.
//
// No simplification.
func init() {
	royalGuard := effects.WithTargetsFrom(
		effects.WhenThisEnters("Thancred Waters — Royal Guard: another target legendary permanent you control gains indestructible for as long as you control Thancred Waters",
			func(g *game.Game, item *game.StackItem) error {
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
						Label:    "Thancred Waters — Royal Guard: indestructible",
					}.Apply(ctx)
				}
				return nil
			}),
		effects.AnotherTarget(func(other effects.CardPredicate) *game.TargetSpec {
			return effects.TargetPermanent("another target legendary permanent you control",
				effects.Legendary(), effects.YouControl(), other)
		}))
	effects.Register(effects.Spec{
		OracleID:        "dc36f81e-d1b0-44a4-bf82-4cd276bc0812",
		Name:            "Thancred Waters",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			royalGuard,
			effects.WheneverYouCast(effects.Noncreature(), "Thancred Waters — gains indestructible until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return effects.GrantKeywordUntilEOT{
						Target:   item.SourceCardID,
						Keywords: []string{"indestructible"},
						Label:    "Thancred Waters — indestructible",
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
