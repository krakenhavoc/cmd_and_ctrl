package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Intellect Devourer — {3}{B} Creature — Horror 2/4
//
//	"Devour Intellect — When this creature enters, each opponent exiles a
//	 card from their hand until this creature leaves the battlefield.
//	 Body Thief — You may play lands and cast spells from among cards exiled
//	 with this creature. If you cast a spell this way, you may spend mana as
//	 though it were mana of any color to cast it."
//
// The two abilities are linked (CR 607.2a): the exile is stamped with this
// object, and a standing permission over exile, narrowed to the cards exiled
// with this source, lets its controller play them (a land uses the land
// drop) with mana of any colour for a spell. When the Devourer leaves, the
// cards go back to their owners' hands (CR 610.3), where they are no longer
// "exiled with" it.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a06fa827-900b-46d4-8f48-b4e15964951a",
		Name:         "Intellect Devourer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Devour Intellect — each opponent exiles a card from their hand until this creature leaves the battlefield",
				effects.GapB213ExileFromHandUntilLeaves),
		},
		CastPermissions: []game.CastPermission{{
			Zone:     game.ZoneExile,
			Filter:   game.PermissionFilter{ExiledWithSource: true},
			AnyColor: true,
			Label:    "Body Thief — play lands and cast spells from among cards exiled with Intellect Devourer, spending mana as though it were mana of any color (Intellect Devourer)",
		}},
	})
}
