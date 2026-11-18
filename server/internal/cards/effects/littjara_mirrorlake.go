package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Littjara Mirrorlake — Land:
//
//	"This land enters tapped.
//	 {T}: Add {U}.
//	 {2}{G}{G}{U}, {T}, Sacrifice this land: Create a token that's a copy
//	 of target creature you control, except it enters with an additional
//	 +1/+1 counter on it. Activate only as a sorcery."
//
// The sacrifice is part of the cost, so it is paid on activation and a
// Stifle on the ability leaves the land already gone. The token is a
// copy of the copiable values only (CR 707.2) and enters with the one
// counter as part of its entry (Ochre Jelly's TokenEntryOptions), so a
// counter-doubling replacement and "enters with counters" payoffs see
// it. A target that has left by resolution makes no token (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3b577179-10d0-43f1-ac17-9dd2d12c965d",
		Name:         "Littjara Mirrorlake",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{U}",
			Label:    "Add {U}",
		}},
		Activated: []ActivatedAbility{{
			Label:        "{2}{G}{G}{U}, {T}, Sacrifice this land: Create a token that's a copy of target creature you control, except it enters with an additional +1/+1 counter on it. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{2}{G}{G}{U}"), TapCost(), SacrificeThis()),
			Targets:      TargetCreature("target creature you control", YouControl()),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				target := firstLegalCardTarget(ctx)
				if target == uuid.Nil {
					return nil
				}
				tmpl, ok := TokenCopyTemplate(g, target)
				if !ok {
					return nil
				}
				_, err := g.CreateTokensForEffect(item.Controller, copiedToken(tmpl), 1,
					game.TokenEntryOptions{Counters: map[string]int{game.CounterPlusOne: 1}})
				return err
			},
		}},
	})
}
