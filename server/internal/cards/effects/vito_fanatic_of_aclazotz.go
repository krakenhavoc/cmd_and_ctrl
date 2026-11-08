package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vito, Fanatic of Aclazotz — Legendary Creature — Vampire Demon
// {2}{W}{B}, 4/4 (The Lost Caverns of Ixalan, #4470):
//
//	"Flying
//	 Whenever you sacrifice another permanent, you gain 2 life if this is
//	 the first time this ability has resolved this turn. If it's the
//	 second time, each opponent loses 2 life. If it's the third time,
//	 create a 4/3 white and black Vampire Demon creature token with
//	 flying."
//
// The trigger is Pirate Peddlers' "you sacrifice another permanent"
// (EventSacrifice by the controller, any permanent but Vito). Which
// sentence applies is Belladonna Took's read: the engine's per-object
// tally of how many times this ability has resolved this turn, counting
// the resolution in progress. A fourth or later resolution does nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9768d43b-8b3f-4d03-ac1b-e603af832875",
		Name:            "Vito, Fanatic of Aclazotz",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			On(game.EventSacrifice, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Actor == source.Controller && ev.CardID != uuid.Nil && ev.CardID != source.InstanceID
			}, vitoFanaticLabel, vitoFanaticResolves),
		},
	})
}

const vitoFanaticLabel = "Vito, Fanatic of Aclazotz — gain 2 life, then each opponent loses 2 life, then a 4/3 Vampire Demon"

// vitoFanaticResolves picks the sentence by how many times the ability
// has resolved this turn.
func vitoFanaticResolves(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	switch g.ResolvedThisTurn(item.SourceCardID, vitoFanaticLabel) {
	case 1:
		return GainLife{Player: item.Controller, Amount: 2}.Apply(ctx)
	case 2:
		return EachOpponentLosesLife{N: 2}.Apply(ctx)
	case 3:
		return CreateToken{Template: TokenCard("4/3 white and black Vampire Demon with flying"), N: 1}.Apply(ctx)
	}
	return nil
}
