package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Portcullis — Artifact {4} (Stronghold, S82 OLD b751, #5627):
//
//	"Whenever a creature enters, if there are two or more other creatures on the
//	 battlefield, exile that creature. Return that card to the battlefield under
//	 its owner's control when this artifact leaves the battlefield."
//
// Fiend Hunter's pair of linked abilities (CR 607.2a): the enter trigger exiles
// the creature that entered, and the leave trigger returns every card this
// Portcullis exiled and that is still in exile, together, under its owner's
// control. The "if" is intervening (CR 603.4): it is true when the creature
// enters and again as the trigger resolves, counting creatures other than the
// one that entered, whoever controls them. A token exiled this way ceases to
// exist and does not return. If the Portcullis leaves before the trigger
// resolves, nothing returns and the creature stays exiled.
//
// No simplification.
func init() {
	const exileLabel = "Portcullis — exile that creature"
	enters := effects.On(game.EventETB, func(ev game.Event, source *game.Card, c game.Characteristic, g *game.Game) bool {
		return effects.AnotherCreatureEntered(ev, source, c, g) && effects.OldB751TwoOrMoreOtherCreatures(g, ev.CardID)
	}, exileLabel, func(g *game.Game, item *game.StackItem) error {
		if item.Trigger == nil {
			return nil
		}
		id := item.Trigger.Event.CardID
		if !effects.OnBattlefield(g, id) || !effects.OldB751TwoOrMoreOtherCreatures(g, id) {
			return nil
		}
		return effects.ExileTarget{Target: id}.Apply(effects.NewContext(g, item))
	})
	effects.Register(effects.Spec{
		OracleID:     "6230c179-0263-4fa4-b0ff-f2cb621de147",
		Name:         "Portcullis",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			enters,
			effects.On(game.EventLTB, effects.Self, "Portcullis — return the exiled cards to the battlefield under their owners' control",
				effects.B41ReturnCardsExiledWithToTheBattlefield(exileLabel)),
		},
	})
}
