package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Darksteel Brute — Artifact {2} (Darksteel, #5605):
//
//	"Indestructible (Damage and effects that say "destroy" don't destroy
//	 this artifact.)
//	 {3}: This artifact becomes a 2/2 Beast artifact creature until end of
//	 turn."
//
// Indestructible is a printed keyword (CR 702.12). The animation is
// Guardian Idol's: it adds the creature type and the Beast subtype and
// sets its base power and toughness to 2/2 in layer 7b until end of turn
// (CR 613.4b), and it keeps being an artifact.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4f7d2774-2aec-4992-b35b-e03265e517af",
		Name:            "Darksteel Brute",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"indestructible"},
		Activated: []ActivatedAbility{{
			Label:   "{3}: This artifact becomes a 2/2 Beast artifact creature until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    ManaCost("{3}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !sourceIsStillThisPermanent(g, item) {
					return nil
				}
				ctx := NewContext(g, item)
				return ScopedEffectFor{
					Target: item.SourceCardID,
					Mods: append([]game.Mod{
						game.AddTypesMod("Creature"),
						game.AddSubtypesMod(ctx.TextSubtypes("Beast")...),
					}, game.SetBasePTMods(2, 2)...),
					Duration: DurationUntilEndOfTurn(ctx),
					Label:    "Darksteel Brute — a 2/2 Beast artifact creature until end of turn",
				}.Apply(ctx)
			},
		}},
	})
}
