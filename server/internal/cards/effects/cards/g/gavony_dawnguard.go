package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gavony Dawnguard — {1}{W}{W} Creature — Human Soldier 3/3 (#2586, ADR
// 0132):
//
//	"Ward {1}
//	 If it's neither day nor night, it becomes day as this creature
//	 enters.
//	 Whenever day becomes night or night becomes day, look at the top
//	 four cards of your library. You may reveal a creature card with mana
//	 value 3 or less from among them and put it into your hand. Put the
//	 rest on the bottom of your library in any order."
//
// Oath of Nissa's dig: the look is private, the card taken is revealed,
// and the rest go under in an order the controller picks.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2788cc1e-18b6-4506-be4c-f4056600e33c",
		Name:         "Gavony Dawnguard",
		Completeness: effects.CompletenessFull,
		AsEnters:     effects.BecomesDayAsEnters(),
		Triggered: []game.TriggeredAbility{
			effects.Ward(effects.WardMana("{1}"), "Gavony Dawnguard — ward {1}"),
			effects.WheneverDayBecomesNightOrNightBecomesDay("Gavony Dawnguard — look at the top four cards of your library",
				gavonyDawnguardDig),
		},
	})
}

func gavonyDawnguardDig(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	player := ctx.Controller()
	return effects.TakeFromLibraryToHand{
		Player:   player,
		Cards:    g.LookAtTopOfLibraryForEffect(player, 4),
		Match:    effects.And(effects.Creature(), effects.ManaValueLE(3)),
		Max:      1,
		Optional: true,
		Label:    "Gavony Dawnguard — you may reveal a creature card with mana value 3 or less and put it into your hand",
		Key:      effects.TakeRevealRestBottomAny,
	}.Apply(ctx)
}
