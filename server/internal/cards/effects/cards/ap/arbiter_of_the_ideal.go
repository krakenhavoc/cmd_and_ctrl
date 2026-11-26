package cardsap

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Arbiter of the Ideal — {4}{U}{U} Creature — Sphinx 4/5 (Born of the Gods):
//
//	"Flying
//	 Inspired — Whenever this creature becomes untapped, reveal the top card of your
//	 library. If it's an artifact, creature, or land card, you may put it onto the
//	 battlefield with a manifestation counter on it. That permanent is an enchantment in
//	 addition to its other types."
//
// The reveal happens whatever the card is; the "may" is asked only for an artifact,
// creature or land card (read as it is, CR 608.2). The permanent enters with the counter
// as one entry, and "an enchantment in addition" is part of how it enters (CR 205.1b,
// layer 4), so it is an enchantment from the moment it exists. Left on top when declined.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a6cefecd-459e-439a-a62c-689c2919b689",
		Name:            "Arbiter of the Ideal",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventUntapCard, effects.Self, "Arbiter of the Ideal — reveal the top card of your library; you may put an artifact, creature or land card onto the battlefield with a manifestation counter, as an enchantment too (inspired)",
				func(g *game.Game, item *game.StackItem) error {
					revealed := g.RevealTopOfLibraryForEffect(item.Controller, item.SourceCardID, 1, "Arbiter of the Ideal — reveal the top card of your library")
					if len(revealed) == 0 {
						return nil
					}
					c, ok := g.LookupCardForEffect(revealed[0])
					if !ok || !(c.IsArtifact() || c.IsCreature() || c.IsLand()) {
						return nil
					}
					return effects.MayChoice{
						Question: "Arbiter of the Ideal — put " + c.Name + " onto the battlefield with a manifestation counter on it?",
						Then:     arbiterOfTheIdealPut,
						Carry:    []uuid.UUID{revealed[0]},
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

// arbiterOfTheIdealPut is the keyed answer: the revealed card, still on top of the
// library, enters under the chooser's control with a manifestation counter and as an
// enchantment in addition to its other types. An on-disk key.
var arbiterOfTheIdealPut = effects.MayYes("confirm/arbiter-of-the-ideal-put", func(ctx *effects.Context, r game.ConfirmAnswer) error {
	if len(r.Carry) == 0 {
		return nil
	}
	id := r.Carry[0]
	if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneLibrary {
		return nil
	}
	_, err := ctx.Game.PutFromLibraryOntoBattlefieldForEffect(id, game.LibraryEntryOptions{
		Controller:         r.Chooser,
		EntersWithCounters: map[string]int{"manifestation": 1},
		EntersAs: &game.EntersAs{
			AddTypes: []string{"Enchantment"},
			Label:    "Arbiter of the Ideal — an enchantment in addition to its other types",
		},
	})
	return err
})
