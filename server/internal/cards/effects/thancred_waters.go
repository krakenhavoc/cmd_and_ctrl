package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	royalGuard := withTargetsFrom(
		WhenThisEnters("Thancred Waters — Royal Guard: another target legendary permanent you control gains indestructible for as long as you control Thancred Waters",
			func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					d, ok := DurationWhileYouControlSource(ctx, item.SourceCardID, item.Controller)
					if !ok {
						return nil
					}
					return ScopedEffectFor{
						Target:   t.ID,
						Mods:     []game.Mod{game.AddKeywordsMod("indestructible")},
						Duration: d,
						Label:    "Thancred Waters — Royal Guard: indestructible",
					}.Apply(ctx)
				}
				return nil
			}),
		AnotherTarget(func(other CardPredicate) *game.TargetSpec {
			return TargetPermanent("another target legendary permanent you control",
				Legendary(), YouControl(), other)
		}))
	Register(Spec{
		OracleID:        "dc36f81e-d1b0-44a4-bf82-4cd276bc0812",
		Name:            "Thancred Waters",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			royalGuard,
			WheneverYouCast(Noncreature(), "Thancred Waters — gains indestructible until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return GrantKeywordUntilEOT{
						Target:   item.SourceCardID,
						Keywords: []string{"indestructible"},
						Label:    "Thancred Waters — indestructible",
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
