package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mirrored Depths — Plane — Karsus (Planechase Anthology, #4870):
//
//	"Whenever a player casts a spell, that player flips a coin. If the
//	 player loses the flip, counter that spell.
//	 Whenever chaos ensues, target player reveals the top card of their
//	 library. If it's a nonland card, you may cast it without paying its
//	 mana cost."
//
// A plane (ADR 0216 §6). The first ability is the caster's own called flip
// (CR 705.2): the trigger resolves above the spell, a lost flip counters it,
// a won one leaves it. The chaos ability targets a player as it goes on the
// stack, reveals the top card of that player's library, and the planar
// controller is offered a free cast of it during the resolution (CR 608.2g;
// Commander Spock's offer): a land is not offered, and a card not cast stays
// on top of its owner's library.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b6329c2f-ea7f-4611-986c-f82bacf2a872",
		Name:         "Mirrored Depths",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, planesR6APlayerCastsASpell,
				"Mirrored Depths — that player flips a coin; if they lose the flip, counter that spell",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					ev := ctx.Trigger().Event
					if ev.CardID == uuid.Nil || g.StackItemForEffect(ev.CardID) == nil {
						return nil
					}
					g.FlipCoinForEffect(game.CoinFlipSpec{
						Flipper:  ev.Actor,
						Source:   ctx.Source(),
						Question: "Mirrored Depths — call the coin flip (lose it and your spell is countered)",
						ThenKey:  mirroredDepthsFlipped,
						Carry:    []uuid.UUID{ev.CardID},
						Asked:    game.NewAskedItem(item),
					})
					return nil
				}),
			Targeting(WheneverChaosEnsues("Mirrored Depths — target player reveals the top card of their library; if it's a nonland card, you may cast it without paying its mana cost",
				mirroredDepthsReveal), TargetPlayer("target player")),
		},
	})
}

// mirroredDepthsFlipped counters the spell on a lost flip. An on-disk key:
// never renamed, never reused.
var mirroredDepthsFlipped = CoinFlipThen("coin-flip/mirrored-depths-counter", func(ctx *Context, r game.CoinFlipped) error {
	if len(r.Carry) == 0 || wonFirstFlip(r) {
		return nil
	}
	return CounterTarget{StackID: r.Carry[0]}.Apply(ctx)
})

const mirroredDepthsLabel = "Mirrored Depths"

func mirroredDepthsReveal(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var victim uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetPlayer {
			victim = t.ID
			break
		}
	}
	if victim == uuid.Nil {
		return nil
	}
	you, source := ctx.Controller(), ctx.Source()
	revealed := g.RevealTopOfLibraryForEffect(victim, source, 1, "Mirrored Depths — revealed from the top of the library")
	if len(revealed) == 0 {
		return nil
	}
	if c, ok := g.LookupCardForEffect(revealed[0]); !ok || c.IsLand() {
		return nil
	}
	perm := anySpell.on(g.FreeCastFromLibraryPermission(you, source, mirroredDepthsLabel))
	perm.LapseOnPass = game.LapseStaysInHand
	if len(freeCastCandidates(g, &perm, revealed)) == 0 {
		return nil
	}
	g.OfferResolutionCastForEffect(game.ResolutionCastOffer{
		Chooser:      you,
		Source:       source,
		Reason:       mirroredDepthsLabel,
		Cards:        revealed,
		Permission:   &perm,
		Keyword:      game.MayCastKeywordFreeCast,
		AcceptLabel:  "Cast it free",
		DeclineLabel: "Don't cast it",
	})
	return nil
}
