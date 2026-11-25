package cardssy

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tevesh Szat, Doom of Fools — Legendary Planeswalker — Szat {4}{B}, loyalty 4
// (Secret Lair Drop):
//
//	"+2: Create two 0/1 black Thrull creature tokens.
//	 +1: You may sacrifice another creature or planeswalker. If you do, draw
//	 two cards, then draw another card if the sacrificed permanent was a
//	 commander.
//	 −10: Gain control of all commanders. Put all commanders from the command
//	 zone onto the battlefield under your control.
//	 Tevesh Szat, Doom of Fools can be your commander.
//	 Partner"
//
// +2 is two Thrull tokens. +1 is Ruthless Technomancer's optional sacrifice
// (the "you may" and the "if you do" are the sacrifice's own answer, so a
// commander that goes to the command zone was still sacrificed, CR 701.21a),
// over another creature or planeswalker, with the extra card judged on the
// permanent as it last existed (CR 608.2h). −10 takes every commander on the
// battlefield for good (an indefinite layer 2 effect, whoever controls it; a
// copy of a commander is not one) and then puts every commander card in a
// command zone onto the battlefield as one entry under the activator's control.
// "Can be your commander" and Partner are deck-construction rules read off the
// oracle text.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "79a1783f-ea7e-4cdf-95f5-53b383d38c66",
		Name:            "Tevesh Szat, Doom of Fools",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 4,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+2: Create two 0/1 black Thrull creature tokens.",
				Cost:    effects.LoyaltyCost(2),
				Purpose: game.Purpose{Tokens: 2, CreatureTokens: 2},
				Effect:  effects.Do(effects.CreateToken{Template: effects.TokenCard("0/1 black Thrull"), N: 2}),
			},
			{
				Label:   "+1: You may sacrifice another creature or planeswalker. If you do, draw two cards, then draw another card if the sacrificed permanent was a commander.",
				Cost:    effects.LoyaltyCost(1),
				Purpose: game.Purpose{Draws: 3},
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.MaySacrificeOne(g, item, "Tevesh Szat — sacrifice another creature or planeswalker to draw two cards?",
						effects.Or(effects.Creature(), effects.Planeswalker()), true, teveshSzatSacrificed)
				},
			},
			{
				Label:  "−10: Gain control of all commanders. Put all commanders from the command zone onto the battlefield under your control.",
				Cost:   effects.LoyaltyCost(-10),
				Effect: teveshSzatTakeCommanders,
			},
		},
	})
}

// teveshSzatSacrificed is the "if you do": two cards, and one more when the
// sacrificed permanent was a commander. Keyed (on disk).
var teveshSzatSacrificed = effects.RegisterSacrificedOne("prompt-run/tevesh-szat-draw", func(ctx *effects.Context, sacrificed uuid.UUID) error {
	n := 2
	if c, ok := ctx.Game.LookupCardForEffect(sacrificed); ok && c.IsCommander {
		n = 3
	}
	return effects.DrawCards{N: n}.Apply(ctx)
})

// teveshSzatTakeCommanders is −10.
func teveshSzatTakeCommanders(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	var onField []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsCommander && c.Controller != ctx.Controller() {
			onField = append(onField, c.InstanceID)
		}
	}
	for _, id := range onField {
		g.GainControlForEffect(item.SourceCardID, id, ctx.Controller(), game.IndefiniteDuration(), "Tevesh Szat — gain control of all commanders")
	}
	var fromCommandZone []uuid.UUID
	for _, p := range g.Seats {
		if p == nil || p.Command == nil {
			continue
		}
		for _, c := range p.Command.Cards {
			if c.IsCommander && c.IsPermanent() {
				fromCommandZone = append(fromCommandZone, c.InstanceID)
			}
		}
	}
	if len(fromCommandZone) == 0 {
		return nil
	}
	return effects.PutCardsFromZoneTogetherWith(ctx, game.ZoneCommand, fromCommandZone,
		game.ZoneEntryOptions{Controller: ctx.Controller()}, game.TailRef{}, 0)
}
