package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mistveil Plains — Land — Plains (EDHREC rank 3229):
//
//	"({T}: Add {W}.)
//	 This land enters tapped.
//	 {W}, {T}: Put target card from your graveyard on the bottom of
//	 your library. Activate only if you control two or more white
//	 permanents."
//
// Shadowmoor's white hybrid land: a repeatable tuck of your own
// graveyard card, for a commander deck the way to get a key spell back
// into the library. The {W} is the Plains type's intrinsic ability.
// "Activate only if you control two or more white permanents" is the
// activation condition (CR 602.1b, #743), ControlsAtLeast(2, white).
// The target is re-checked on resolution, so a card exiled from the
// graveyard in response is left alone.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bb5c1817-ac22-4779-9005-251bc354f181",
		Name:         "Mistveil Plains",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		Activated: []ActivatedAbility{{
			Label:     "{W}, {T}: Put target card from your graveyard on the bottom of your library. Activate only if you control two or more white permanents.",
			Cost:      Plus(ManaCost("{W}"), TapCost()),
			Targets:   TargetCardInGraveyard("target card from your graveyard", YouOwn()),
			Condition: ControlsAtLeast(2, MatchColor("W")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					if err := g.TuckToLibraryForEffect(t.ID, true); err != nil {
						return err
					}
				}
				return nil
			},
		}},
	})
}
