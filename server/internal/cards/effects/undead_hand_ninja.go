package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Undead Hand Ninja — Creature — Zombie Ninja Villain {1}{B}, 2/1 (Marvel
// Super Heroes Commander, #4455):
//
//	"Deathtouch
//	 Whenever one or more creature cards leave your graveyard, each
//	 opponent loses 1 life and you gain 1 life."
//
// Desecrated Tomb's batch trigger (OncePerBatch over the creature-card
// zone moves out of your graveyard) with the drain of Cauldron Familiar's enters
// trigger (EachOpponentLosesLife, GainLife).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3c02d725-a05e-4086-9caa-ddb005d7ae18",
		Name:            "Undead Hand Ninja",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventZoneMove,
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return b02CreatureCardLeftYourGraveyard(ev, source, g)
				},
				"Undead Hand Ninja — each opponent loses 1 life and you gain 1 life",
				Do(EachOpponentLosesLife{N: 1}, GainLife{Amount: 1}))),
		},
	})
}
