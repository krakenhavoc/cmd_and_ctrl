package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Deathless Knight — Creature — Skeleton Knight {B/G}{B/G}{B/G}{B/G}, 4/2
// (Throne of Eldraine, S82 old sets):
//
//	"Haste
//	 When you gain life for the first time each turn, return this card from
//	 your graveyard to your hand."
//
// A graveyard ability (CR 113.6, InGraveyard): the harvest hands the predicate
// a source controlled by its owner. The first-gain reader is Vanguard
// Seraph's, read off the turn's log, so a gain made while the Knight was on the
// battlefield or in the graveyard before it died still uses up the turn's
// first. The return does nothing if the card has left the graveyard.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f96f1a94-6226-4e1f-a7fb-a43a276b3c3a",
		Name:            "Deathless Knight",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			effects.InGraveyard(effects.On(game.EventChangeLife, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.SocB28FirstLifeGainThisTurn(ev, source, g)
			}, "Deathless Knight — return this card from your graveyard to your hand", effects.ReturnThisCardFromYourGraveyardToYourHand)),
		},
	})
}
