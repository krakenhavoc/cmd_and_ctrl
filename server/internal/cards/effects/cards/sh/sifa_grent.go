package cardssh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sifa Grent — Legendary Planeswalker — Sifa {3}{B}{R}, loyalty 4 (Mystery
// Booster Commander, #5948):
//
//	"Whenever a creature dies, target opponent loses 1 life and you gain 1 life.
//	 +1: Goad up to two target creatures.
//	 −6: Each opponent exiles the top two cards of their library. Until end of
//	 turn, you may play those cards without paying their mana costs.
//	 Sifa Grent can be your commander."
//
// The static trigger is a plain dies watch over every creature, whoever's, with
// the opponent chosen as it goes on the stack (CR 603.3d); it is removed with no
// opponent left. The +1 is Galvanizing Tenor's goad over up to two targets. The
// −6 is Fallen Shinobi's exile with a free permission, once for each opponent,
// each library off its own owner, and the window is this turn only
// (ExileTopWithPermission.Window). "Can be your commander" is deck-building,
// read from the oracle text by the deck validator.
//
// No simplification.
func init() {
	dies := effects.Targeting(
		effects.On(game.EventLTB, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
			_, ok := effects.DiedCreature(ev, g)
			return ok
		}, "Sifa Grent — target opponent loses 1 life and you gain 1 life", sifaGrentDrain),
		effects.TargetPlayer("target opponent", effects.Opponent()))
	effects.Register(effects.Spec{
		OracleID:     "9a3d148c-e1bb-4e7a-a588-00a24041ebf7",
		Name:         "Sifa Grent",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{dies},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+1: Goad up to two target creatures.",
				Cost:    effects.LoyaltyCost(1),
				Targets: effects.TargetCreature("up to two target creatures").WithCount(0, 2),
				Effect:  effects.GalvanizingTenorGoad,
			},
			{
				Label:  "−6: Each opponent exiles the top two cards of their library. Until end of turn, you may play those cards without paying their mana costs.",
				Cost:   effects.LoyaltyCost(-6),
				Effect: sifaGrentExile,
			},
		},
	})
}

func sifaGrentDrain(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	t, ok := effects.FirstLegalTarget(ctx)
	if !ok || t.Kind != game.TargetPlayer {
		return nil
	}
	if err := g.ChangePlayerLifeForEffect(item.SourceCardID, t.ID, -1); err != nil {
		return err
	}
	return effects.GainLife{Player: item.Controller, Amount: 1}.Apply(ctx)
}

func sifaGrentExile(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	window := g.UntilEndOfTurnDuration()
	var err error
	effects.EachOpponent(g, item.Controller, func(opp uuid.UUID) bool {
		err = effects.ExileTopWithPermission{From: opp, GrantTo: item.Controller, N: 2, Free: true, Window: &window}.Apply(ctx)
		return err != nil
	})
	return err
}
