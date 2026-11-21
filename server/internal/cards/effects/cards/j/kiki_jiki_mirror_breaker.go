package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kiki-Jiki, Mirror Breaker — Legendary Creature — Goblin Shaman
// {2}{R}{R}{R}, 2/2:
//
//	"Haste
//	 {T}: Create a token that's a copy of target nonlegendary creature you
//	 control, except it has haste. Sacrifice it at the beginning of the
//	 next end step."
//
// Reflection of Kiki-Jiki's ability without the {1} and without
// "another" (Kiki-Jiki is legendary, so it can never be its own target).
// Its own haste is printed, so the {T} can be paid the turn it arrives.
//
// "Except it has haste" is part of the copy (TokenCopyGainsHaste), so a
// later copy of the token has haste too. The copy runs the ordinary
// token entry, so the copied creature's enters triggers fire for it.
//
// The sacrifice is a delayed trigger over the token that landed,
// scheduled from a keyed tail: the tail runs even when the token's entry
// stopped to ask a question (an "as this enters, choose" clause), so the
// token is never left without its sacrifice. "Sacrifice IT" is the token,
// never Kiki-Jiki. The target is re-read at resolution, so a creature
// that left in response makes nothing.
//
// No simplification.
func init() {
	const text = "{T}: Create a token that's a copy of target nonlegendary creature you control, except it has haste. Sacrifice it at the beginning of the next end step."
	effects.Register(effects.Spec{
		OracleID:        "a34b7416-cfe3-4a1e-a8c1-a3056b747519",
		Name:            "Kiki-Jiki, Mirror Breaker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Activated: []effects.ActivatedAbility{{
			Label:   text,
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target nonlegendary creature you control", effects.YouControl(), effects.Not(effects.Legendary())),
			Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1},
			Effect:  kikiJikiMirrorBreakerCopy,
		}},
	})
}

// kikiJikiMirrorBreakerCopy makes the hasty copy; its tail schedules the
// sacrifice.
func kikiJikiMirrorBreakerCopy(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	id, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return effects.CreateTokenCopiesTogether{
		Controller: item.Controller,
		Copies:     []uuid.UUID{id},
		Except:     effects.TokenCopyGainsHaste,
		Then:       ctx.ItemTail(kikiJikiMirrorBreakerSacrifice),
	}.Apply(ctx)
}

// kikiJikiMirrorBreakerSacrifice is "Sacrifice it at the beginning of the
// next end step." over the token that landed (CR 603.7). An on-disk key:
// never renamed, never reused.
var kikiJikiMirrorBreakerSacrifice = effects.CardTail("kiki-jiki-mirror-breaker/sacrifice-the-token", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	return effects.ScheduleDelayedTrigger{
		At:         game.StepEnd,
		Controller: ctx.Item.Controller,
		Label:      "Kiki-Jiki, Mirror Breaker — sacrifice the token",
		Cards:      o.IDs,
		Body:       effects.SacrificeListedCardsBody,
	}.Apply(ctx)
})
