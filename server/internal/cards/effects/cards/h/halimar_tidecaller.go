package cardsh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Halimar Tidecaller — Creature — Human Wizard Ally {2}{U}, 2/3 (Battle for
// Zendikar, #5547):
//
//	"When this creature enters, you may return target card with awaken from
//	 your graveyard to your hand.
//	 Land creatures you control have flying."
//
// "With awaken" is the card's printed awaken cost (the catalog's awaken
// alternative cost, as Rootcoil Creeper reads flashback); a card that only
// gained awaken, or one the catalog does not know, is not one. The land
// creatures are read through the layered view, so an awakened land counts.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3aa709b5-0465-4103-aea1-8cfb15086f76",
		Name:         "Halimar Tidecaller",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.KeywordGrant(effects.B31LandCreaturesYouControl, "flying")},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(effects.WhenThisEnters("Halimar Tidecaller — you may return target card with awaken from your graveyard to your hand", effects.ReturnTargetedCardToHand),
				effects.TargetCardInGraveyard("target card with awaken from your graveyard", effects.YouOwn(), S82OldB313HasPrintedAwaken)),
				"Halimar Tidecaller — return the card with awaken to your hand?"),
		},
	})
}

// S82OldB313HasPrintedAwaken is the predicate for "card with awaken": the
// card prints an awaken cost.
func S82OldB313HasPrintedAwaken(_ *game.Game, _ uuid.UUID, c game.Card) bool {
	return game.AlternativeCostByKey(game.CatalogKey(c), effects.AwakenKey) != nil
}
