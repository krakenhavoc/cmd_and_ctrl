package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Transplant Theorist — Artifact Creature — Phyrexian Artificer {3}{U},
// 2/4 (Phyrexia: All Will Be One):
//
//	"Whenever this creature or another artifact you control enters, you may
//	 draw a card. If you do, discard a card.
//	 {2}: Put target card from your graveyard on the bottom of your
//	 library."
//
// The loot is Rook Turret's linked "you may draw, if you do discard"; it
// triggers on any artifact entering under your control, the Theorist
// included. The activated ability tucks one card of your graveyard.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4a60db66-0347-4efc-8507-ab7ed8d984fe",
		Name:         "Transplant Theorist",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := effects.EnteredUnderYourControl(ev, source, g, false)
				return ok && c.IsArtifact()
			}, "Transplant Theorist — you may draw a card; if you do, discard a card",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayChoice{
						Question: "Transplant Theorist — draw a card, then discard a card?",
						Draws:    1,
						YesLabel: "Draw",
						NoLabel:  "Decline",
						Then:     effects.CelestusAnswer,
					}.Apply(effects.NewContext(g, item))
				}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}: Put target card from your graveyard on the bottom of your library.",
			Cost:    effects.ManaCost("{2}"),
			Targets: effects.TargetCardInGraveyard("target card from your graveyard", effects.YouOwn()),
			Effect:  effects.S82One13TuckTargetToLibraryBottom,
		}},
	})
}
