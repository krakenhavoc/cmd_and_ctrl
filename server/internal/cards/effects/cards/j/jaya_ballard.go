package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jaya Ballard — Legendary Planeswalker — Jaya {2}{R}{R}{R}, loyalty 5:
//
//	"+1: Add {R}{R}{R}. Spend this mana only to cast instant or sorcery
//	 spells.
//	 +1: Discard up to three cards, then draw that many cards.
//	 −8: You get an emblem with "You may cast instant and sorcery spells
//	 from your graveyard. If a spell cast this way would be put into your
//	 graveyard, exile it instead.""
//
// The first +1 is a loyalty ability, so it uses the stack and is not a mana
// ability; the mana carries the cast-an-instant-or-sorcery restriction and
// empties with the step. The emblem is a standing graveyard cast permission
// (EmblemSpec.CastPermissions, CR 114.3) with the exile-instead rider; it
// lasts the game (CR 114.2).
//
// No simplification.
func init() {
	const emblemText = "You may cast instant and sorcery spells from your graveyard. If a spell cast this way would be put into your graveyard, exile it instead."
	effects.Register(effects.Spec{
		OracleID:        "a56c5ca5-70eb-4d5f-8116-2acbe5f5a3cb",
		Name:            "Jaya Ballard",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 5,
		Emblem: &effects.EmblemSpec{
			Label: "Jaya Ballard emblem",
			Text:  emblemText,
			CastPermissions: []game.CastPermission{{
				Zone:                    game.ZoneGraveyard,
				CastOnly:                true,
				Filter:                  game.PermissionFilter{InstantOrSorceryOnly: true},
				ExileInsteadOfGraveyard: true,
				AltCostKey:              "jaya_ballard_emblem",
				Label:                   "Cast an instant or sorcery spell from your graveyard; if it would be put into your graveyard, exile it instead (Jaya Ballard emblem)",
			}},
		},
		Activated: []effects.ActivatedAbility{
			{
				Label: "+1: Add {R}{R}{R}. Spend this mana only to cast instant or sorcery spells.",
				Cost:  effects.LoyaltyCost(1),
				Effect: effects.Do(effects.AddMana{
					Produced:     "{R}{R}{R}",
					Restrictions: []string{effects.ManaRestrictCast, effects.ManaRestrictAnyType("Instant", "Sorcery")},
				}),
			},
			{
				Label: "+1: Discard up to three cards, then draw that many cards.",
				Cost:  effects.LoyaltyCost(1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.B39MayDiscardThenDraw(3, false, "Jaya Ballard — discard up to three cards, then draw that many",
						effects.DrawAsManyAsDiscarded)(effects.NewContext(g, item))
				},
			},
			{
				Label:  "−8: You get an emblem with \"" + emblemText + "\"",
				Cost:   effects.LoyaltyCost(-8),
				Effect: effects.Do(effects.CreateEmblem{}),
			},
		},
	})
}
