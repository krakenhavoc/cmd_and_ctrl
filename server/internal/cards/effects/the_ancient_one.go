package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Ancient One — Legendary Creature — Spirit God {U}{B}, 8/8 (The Lost
// Caverns of Ixalan, #4470):
//
//	"Descend 8 — The Ancient One can't attack or block unless there are
//	 eight or more permanent cards in your graveyard.
//	 {2}{U}{B}: Draw a card, then discard a card. When you discard a card
//	 this way, target player mills cards equal to its mana value."
//
// The static is a layer-6 restriction read live on every recompute
// (CR 611.3a). The ability is Cait Sith's reflexive shape: the draw, then
// the controller's own discard (a keyed continuation, so a table waiting on
// the pick is a restore point), and only if a card was really discarded a
// reflexive trigger (CR 603.12) that targets a player as it goes on the
// stack, the mana value of the discarded card (read in the graveyard, a
// split card's halves combined) carried as its amount. An empty hand
// discards nothing and the trigger never exists.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6615f881-960c-4216-a32a-63856bc3afd0",
		Name:         "The Ancient One",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label: "Descend 8 — The Ancient One can't attack or block unless there are eight or more permanent cards in your graveyard.",
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return selfOnly(target, g, source) && permanentCardsInGraveyard(g, source.Controller) < 8
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Restrictions |= game.CantAttack | game.CantBlock
			},
		}},
		Activated: []ActivatedAbility{{
			Label:   "{2}{U}{B}: Draw a card, then discard a card. When you discard a card this way, target player mills cards equal to its mana value.",
			Cost:    ManaCost("{2}{U}{B}"),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return g.DrawNThenForEffect(item.Controller, 1, game.DrawThen{
					Ref: ancientOneDrawThen, Player: item.Controller, Source: item.SourceCardID, N: 1,
				})
			},
		}},
	})
}

// ancientOneDrawThen is the discard after the draw. An on-disk key: never
// renamed, never reused.
var ancientOneDrawThen = game.RegisterDrawThen("the-ancient-one/discard", func(g *game.Game, d game.DrawThen) error {
	return g.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
		Player: d.Player, Source: d.Source, N: d.N,
	}, ancientOneDiscarded, nil)
})

// ancientOneDiscarded is "when you discard a card this way": the reflexive
// mill, with the discarded card's mana value as its amount.
var ancientOneDiscarded = game.RegisterDiscardThen("choose-cards/the-ancient-one-discard", func(g *game.Game, r game.DiscardsDone) error {
	cards := r.Discarded.Cards()
	if len(cards) == 0 {
		return nil
	}
	mv := 0
	if c, ok := g.LookupCardForEffect(cards[0]); ok {
		mv, _ = g.ManaValueForEffect(c)
	}
	return ReflexiveTrigger{
		Label:  "The Ancient One — target player mills cards equal to the discarded card's mana value",
		Cards:  []uuid.UUID{cards[0]},
		Body:   ancientOneMillBody,
		Params: game.EffectParams{Amount: mv},
	}.Apply(keyedContext(g, r.Player, r.Source))
})

// ancientOneMillBody is the reflexive half: the chosen player mills
// Params.Amount cards.
var ancientOneMillBody = game.ReflexiveBody("the-ancient-one/mill", func(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetPlayer {
			return MillCards{Player: t.ID, N: p.Amount}.Apply(ctx)
		}
	}
	return nil
}, constTargets(func() *game.TargetSpec { return TargetPlayer("target player") }))
