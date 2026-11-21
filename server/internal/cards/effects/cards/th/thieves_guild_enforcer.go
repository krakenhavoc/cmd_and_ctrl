package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thieves' Guild Enforcer — Creature — Human Rogue {B}, 1/1 (Foundations Jumpstart):
//
//	"Flash
//	 Whenever this creature or another Rogue you control enters, each
//	 opponent mills two cards.
//	 As long as an opponent has eight or more cards in their graveyard, this
//	 creature gets +2/+1 and has deathtouch."
//
// One enters trigger for this creature and any other Rogue under your
// control (read through the layered view); the mill is Altar of the Brood's.
// The static is Blackbloom Rogue's per-opponent graveyard count, read live.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "84c0db59-255d-41af-a3d4-3f3548f1195f",
		Name:            "Thieves' Guild Enforcer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := effects.EnteredUnderYourControl(ev, source, g, false)
				return ok && (c.InstanceID == source.InstanceID || c.HasSubtype(source.TextSubtype("Rogue")))
			}, "Thieves' Guild Enforcer — each opponent mills two cards", func(g *game.Game, item *game.StackItem) error {
				return effects.B12EachOpponentMills(g, item, 2)
			}),
		},
		Static: thievesGuildEnforcerStatics(),
	})
}

func thievesGuildEnforcerStatics() []game.StaticAbility {
	gate := func(target *game.Card, g *game.Game, source *game.Card) bool {
		return effects.SelfOnly(target, g, source) && effects.AnOpponentHasGraveyardOfAtLeast(g, source.Controller, 8)
	}
	pump := effects.B16Anthem(gate, 2, 1)
	pump.Label = "As long as an opponent has eight or more cards in their graveyard, this creature gets +2/+1."
	dt := effects.KeywordGrant(gate, "deathtouch")
	dt.Label = "As long as an opponent has eight or more cards in their graveyard, this creature has deathtouch."
	return []game.StaticAbility{pump, dt}
}
