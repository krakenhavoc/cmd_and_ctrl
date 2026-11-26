package cardsex

import (
	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Faerie Impostor — Creature — Faerie Rogue {U}, 2/1 (Duel Decks: Speed vs.
// Cunning, #5555):
//
//	"Flying
//	 When this creature enters, sacrifice it unless you return another
//	 creature you control to its owner's hand."
//
// Quickling's trigger, asked as it resolves: with no other creature the
// Impostor is sacrificed at once, and naming none, naming one that has left
// or a return that does not reach a hand sacrifices it too.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0d78ba2c-4de0-47d9-b1b9-13dc9de8844b",
		Name:            "Faerie Impostor",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Faerie Impostor — sacrifice it unless you return another creature you control to its owner's hand", faerieImpostorEnters),
		},
	})
}

func faerieImpostorEnters(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	var others []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == ctx.Controller() && c.IsCreature() && c.InstanceID != item.SourceCardID {
			others = append(others, c.InstanceID)
		}
	}
	return effects.SacrificeUnlessYouReturnOne(ctx, item, others,
		"Faerie Impostor — return another creature you control to its owner's hand (or sacrifice Faerie Impostor)",
		faerieImpostorPicked)
}

// faerieImpostorPicked returns the chosen creature, or sacrifices the
// Impostor. An on-disk key: never renamed, never reused.
var faerieImpostorPicked = effects.AskedPick("choose-cards/faerie-impostor-return", func(ctx *effects.Context, r game.CardsChosen) error {
	return effects.ReturnPickedOrSacrifice(ctx, r, faerieImpostorBounced)
})

// faerieImpostorBounced sacrifices the Impostor when the creature did not
// reach a hand. An on-disk key: never renamed, never reused.
var faerieImpostorBounced = effects.CardTail("faerie-impostor/after-bounce", effects.SacrificeSourceUnlessBounced)
