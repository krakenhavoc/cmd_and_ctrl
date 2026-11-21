package cardsn

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nahiri, the Unforgiving — Legendary Planeswalker — Nahiri {1}{R}{R/W/P}{W},
// starting loyalty 5 (Phyrexia: All Will Be One, #5479):
//
//	"Compleated ({R/W/P} can be paid with {R}, {W}, or 2 life. If life was
//	 paid, this planeswalker enters with two fewer loyalty counters.)
//	 +1: Until your next turn, up to one target creature attacks a player
//	     each combat if able.
//	 +1: Discard a card, then draw a card.
//	 0: Exile target creature or Equipment card with mana value less than
//	    Nahiri's loyalty from your graveyard. Create a token that's a copy
//	    of it. That token gains haste. Exile it at the beginning of the
//	    next end step."
//
// Compleated is ADR 0183's replacement (effects.Compleated); the hybrid
// Phyrexian symbol is one symbol that any of {R}, {W} or 2 life pays
// (CR 107.4f). The first +1 is an attack requirement obeyed only by an
// attack on a player (game.AttackAPlayerIfAbleMod: an attack on a
// planeswalker or battle does not obey it, CR 508.1d), pinned to the target
// until its Nahiri's controller's next turn. The loot discards first, then
// draws even from an empty hand. The 0's target is judged against Nahiri's
// loyalty both when it is chosen and on resolution (CR 608.2b); the card is
// exiled, then a token copying it (its copiable values, CR 707.2) is made
// with haste and exiled at the beginning of the next end step (CR 603.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d71d42e0-cd56-4f53-95ff-ecab7f8c6552",
		Name:            "Nahiri, the Unforgiving",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 5,
		PrintedKeywords: []string{"compleated"},
		Replacements:    []game.ReplacementEffect{effects.Compleated()},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+1: Until your next turn, up to one target creature attacks a player each combat if able.",
				Cost:    effects.LoyaltyCost(1),
				Targets: effects.TargetCreature("up to one target creature").WithCount(0, 1),
				Effect:  nahiriUnforgivingPlusOne,
			},
			{
				Label:   "+1: Discard a card, then draw a card.",
				Cost:    effects.LoyaltyCost(1),
				Purpose: game.Purpose{Draws: 1, Discards: 1},
				Effect: func(g *game.Game, item *game.StackItem) error {
					return g.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
						Player: item.Controller,
						Source: item.SourceCardID,
						N:      1,
					}, effects.TheRuinousWreckingCrewDraw, nil)
				},
			},
			{
				Label: "0: Exile target creature or Equipment card with mana value less than Nahiri's loyalty from your graveyard. Create a token that's a copy of it. That token gains haste. Exile it at the beginning of the next end step.",
				Cost:  effects.LoyaltyCost(0),
				Targets: effects.RelativeToSource(
					effects.TargetCardInGraveyard("target creature or Equipment card with mana value less than Nahiri's loyalty from your graveyard",
						effects.YouOwn(), effects.Or(effects.Creature(), effects.Subtype("Equipment"))),
					effects.S82OneB20ManaValueBelowLoyalty()),
				Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1},
				Effect:  nahiriUnforgivingZero,
			},
		},
	})
}

func nahiriUnforgivingPlusOne(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	targets := ctx.LegalTargets()
	if len(targets) == 0 || targets[0].Kind != game.TargetCard {
		return nil
	}
	return effects.ScopedEffectFor{
		Target:   targets[0].ID,
		Mods:     []game.Mod{game.AttackAPlayerIfAbleMod()},
		Duration: effects.DurationUntilYourNextTurn(ctx, ctx.Controller()),
		Label:    "Nahiri, the Unforgiving — attacks a player each combat if able until your next turn",
	}.Apply(ctx)
}

func nahiriUnforgivingZero(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	targets := ctx.LegalTargets()
	if len(targets) == 0 || targets[0].Kind != game.TargetCard {
		return nil
	}
	return effects.ExileTarget{
		Target: targets[0].ID,
		ThenTail: game.KeyedTail{
			Ref: nahiriUnforgivingCopy, Player: item.Controller, Source: item.SourceCardID,
			Carry: []uuid.UUID{targets[0].ID},
		},
	}.Apply(ctx)
}

// nahiriUnforgivingCopy is the 0's second sentence onward: the token copy of
// the exiled card, with haste, exiled at the next end step. It runs whether
// or not the exile reached exile, as the printed sentences are independent.
// An on-disk key: never renamed, never reused.
var nahiriUnforgivingCopy = effects.CardTail("then/nahiri-the-unforgiving-copy", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.Tail.Carry) == 0 {
		return nil
	}
	return effects.TrekB40CreateCopyExiledAtNextEndStep(ctx, o.Tail.Player, o.Tail.Carry[0],
		effects.TokenCopyGainsHaste, "Nahiri, the Unforgiving — exile the token")
})
