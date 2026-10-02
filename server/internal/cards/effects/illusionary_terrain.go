package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Illusionary Terrain — Enchantment {U}{U}:
//
//	"Cumulative upkeep {2} (At the beginning of your upkeep, put an age
//	 counter on this permanent, then sacrifice it unless you pay its
//	 upkeep cost for each age counter on it.)
//	 As this enchantment enters, choose two basic land types.
//	 Basic lands of the first chosen type are the second chosen type."
//
// ADR 0109 owner decision 4 (#1881). The two types are chosen together as
// it enters (CR 614.12), as one ordered pair of two different types, the
// first named first ("Island, then Swamp"): twenty answers, each its own
// gated static (ChosenIs), so until the choice is made nothing changes.
// Then it is effects.SetsBasicLandType, the static form of CR 305.7, over
// every basic land of the first type, everyone's: in layer 4 its land types
// are replaced by the second (other subtypes stay, CR 205.1a), it loses
// the abilities its rules text gives it, and it taps for the second type's
// colour (CR 305.6). It stays basic, because a type change leaves
// supertypes alone. Cumulative upkeep {2} is CR 702.24.
//
// No simplification.
func init() {
	var words []string
	var statics []game.StaticAbility
	for _, from := range game.BasicLandTypes {
		for _, to := range game.BasicLandTypes {
			if from == to {
				continue
			}
			word := from + ", then " + to
			words = append(words, word)
			first := from
			statics = append(statics, StaticWhenChosen(word, SetsBasicLandType(
				func(target *game.Card, _ *game.Game, _ *game.Card) bool {
					return target.IsLand() && target.HasSupertype("basic") && target.HasSubtype(first)
				}, nil, []string{to})))
		}
	}
	Register(Spec{
		OracleID:     "17ae63fa-fd63-45e6-81c1-bf30851bf77c",
		Name:         "Illusionary Terrain",
		Completeness: CompletenessFull,
		AsEnters: func(card *game.Card, ctx *Context) error {
			ctx.Game.QueueChooseOptionAsEntersForEffect(card.Controller, card.InstanceID,
				"Illusionary Terrain — choose two basic land types: basic lands of the first are the second", words)
			return nil
		},
		Static: statics,
		Triggered: []game.TriggeredAbility{
			CumulativeUpkeep("Illusionary Terrain — cumulative upkeep {2}", "{2}"),
		},
	})
}
