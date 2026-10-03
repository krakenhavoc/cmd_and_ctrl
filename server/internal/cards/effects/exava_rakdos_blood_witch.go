package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Exava, Rakdos Blood Witch — Legendary Creature — Human Cleric
// {2}{B}{R}, 3/3:
//
//	"First strike, haste
//	 Unleash (You may have this creature enter with a +1/+1 counter on
//	 it. It can't block as long as it has a +1/+1 counter on it.)
//	 Each other creature you control with a +1/+1 counter on it has
//	 haste."
//
// Unleash is the engine's keyword (ADR 0109 §10); the haste is a layer-6
// grant read off the live counters, so a creature gains it the moment a
// +1/+1 counter lands and loses it when the last one comes off.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "524249cd-68d9-472a-89cd-5872641ca6de",
		Name:            "Exava, Rakdos Blood Witch",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike", "haste", game.KeywordUnleash},
		Static:          []game.StaticAbility{CreaturesYouControlWithCountersHave(game.CounterPlusOne, true, "haste")},
	})
}
