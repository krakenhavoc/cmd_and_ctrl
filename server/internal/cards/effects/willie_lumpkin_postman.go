package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Willie Lumpkin, Postman — Legendary Creature — Human Citizen {W}{U}, 1/3
// (Marvel Super Heroes Commander, #4455):
//
//	"Willie Lumpkin can't be blocked.
//	 Whenever Willie Lumpkin deals combat damage to an opponent, you draw a
//	 card and that player may draw a card. If they do, that player can't
//	 attack you or permanents you control during their next turn."
//
// "Can't be blocked" is Blighted Agent's Restriction bit. You draw first,
// then the damaged player is asked (MayChoice, which also refuses the
// choice when a draw limit stops them drawing, CR 121.3). On a yes they
// draw and The Second Doctor's restriction is put on them
// (GrantCantAttackPlayerForEffect: their next turn, you and your
// permanents). "If they do" is the yes: a player with an empty library who
// agrees still draws nothing and is still restricted, which is the printed
// choice.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b08633e9-7b8f-4f69-a3b7-8d982b17574c",
		Name:         "Willie Lumpkin, Postman",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{RestrictSelf(game.CantBeBlocked)},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer(
				"Willie Lumpkin, Postman — you draw a card and that player may draw a card; if they do, they can't attack you or your permanents during their next turn",
				willieLumpkinHits),
		},
	})
}

func willieLumpkinHits(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := (DrawCards{N: 1}).Apply(ctx); err != nil {
		return err
	}
	victim := item.Trigger.Event.Target
	if p := g.PlayerByIDForEffect(victim); p == nil || p.Eliminated || victim == item.Controller {
		return nil
	}
	return MayChoice{
		Player:   victim,
		Question: "Willie Lumpkin, Postman — draw a card? (you then can't attack Willie's controller or their permanents during your next turn)",
		Draws:    1,
		Then:     willieLumpkinAnswer,
	}.Apply(ctx)
}

// willieLumpkinAnswer is the damaged player's answer, keyed: on a yes they
// draw and cannot attack the controller during their next turn.
var willieLumpkinAnswer = MayAnswer("confirm/willie-lumpkin-postman-draw", func(ctx *Context, r game.ConfirmAnswer) error {
	if !r.Accepted {
		return nil
	}
	if err := (DrawCards{Player: r.Chooser, N: 1}).Apply(ctx); err != nil {
		return err
	}
	ctx.Game.GrantCantAttackPlayerForEffect(r.Chooser, ctx.Controller(),
		"Willie Lumpkin, Postman — can't attack its controller or their permanents", ctx.Source())
	return nil
})
