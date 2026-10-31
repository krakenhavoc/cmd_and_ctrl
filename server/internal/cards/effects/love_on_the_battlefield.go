package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Love on the Battlefield — Enchantment {1}{U}{R} (Marvel Super Heroes
// Commander, #4455):
//
//	"Whenever you attack with exactly two creatures, those creatures gain
//	 first strike until end of turn, then draw a card. Whenever either of
//	 those creatures deals combat damage to a player this combat, put a
//	 +1/+1 counter on it."
//
// One trigger per attack declaration (CR 603.2c) that fires only when the
// attackers are exactly two. As it resolves the two attackers still attacking
// gain first strike, the controller draws, and each of them gets its own
// "this combat" watch (whenThatCreatureHitsAPlayer), so a creature that hits a
// player grows and a creature that was blocked does not. The watch ends with
// the combat phase and is bound to the creature as the object it is now.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "645dbd86-fe30-4897-a3ca-633f07ec5b54",
		Name:         "Love on the Battlefield",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventAttack, loveOnTheBattlefieldExactlyTwo,
				"Love on the Battlefield — the two attackers gain first strike, draw a card, and grow when they hit a player",
				loveOnTheBattlefieldResolve)),
		},
	})
}

// loveOnTheBattlefieldExactlyTwo is "you attack with exactly two creatures".
func loveOnTheBattlefieldExactlyTwo(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return attackDeclaredByYou(ev, source.Controller) && len(b13AttackingCreaturesYouControl(g, source.Controller)) == 2
}

func loveOnTheBattlefieldResolve(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item).asGroupMember()
	pair := b13AttackingCreaturesYouControl(g, item.Controller)
	for _, id := range pair {
		if err := (GrantKeywordUntilEOT{Target: id, Keywords: []string{"first strike"}, Label: "Love on the Battlefield — first strike"}).Apply(ctx); err != nil {
			return err
		}
	}
	if err := (DrawCards{Player: item.Controller, N: 1}).Apply(ctx); err != nil {
		return err
	}
	until, ok := DurationUntilEndOfCombat(ctx)
	if !ok {
		return nil
	}
	for _, id := range pair {
		if err := whenThatCreatureHitsAPlayer(ctx, id, "Love on the Battlefield — put a +1/+1 counter on it",
			false, until, plusOneCounterOnThePayloadBody); err != nil {
			return err
		}
	}
	return nil
}
