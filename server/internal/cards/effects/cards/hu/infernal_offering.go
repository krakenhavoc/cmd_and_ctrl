package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Infernal Offering — Sorcery {4}{B} (Commander 2021, #5504):
//
//	"Choose an opponent. You and that player each sacrifice a creature.
//	 Each player who sacrificed a creature this way draws two cards.
//	 Choose an opponent. Return a creature card from your graveyard to the
//	 battlefield, then that player returns a creature card from their
//	 graveyard to the battlefield."
//
// Two separate "choose an opponent" clauses, so the second may name a
// different opponent. The sacrifices are one run of the ordered collector
// (Myrkul's Edict's): the caster and the chosen opponent are asked in
// APNAP order and everything chosen leaves together. Only a player who
// really sacrificed a creature draws two. The returns then run, the
// caster's first and the chosen player's after it has entered, each from
// their own graveyard and each mandatory
// (OldB054ReturnCreatureThenChosen).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "348201c4-79de-4278-9896-d8b1567cc5cc",
		Name:         "Infernal Offering",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 2},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.ChoosePlayer{
				Among:    effects.Opponents,
				Question: "Infernal Offering — choose an opponent; you and that player each sacrifice a creature",
				ThenKey:  infernalOfferingFirstChosen,
			}.Apply(ctx)
		},
	})
}

var infernalOfferingSacrificed = effects.RunThen("prompt-run/infernal-offering-sacrificed", func(ctx *effects.Context, r game.RunDone) error {
	for _, landed := range r.Landed {
		if len(landed.Cards) == 0 {
			continue
		}
		if err := (effects.DrawCards{Player: landed.Seat, N: 2}).Apply(ctx); err != nil {
			return err
		}
	}
	return infernalOfferingChooseAgain(ctx)
})

// infernalOfferingChooseAgain is the second "Choose an opponent".
func infernalOfferingChooseAgain(ctx *effects.Context) error {
	return effects.ChoosePlayer{
		Among:    effects.Opponents,
		Question: "Infernal Offering — choose an opponent; you each return a creature card from your graveyard to the battlefield",
		ThenKey:  infernalOfferingSecondChosen,
	}.Apply(ctx)
}

// infernalOfferingFirstChosen is the first choice's answer: you and that
// player sacrifice. An on-disk key.
var infernalOfferingFirstChosen = effects.ChoosePlayerThen("option-pick/infernal-offering-first-chosen", func(ctx *effects.Context, chosen uuid.UUID, _ []uuid.UUID) error {
	if chosen == uuid.Nil {
		return infernalOfferingChooseAgain(ctx)
	}
	return ctx.PlayersSacrificeThen([]uuid.UUID{ctx.Controller(), chosen},
		effects.SacrificeSpec("a creature", effects.Creature()),
		"Infernal Offering — sacrifice a creature", infernalOfferingSacrificed, ctx.RunData(nil))
})

// infernalOfferingSecondChosen is the second choice's answer: the returns.
// An on-disk key.
var infernalOfferingSecondChosen = effects.ChoosePlayerThen("option-pick/infernal-offering-second-chosen", func(ctx *effects.Context, chosen uuid.UUID, _ []uuid.UUID) error {
	if chosen == uuid.Nil {
		return effects.OldB054OfferCreature(ctx, ctx.Controller(), false)
	}
	return effects.OldB054ReturnCreatureThenChosen(ctx)
})
