package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Threats Around Every Corner — Enchantment {3}{G}:
//
//	"When this enchantment enters, manifest dread.
//	 Whenever a face-down permanent you control enters, search your
//	 library for a basic land card, put it onto the battlefield tapped,
//	 then shuffle."
//
// Any face-down permanent counts, not only a creature, so the
// trigger reads the entering card's face-down state and nothing else.
// The enchantment's own manifest enters after it, so it fetches a land
// for its own dread.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5c104ac8-738a-4a9d-a68b-5706018d2061",
		Name:         "Threats Around Every Corner",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, Self, "Threats Around Every Corner — manifest dread", Do(ManifestDread{})),
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, false)
				return ok && c.FaceDown
			}, "Threats Around Every Corner — search for a basic land card, put it onto the battlefield tapped",
				func(g *game.Game, item *game.StackItem) error {
					return SearchLibrary{
						Player:        item.Controller,
						Predicate:     IsBasicLand,
						Dest:          game.ZoneBattlefield,
						Limit:         1,
						Reveal:        true,
						Shuffle:       true,
						TappedOnEntry: true,
						Source:        item.SourceCardID,
						Reason:        "Threats Around Every Corner — a basic land card, onto the battlefield tapped",
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
