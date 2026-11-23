package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Relive the Past — Sorcery {5}{G}{W} (Murders at Karlov Manor, #4469):
//
//	"Return up to one target artifact card, up to one target land card,
//	 and up to one target non-Aura enchantment card from your graveyard
//	 to the battlefield. They are 5/5 Elemental creatures in addition to
//	 their other types."
//
// Three "up to one" clauses (Clauses), returned together as one entry
// (ReturnFromGraveyardTogether) with Excava's "in addition to its other
// types" overwrite (ADR 0210, game.EntersAs): layer 4 adds Creature and
// Elemental, layer 7b sets 5/5. The set lasts while that object is on
// the battlefield. A target that left the graveyard is skipped (CR
// 608.2b). A card that is both an artifact and an enchantment may only be
// chosen for one clause, as the targets must differ.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dcc51bcb-a61e-471f-8a01-7fcbc069fcba",
		Name:         "Relive the Past",
		Completeness: effects.CompletenessFull,
		Targets: effects.Clauses(
			effects.TargetCardInGraveyard("up to one target artifact card from your graveyard", effects.YouOwn(), effects.Artifact()).WithCount(0, 1),
			effects.TargetCardInGraveyard("up to one target land card from your graveyard", effects.YouOwn(), effects.Land()).WithCount(0, 1),
			effects.TargetCardInGraveyard("up to one target non-Aura enchantment card from your graveyard",
				effects.YouOwn(), effects.And(effects.Enchantment(), effects.Not(effects.Subtype("Aura")))).WithCount(0, 1),
		),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			ids := effects.LegalTargetCardIDs(ctx)
			if len(ids) == 0 {
				return nil
			}
			return effects.ReturnFromGraveyardTogether{
				Targets:    ids,
				Controller: item.Controller,
				EntersAs: &game.EntersAs{
					AddTypes:    []string{"Creature"},
					AddSubtypes: ctx.TextSubtypes("Elemental"),
					BasePT:      &game.EntersAsPT{Power: 5, Toughness: 5},
					Label:       "Relive the Past — a 5/5 Elemental creature in addition to its other types",
				},
			}.Apply(ctx)
		},
	})
}
