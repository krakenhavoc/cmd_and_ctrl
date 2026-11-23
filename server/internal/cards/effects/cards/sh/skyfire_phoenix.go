package cardssh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Skyfire Phoenix — Creature — Phoenix {2}{R}{R}, 3/3 (Commander 2019,
// #5517):
//
//	"Flying, haste
//	 When you cast your commander, return this card from your graveyard to
//	 the battlefield."
//
// The trigger watches from the graveyard only (InGraveyard), on any cast of
// a commander its owner owns, from whatever zone the commander is cast (CR
// 903.3: "your commander" is a commander you own). The return is Cauldron
// Familiar's body: a Phoenix that has left the graveyard by the time the
// trigger resolves stays where it is.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8fdb3350-a6ff-4d02-a649-10663436c40f",
		Name:            "Skyfire Phoenix",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		Triggered: []game.TriggeredAbility{
			effects.InGraveyard(effects.WheneverYouCast(c19YourCommanderSpell,
				"Skyfire Phoenix — return it from your graveyard to the battlefield",
				effects.ReturnThisFromGraveyardToBattlefield)),
		},
	})
}

// c19YourCommanderSpell is "your commander" as a spell: a commander the
// caster owns.
func c19YourCommanderSpell(_ *game.Game, caster uuid.UUID, c game.Card) bool {
	return c.IsCommander && c.Owner == caster
}
