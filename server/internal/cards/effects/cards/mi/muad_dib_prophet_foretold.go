package cardsmi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Muad'Dib, Prophet Foretold — Legendary Creature — Human Noble Warrior
// {G}{U}{R}, 3/3:
//
//	"You may look at the top card of your library any time.
//	 Landfall — Whenever a land you control enters, you may mill a card.
//	 Delirium — As long as there are four or more card types among cards in
//	 your graveyard, you may play the top card of your library."
//
// The look is private (LibraryTopOwner). The landfall is an optional mill of
// one, asked as it resolves (CR 603.5). The delirium clause is Augur of
// Autumn's gated permission: the standing "play the top card" (lands and
// spells, CR 305.1, 401.5) is offered only while the graveyard holds four card
// types (Dragon's Rage Channeler's count), read live each time it is asked, so
// it opens as the graveyard fills and closes when it empties. Ordinary timing
// applies.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:          "61a6317a-6abb-4899-a7e4-36035042f8e4",
		Name:              "Muad'Dib, Prophet Foretold",
		Completeness:      effects.CompletenessFull,
		LibraryTopVisible: game.LibraryTopOwner,
		Triggered: []game.TriggeredAbility{
			effects.Optional(
				effects.Landfall("Muad'Dib, Prophet Foretold — you may mill a card (landfall)", effects.Do(effects.MillCards{N: 1})),
				"Muad'Dib, Prophet Foretold — mill a card?"),
		},
		GatedCastPermissions: []game.CastPermissionGate{{
			Permission: effects.PlayFromTopOfYourLibrary(game.PermissionFilter{},
				"Delirium — play the top card of your library (Muad'Dib, Prophet Foretold)"),
			Condition: func(g *game.Game, controller, _ uuid.UUID) bool {
				return effects.DrcDeliriumMet(g, controller)
			},
		}},
	})
}
