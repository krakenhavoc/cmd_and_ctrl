package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const gigantoplasmGrant = "gigantoplasm/set-base-pt"

// Gigantoplasm — Creature — Shapeshifter {3}{U}, 0/0 (Ravnica Remastered,
// #5948):
//
//	"You may have this creature enter as a copy of any creature on the
//	 battlefield, except it has '{X}: This creature has base power and
//	 toughness X/X.'"
//
// Colossal Clone Master's clone: the copy is the replacement effect of CR
// 614.1c, optional, over any creature on the battlefield, and the "except" is
// part of the copiable values (CR 707.9a): the granted ability is a bundle in
// the copy's values, so a Clone of the copy has it too. The ability sets the
// base power and toughness (layer 7b, CR 613.4b) with no duration, from the
// X paid as it was activated, pinned to this object as it resolves (CR 611.2c).
// Declined, or with no creature to copy, it is a 0/0 and dies.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "52fdfe4b-10a7-421d-80ea-f6c39013062b",
		Name:         "Gigantoplasm",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key:  gigantoplasmGrant,
			Text: "{X}: This creature has base power and toughness X/X.",
			Activated: []effects.ActivatedAbility{{
				Label:   "{X}: This creature has base power and toughness X/X.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    effects.ManaCost("{X}"),
				Effect:  gigantoplasmSetBasePT,
			}},
		}},
		Replacements: []game.ReplacementEffect{
			effects.EntersAsCopyOf(
				"Gigantoplasm",
				effects.AnyCreatureOnBattlefield,
				func(_ *game.ReplacementEvent, v *game.PrintedValues, _ *game.Game, _ *game.Card) {
					v.GrantAbility(gigantoplasmGrant)
				},
			),
		},
	})
}

// gigantoplasmSetBasePT is "This creature has base power and toughness X/X":
// a layer 7b effect from the resolution, pinned to the source (CR 611.2c) and
// lasting as long as it stays on the battlefield.
func gigantoplasmSetBasePT(g *game.Game, item *game.StackItem) error {
	if !effects.SourceIsStillThisPermanent(g, item) {
		return nil
	}
	ctx := effects.NewContext(g, item)
	x := ctx.X()
	if x < 0 {
		x = 0
	}
	return effects.ScopedEffectFor{
		Target:   item.SourceCardID,
		Mods:     game.SetBasePTMods(x, x),
		Duration: game.IndefiniteDuration(),
		Label:    "Gigantoplasm — base power and toughness X/X",
	}.Apply(ctx)
}
