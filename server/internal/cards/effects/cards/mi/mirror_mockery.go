package cardsmi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mirror Mockery — Enchantment — Aura {1}{U} (Dragons of Tarkir, S82 old batch
// 332):
//
//	"Enchant creature
//	 Whenever enchanted creature attacks, you may create a token that's a copy of
//	 that creature. Exile that token at end of combat."
//
// The trigger belongs to the Aura (WheneverEnchantedCreatureAttacks), so "you" is the
// Aura's controller, who owns the token however the creature is controlled. The
// "may" is asked as it resolves. The token is a copy of the attacking creature
// (Invocation-style copy through CreateTokenCopiesTogether, CR 707.2), not attacking,
// and the tail reads the token it made so a delayed trigger at the end of combat
// exiles exactly that token. The copy's own enters triggers fire. A creature that has
// left the battlefield is copied as it last existed (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6a802318-9123-4058-8554-5023eace33f4",
		Name:         "Mirror Mockery",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit, Tokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.WheneverEnchantedCreatureAttacks("Mirror Mockery — you may create a token that's a copy of that creature; exile it at end of combat",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayChoice{
						Question: "Mirror Mockery — create a token that's a copy of that creature?",
						Then:     mirrorMockeryCopy,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

// mirrorMockeryCopy is the yes branch: the token copy, whose tail schedules the
// exile. An on-disk key: never renamed, never reused.
var mirrorMockeryCopy = effects.MayYesNo("confirm/mirror-mockery-copy",
	func(ctx *effects.Context, _ game.ConfirmAnswer) error {
		return effects.CreateTokenCopiesTogether{
			Controller: ctx.Controller(),
			Copies:     []uuid.UUID{ctx.Trigger().Event.CardID},
			Then:       ctx.ItemTail(mirrorMockeryExile),
		}.Apply(ctx)
	}, nil)

// mirrorMockeryExile exiles the token at end of combat. An on-disk key.
var mirrorMockeryExile = effects.CardTail("old-b332/mirror-mockery-exile", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	return effects.ScheduleDelayedTrigger{At: game.StepEndCombat, Controller: ctx.Controller(),
		Label: "Mirror Mockery — exile the token", Cards: o.IDs, Body: effects.ExileListedCardsBody}.Apply(ctx)
})
