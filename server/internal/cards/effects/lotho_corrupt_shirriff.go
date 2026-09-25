package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lotho, Corrupt Shirriff — Legendary Creature — Halfling Rogue
// {W}{B}, 2/1:
//
//	"Whenever a player casts their second spell each turn, you lose 1
//	 life and create a Treasure token."
//
// "A player" is every seat, Lotho's controller included — unlike the
// "an opponent's first spell" shape elsewhere in the catalog, this
// has no controller filter at all. g.CastTallyFor(ev.Actor).Total is
// bumped before the cast event fires, so ==2 catches exactly the
// SECOND spell and nothing after it — a third and fourth spell the
// same turn trigger nothing more, which is the printed card.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "135cc078-1c45-465e-b0b9-7168c89f56f4",
		Name:         "Lotho, Corrupt Shirriff",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				return g.CastTallyFor(ev.Actor).Total == 2
			}, "Lotho, Corrupt Shirriff — lose 1 life, create a Treasure",
				Do(GainLife{Amount: -1}, CreateToken{Template: TreasureToken(), N: 1})),
		},
	})
}
