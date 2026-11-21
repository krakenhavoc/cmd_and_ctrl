package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Old One Eye — Legendary Creature — Tyranid {5}{G}, 6/6:
//
//	"Trample
//	 Other creatures you control have trample.
//	 When Old One Eye enters, create a 5/5 green Tyranid creature token.
//	 Fast Healing — At the beginning of your first main phase, you may discard
//	 two cards. If you do, return this card from your graveyard to your hand."
//
// Fast Healing is a graveyard trigger. "You may discard two cards" is two or
// none (a keyed discard with the exactly-two set rule), and the return runs
// only for the same graveyard object that triggered (CR 400.7).
func init() {
	effects.Register(effects.Spec{
		OracleID:        "86ef2ceb-b7e6-4451-92ec-8528ea191d3a",
		Name:            "Old One Eye",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Static: []game.StaticAbility{
			effects.Labelled(effects.KeywordGrant(effects.OtherCreaturesYouControl, "trample"), "Other creatures you control have trample."),
		},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.WhenThisEnters("Old One Eye — create a 5/5 green Tyranid creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("5/5 green Tyranid"), N: 1})),
				game.Purpose{Tokens: 1, CreatureTokens: 1}),
			effects.InGraveyard(effects.AtYourPrecombatMain("Old One Eye — you may discard two cards; if you do, return it to your hand", oldOneEyeAsk)),
		},
	})
}

func oldOneEyeAsk(g *game.Game, item *game.StackItem) error {
	if !effects.CardObjectStillInGraveyard(g, item) {
		return nil
	}
	return g.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
		Player:   item.Controller,
		Source:   item.SourceCardID,
		N:        2,
		UpTo:     true,
		Question: "Old One Eye — discard two cards to return it to your hand (two or none)",
		Asked:    game.NewAskedItem(item),
	}, oldOneEyeDiscarded, nil)
}

var oldOneEyeDiscarded = game.RegisterDiscardThenRule("choose-cards/old-one-eye-discard-two",
	func(picked []game.Card) bool { return len(picked) == 2 },
	func(g *game.Game, r game.DiscardsDone) error {
		if r.Discarded.Count() != 2 || r.Item == nil || !effects.CardObjectStillInGraveyard(g, r.Item) {
			return nil
		}
		return effects.ReturnFromGraveyard{Target: r.Item.SourceCardID, Dest: game.ZoneHand}.Apply(effects.NewContext(g, r.Item))
	})
