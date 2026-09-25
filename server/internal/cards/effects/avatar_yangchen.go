package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Avatar Yangchen — Legendary Creature — Avatar, 4/5 (colorless):
//
//	"Flying
//	 Whenever you cast your second spell each turn, airbend up to one
//	 other target nonland permanent. (Exile it. While it's exiled, its
//	 owner may cast it for {2} rather than its mana cost.)"
//
// The BACK face of The Legend of Yangchen ("<oracle_id>#1", ADR 0034),
// reached only through chapter III's exile-and-return
// (the_legend_of_yangchen.go).
//
// "Your second spell each turn" is Breeches, the Blastmaker's clock:
// the cast path bumps the per-turn tally BEFORE it emits EventCast, so
// a total of exactly 2 means "this is the second" (game.CastTallyFor).
//
// "Up to one OTHER target nonland permanent" is Aang, the Last
// Airbender's shape with the count on the clause instead of an
// OptionalPrompt: this trigger has no printed "may" on whether it
// goes on the stack, only on how many targets it takes, so Min 0 on
// the target spec is what "up to one" means here. "Other" is exact:
// the clause is built per trigger through AnotherTarget
// (TriggeredAbility.TargetsFrom, which is handed the source), so the
// picker excludes THIS Yangchen by instance rather than relying on
// AirbendOtherTarget's resolution-time self-check to decline a pick
// that should never have been offered.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        theLegendOfYangchenOracleID + "#1",
		Name:            "Avatar Yangchen",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventCast},
			AppliesTo: YouCastYourSecondSpellEachTurn,
			TargetsFrom: AnotherTarget(func(other CardPredicate) *game.TargetSpec {
				return TargetPermanent("up to one other target nonland permanent", Nonland(), other).WithCount(0, 1)
			}),
			Key:    "Avatar Yangchen — airbend a nonland permanent",
			Effect: AirbendOtherTarget,
		}},
	})
}
