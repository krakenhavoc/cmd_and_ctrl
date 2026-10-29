package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Sackville-Bagginses — Legendary Creature — Halfling Citizen {1}{B},
// 2/2 (The Hobbit, #4454):
//
//	"When The Sackville-Bagginses enter, you may sacrifice another
//	 creature or artifact. If you do, draw a card and create a Treasure
//	 token.
//	 Whenever you sacrifice a token, target opponent loses 1 life."
//
// The sacrifice is asked as the trigger resolves (maySacrificeOne), and the
// draw and the Treasure follow only a permanent that really went. The
// second ability watches EventSacrifice, which fires before the zone move
// (docs/adding-cards.md's trigger table), so the sacrificed permanent is
// still there to be read as a token; any token counts, a Treasure included,
// and so does the one this card's own first ability sacrifices. The opponent
// is the trigger's target, picked as it goes on the stack.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e3a665f9-6e51-4e0d-923b-e9552d5978a4",
		Name:         "The Sackville-Bagginses",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("The Sackville-Bagginses — sacrifice another creature or artifact to draw a card and create a Treasure", sackvilleBagginsesSacrifice),
			Targeting(
				On(game.EventSacrifice, sackvilleBagginsesSacrificedAToken, "The Sackville-Bagginses — target opponent loses 1 life", sackvilleBagginsesDrain),
				TargetPlayer("target opponent", Opponent())),
		},
	})
}

// sackvilleBagginsesSacrifice is the enters trigger: "you may sacrifice
// another creature or artifact".
func sackvilleBagginsesSacrifice(g *game.Game, item *game.StackItem) error {
	return maySacrificeOne(g, item, "The Sackville-Bagginses — sacrifice another creature or artifact to draw a card and create a Treasure?",
		Or(Creature(), Artifact()), true, sackvilleBagginsesSacrificed)
}

// sackvilleBagginsesSacrificed is the "if you do": draw a card and create a
// Treasure once a permanent has really gone, keyed (#4290).
var sackvilleBagginsesSacrificed = RegisterSacrificedOne("prompt-run/the-sackville-bagginses-sacrifice", func(ctx *Context, _ uuid.UUID) error {
	if err := (DrawCards{Player: ctx.Controller(), N: 1}).Apply(ctx); err != nil {
		return err
	}
	return CreateToken{Template: TreasureToken(), N: 1}.Apply(ctx)
})

// sackvilleBagginsesSacrificedAToken is "you sacrifice a token".
func sackvilleBagginsesSacrificedAToken(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Actor != source.Controller || ev.CardID == uuid.Nil {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && IsToken(c)
}

// sackvilleBagginsesDrain is the second ability's body: the targeted
// opponent loses 1 life. An opponent who has left by then is nothing to do
// (CR 608.2b).
func sackvilleBagginsesDrain(g *game.Game, item *game.StackItem) error {
	for _, t := range NewContext(g, item).LegalTargets() {
		if t.Kind == game.TargetPlayer {
			return g.ChangePlayerLifeForEffect(item.SourceCardID, t.ID, -1)
		}
	}
	return nil
}
