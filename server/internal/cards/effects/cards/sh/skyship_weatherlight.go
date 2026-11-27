package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skyship Weatherlight — Legendary Artifact {4} (Planeshift, S82 OLD #5615):
//
//	"When Skyship Weatherlight enters, search your library for any number of
//	 artifact and/or creature cards, exile them, then shuffle.
//	 {4}, {T}: Choose a card at random that was exiled with Skyship
//	 Weatherlight. Put that card into its owner's hand."
//
// The search is Kaho's linked exile (CR 607.2a): any number from none to every
// match, linked to this object. The ability is Gorex's random pick over the
// linked cards still in exile (the engine's keyed random stream), into the
// owner's hand.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "57889a8e-0039-45c2-a608-ab315d51aa3a",
		Name:         "Skyship Weatherlight",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Skyship Weatherlight — search your library for any number of artifact and/or creature cards, exile them, then shuffle",
				func(g *game.Game, item *game.StackItem) error {
					ref, ok := g.SourceObjectForEffect(item)
					if !ok {
						return nil
					}
					return effects.SearchLibrary{
						Player:    item.Controller,
						Predicate: func(c game.Card) bool { return c.IsArtifact() || c.IsCreature() },
						Dest:      game.ZoneExile,
						Unbounded: true,
						Shuffle:   true,
						LinkTo:    game.PermissionCardRef(ref),
						Source:    item.SourceCardID,
						Reason:    "Skyship Weatherlight — any number of artifact and/or creature cards",
					}.Apply(effects.NewContext(g, item))
				}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}, {T}: Choose a card at random that was exiled with Skyship Weatherlight. Put that card into its owner's hand.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{4}"), effects.TapCost()),
			Effect:  effects.Do(effects.ReturnRandomCardExiledWithThis{}),
		}},
	})
}
