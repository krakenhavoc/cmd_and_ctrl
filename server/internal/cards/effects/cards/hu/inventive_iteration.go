package cardshu

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Inventive Iteration // Living Breakthrough — an Enchantment — Saga {3}{U}
// (Kamigawa: Neon Dynasty, S82):
//
//	I   — Return up to one target creature or planeswalker to its owner's
//	      hand.
//	II  — Return an artifact card from your graveyard to your hand. If you
//	      can't, draw a card.
//	III — Exile this Saga, then return it to the battlefield
//	      transformed under your control.
//
// Back face (registered under "<oracle_id>#1"), Living Breakthrough, an
// Enchantment Creature — Moonfolk 3/3:
//
//	Flying
//	Whenever you cast a spell, your opponents can't cast spells with the
//	same mana value as that spell until your next turn.
//
// II chooses the artifact card as it resolves (it is not targeted): with no
// artifact card in the graveyard there is nothing to return, so the player
// draws instead (CR 608.2c). The back face's ban is a granted cast
// restriction of its own kind, game.CastBanManaValue, on each opponent, for
// a duration ending at the start of the Breakthrough's controller's next
// turn (the Reflector Mage seam). The value is the cast spell's mana value as
// the trigger resolves, X included while the spell is still on the stack
// (CR 202.3e).
// Chapter I is otawaraChannel's "return the chosen permanent to its owner's
// hand" over an up-to-one target.
//
// No simplification.
func init() {
	const name = "Inventive Iteration"
	effects.Register(effects.Spec{
		OracleID:     InventiveIterationOracleID,
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.ChapterTriggerTargeting(1, effects.SagaChapterLabel(name, 1, "return up to one target creature or planeswalker to its owner's hand"),
				effects.TargetPermanent("up to one target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())).WithCount(0, 1),
				effects.OtawaraChannel),
			effects.ChapterTrigger(2, effects.SagaChapterLabel(name, 2, "return an artifact card from your graveyard to your hand; if you can't, draw a card"),
				inventiveIterationReturnOrDraw),
			effects.ChapterTrigger(3, effects.SagaChapterLabel(name, 3, "exile this Saga, then return it transformed"),
				effects.ChapterExileAndReturnTransformed),
		},
	})
	effects.Register(effects.Spec{
		OracleID:        InventiveIterationOracleID + "#1",
		Name:            "Living Breakthrough",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(nil, "Living Breakthrough — your opponents can't cast spells with the same mana value as that spell until your next turn",
				livingBreakthroughBan),
		},
	})
}

// inventiveIterationOracleID is shared with the back face.
const InventiveIterationOracleID = "e711a213-2a6b-4b34-a788-95b6c0b7ba48"

// inventiveIterationReturnOrDraw is chapter II.
func inventiveIterationReturnOrDraw(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if len(effects.GraveyardCardIDs(ctx, item.Controller, func(c game.Card) bool { return c.IsArtifact() })) == 0 {
		return effects.DrawCards{N: 1}.Apply(ctx)
	}
	return effects.ReturnChosenFromGraveyard{
		Question: "Inventive Iteration — return an artifact card from your graveyard to your hand",
		Match:    func(c game.Card) bool { return c.IsArtifact() },
		Min:      1,
		Max:      1,
		ToHand:   true,
	}.Apply(ctx)
}

// livingBreakthroughBan bans each opponent from casting spells with the
// mana value of the spell just cast until the controller's next turn.
func livingBreakthroughBan(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	spell, ok := g.LookupCardForEffect(item.Trigger.Event.CardID)
	if !ok {
		return nil
	}
	mv, ok := g.ManaValueForEffect(spell)
	if !ok {
		return nil
	}
	for _, opp := range ctx.Opponents() {
		if err := (effects.RestrictCasting{
			Player: opp,
			Rule: game.CastBanRule{
				Kind:      game.CastBanManaValue,
				ManaValue: mv,
			},
			Label:    "Living Breakthrough — can't cast spells with mana value " + strconv.Itoa(mv),
			Duration: effects.DurationUntilYourNextTurn(ctx, ctx.Controller()),
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
