package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Croaking Counterpart — Sorcery {1}{G}{U} (Midnight Hunt promo):
//
//	"Create a token that's a copy of target non-Frog creature, except it's a
//	 1/1 green Frog.
//	 Flashback {3}{G}{U}"
//
// The copy takes the target's copiable values as they are on resolution
// (CR 707.2) with the exception applied (CR 707.9): base power and toughness
// 1/1, green and no other colour, and the creature type Frog in place of its
// creature types. Its card types, supertypes and any non-creature subtypes
// stay. The target is re-judged on resolution, and a creature that is now a
// Frog is no longer legal (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "142b2f24-49f6-4ef8-9731-9ae2e513f796",
		Name:             "Croaking Counterpart",
		Completeness:     effects.CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Flashback("{3}{G}{U}")},
		Targets:          effects.TargetCreature("target non-Frog creature", effects.Not(effects.Subtype("Frog"))),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				frog := ctx.TextSubtype("Frog")
				return effects.CreateTokenCopy{
					Controller: ctx.Controller(),
					Copy:       t.ID,
					N:          1,
					Except: func(tok *game.Card) {
						tok.Power, tok.Toughness = 1, 1
						tok.VariableToughness = false
						tok.PrintedPTKnown = true
						tok.SetCopyExceptionColors("G")
						tok.TypeLine = effects.S82MidB11TypeLineWithCreatureType(tok.TypeLine, frog)
					},
				}.Apply(ctx)
			}
			return nil
		},
	})
}
