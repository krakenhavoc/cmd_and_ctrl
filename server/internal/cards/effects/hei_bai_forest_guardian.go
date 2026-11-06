package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hei Bai, Forest Guardian — Legendary Creature — Bear Spirit {3}{G}, 4/4
// (Avatar: The Last Airbender Eternal, #5075):
//
//	"When Hei Bai enters, reveal cards from the top of your library until
//	 you reveal a Shrine card. You may put that card onto the battlefield.
//	 Then shuffle.
//	 {W}{U}{B}{R}{G}, {T}: For each legendary enchantment you control,
//	 create a 1/1 colorless Spirit creature token with \"This token can't
//	 block or be blocked by non-Spirit creatures.\""
//
// The reveal is Songbirds' Blessing's revealUntil (a token is revealed but
// never the hit) and the hit is a "you may put" pick; whether or not it
// entered, and with no Shrine in the library, the library is shuffled.
// The Spirits are Avatar Kuruk's printed row, one for each legendary
// enchantment you control as the ability resolves, read through the
// layered view.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "11e70692-2057-49c8-90e6-953362f4817f",
		Name:         "Hei Bai, Forest Guardian",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Hei Bai, Forest Guardian — reveal until a Shrine card; you may put it onto the battlefield, then shuffle",
				heiBaiReveal),
		},
		Activated: []ActivatedAbility{{
			Label:   "{W}{U}{B}{R}{G}, {T}: For each legendary enchantment you control, create a 1/1 colorless Spirit creature token with \"This token can't block or be blocked by non-Spirit creatures.\"",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Cost:    Plus(ManaCost("{W}{U}{B}{R}{G}"), TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				n := 0
				for _, c := range g.BattlefieldCardsForEffect() {
					if c.Controller == item.Controller && c.IsEnchantment() && c.IsLegendary() {
						n++
					}
				}
				if n == 0 {
					return nil
				}
				return CreateToken{Template: kurukSpiritToken(), N: n}.Apply(NewContext(g, item))
			},
		}},
	})
}

func heiBaiReveal(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	player := ctx.Controller()
	run, hit := revealUntil(ctx, player, func(c game.Card) bool { return c.HasSubtypeWord("Shrine") }, "Hei Bai, Forest Guardian")
	if hit == uuid.Nil {
		return g.ShuffleLibraryForEffect(player)
	}
	return PutFromLibraryOntoBattlefield{
		Player:   player,
		Cards:    run,
		Match:    func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.InstanceID == hit },
		Max:      1,
		Optional: true,
		Label:    "Hei Bai, Forest Guardian — you may put the Shrine card onto the battlefield",
		ThenTail: heiBaiShuffle,
	}.Apply(ctx)
}

// heiBaiShuffle is "Then shuffle", after the put. An on-disk key: never
// renamed, never reused.
var heiBaiShuffle = PutFromLibraryTail("put-from-library/hei-bai-forest-guardian-shuffle", func(ctx *Context, res PutFromLibraryResult) error {
	return ctx.Game.ShuffleLibraryForEffect(res.Player)
})
