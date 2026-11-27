package cardslu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lumengrid Augur — Creature — Vedalken Wizard {3}{U}, 2/2 (Mirrodin, S82
// OLD, #5606):
//
//	"{1}, {T}: Target player draws a card, then discards a card. If that
//	 player discards an artifact card this way, untap this creature."
//
// Civilized Scholar's loot with its untap rider: the target (any player)
// draws, then chooses the discard from the hand that made, and the
// discard's keyed answer reads what really left the hand. An artifact card
// among them untaps the Augur if it is still on the battlefield; an empty
// hand discards nothing and untaps nothing. A target that is no longer
// legal does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d5409a7c-b928-4819-a560-e80226cf61f0",
		Name:         "Lumengrid Augur",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}: Target player draws a card, then discards a card. If that player discards an artifact card this way, untap this creature.",
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
			Targets: effects.TargetPlayer("target player"),
			Purpose: effects.ForTargets(game.TargetPurpose{Slot: 0, Draws: 1, Discards: 1}),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetPlayer {
						continue
					}
					if ctx.IsNewSourceObject(item.SourceCardID) || !effects.OnBattlefield(g, item.SourceCardID) {
						return effects.DrawThenDiscard(g, t.ID, item.SourceCardID, 1, 1, "")
					}
					return g.DrawNThenForEffect(t.ID, 1, game.DrawThen{
						Ref: lumengridAugurLoot, Player: t.ID, Source: item.SourceCardID, N: 1,
					})
				}
				return nil
			},
		}},
	})
}

// lumengridAugurLoot is the discard half once the draw has settled: the
// discard carries the Augur so its answer can untap it. An on-disk key.
var lumengridAugurLoot = game.RegisterDrawThen("s82-old-b602-lumengrid-augur-loot", func(g *game.Game, d game.DrawThen) error {
	return g.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
		Player: d.Player,
		Source: d.Source,
		N:      d.N,
	}, lumengridAugurDiscarded, []uuid.UUID{d.Source})
})

// lumengridAugurDiscarded is "If that player discards an artifact card this
// way, untap this creature." An on-disk key: never renamed, never reused.
var lumengridAugurDiscarded = game.RegisterDiscardThen("choose-cards/s82-old-b602-lumengrid-augur-discarded", func(g *game.Game, r game.DiscardsDone) error {
	if len(r.Carry) == 0 || !effects.OnBattlefield(g, r.Carry[0]) {
		return nil
	}
	for _, id := range r.Discarded.Cards() {
		if c, ok := g.LookupCardForEffect(id); ok && c.HasCardType("artifact") {
			return g.UntapTargetForEffect(r.Carry[0])
		}
	}
	return nil
})
