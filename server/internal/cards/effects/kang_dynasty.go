package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kang Dynasty — Enchantment — Saga {3}{U} (Marvel Super Heroes Commander,
// #4455):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I, II — For each opponent, tap up to one target creature that player
//	 controls. Goad those creatures. Until your next turn, whenever any of
//	 those creatures deals combat damage to a player, draw a card.
//	 III — Target creature you control gets +1/+1 until end of turn for each
//	 card in your hand and can't be blocked this turn."
//
// I and II are the per-opponent "up to one" clauses of Windgrace's Judgment
// (perOpponentClauses): each pick still legal as it resolves is tapped and
// goaded (CR 701.15), and each gets its own watch until the saga controller's
// next turn (whenThatCreatureHitsAPlayer), so a goaded creature that connects
// with any player, you included, draws you a card. III counts the cards in
// hand as it resolves.
//
// No simplification.
func init() {
	chapter := func(n int) game.TriggeredAbility {
		return ChapterTriggerTargeting(n, SagaChapterLabel("Kang Dynasty", n, "tap and goad a creature of each opponent; draw when they connect"),
			perOpponentClauses(game.MaxPlayers-1, "creature", Creature()), kangDynastyTapGoad)
	}
	Register(Spec{
		OracleID:     "4e420f43-a4a3-44bb-9a4e-a943f93a3534",
		Name:         "Kang Dynasty",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			chapter(1),
			chapter(2),
			ChapterTriggerTargeting(3, SagaChapterLabel("Kang Dynasty", 3, "+1/+1 for each card in your hand and can't be blocked"),
				TargetCreature("target creature you control", YouControl()), kangDynastyFinale),
		},
	})
}

func kangDynastyTapGoad(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var ids []uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			ids = append(ids, t.ID)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	for _, id := range ids {
		if err := (TapTarget{Target: id}).Apply(ctx); err != nil {
			return err
		}
	}
	if err := goadAndScheduleClear(ctx, ids); err != nil {
		return err
	}
	until := DurationUntilYourNextTurn(ctx, item.Controller)
	for _, id := range ids {
		if err := whenThatCreatureHitsAPlayer(ctx, id, "Kang Dynasty — draw a card", false, until, drawACardForTheControllerBody); err != nil {
			return err
		}
	}
	return nil
}

func kangDynastyFinale(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	hand := 0
	if p := g.PlayerByIDForEffect(item.Controller); p != nil {
		hand = p.Hand.Size()
	}
	if err := (BoostUntilEOT{Target: id, Power: hand, Toughness: hand, Label: "Kang Dynasty — +1/+1 for each card in your hand"}).Apply(ctx); err != nil {
		return err
	}
	return RestrictUntilEOT{Target: id, Restrictions: game.CantBeBlocked, Label: "Kang Dynasty — can't be blocked"}.Apply(ctx)
}
