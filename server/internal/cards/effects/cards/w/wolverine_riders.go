package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wolverine Riders — Creature — Elf Warrior {4}{G}{G}, 4/4 (EDHREC
// rank 2896):
//
//	"At the beginning of each upkeep, create a 1/1 green Elf Warrior
//	 creature token.
//	 Whenever another Elf you control enters, you gain life equal to
//	 its toughness."
//
// An Elf every upkeep — EVERY upkeep, not just the controller's, so
// a four-player table makes four a round — and lifegain for each Elf
// that arrives, the Riders' own tokens included. The upkeep trigger
// watches EventBeginUpkeep with no owner gate; the lifegain is
// Verdant Sun's Avatar's shape narrowed to Elves (effective
// subtypes, so a changeling counts), reading the entering creature's
// toughness as the trigger resolves, and as it last existed on the
// battlefield if it has since left (CR 608.2h; GainLifeEqualToSubject).
// Both rows declare a footprint, so a batch with Marwyn, the
// Nurturer's counter needs no CR 603.3b order (ADR 0018, #3013).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "09070db6-01f6-4a8a-b167-a7825e6959f4",
		Name:         "Wolverine Riders",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginUpkeep, func(ev game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Kind == game.EventBeginUpkeep
			}, "Wolverine Riders — create a 1/1 green Elf Warrior", effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 green Elf Warrior"), N: 1})),
			{
				Watches: []game.EventKind{game.EventETB},
				Key:     "Wolverine Riders — gain life equal to the Elf's toughness",
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					_, ok := effects.B27AnotherElfYouControlEntered(ev, source, g)
					return ok
				},
				Effect: effects.Do(effects.GainLifeEqualToSubject{Stat: effects.SubjectToughness}),
			},
		},
	})
}
