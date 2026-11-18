package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// All-Seeing Arbiter — Creature — Avatar {4}{U}{U}, 5/4 (Streets of New
// Capenna):
//
//	"Flying
//	 Whenever this creature enters or attacks, draw two cards, then discard a
//	 card.
//	 Whenever you discard a card, target creature an opponent controls gets
//	 -X/-0 until your next turn, where X is the number of different mana
//	 values among cards in your graveyard."
//
// The first ability is one trigger watching two events (Sun Titan's shape); its
// discard is the player's own choice from a hand that includes the draws.
//
// The second fires once per card discarded, whoever made you discard it
// (including the first ability's own discard), and picks its target as it goes
// on the stack. X is counted as it resolves, so the card just discarded is
// already in the graveyard and counts (CR 608.2h: the count is made on
// resolution), and the -X/-0 lasts until the start of its controller's next
// turn (CR 611.2). An X of 0 gives -0/-0.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f15460d9-0133-436c-b0fe-1661989124f0",
		Name:            "All-Seeing Arbiter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersOrAttacks("All-Seeing Arbiter — draw two cards, then discard a card",
				func(g *game.Game, item *game.StackItem) error { return b16DrawThenDiscard(g, item, 2) }),
			Targeting(On(game.EventDiscardCard, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return discardedByYou(ev, source)
			}, "All-Seeing Arbiter — target creature an opponent controls gets -X/-0 until your next turn",
				allSeeingArbiterShrink),
				TargetCreature("target creature an opponent controls", OpponentControls())),
		},
	})
}

// allSeeingArbiterShrink is the discard trigger's body: -X/-0, X the number
// of different mana values among cards in the controller's graveyard now.
func allSeeingArbiterShrink(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	x := sncGraveyardMVs(g, item.Controller, uuid.Nil)
	if x == 0 {
		return nil
	}
	return ScopedEffectFor{
		Target:   id,
		Mods:     []game.Mod{game.ModifyPTMod(-x, 0)},
		Duration: DurationUntilYourNextTurn(ctx, ctx.Controller()),
		Label:    "All-Seeing Arbiter — -X/-0 until your next turn",
	}.Apply(ctx)
}
