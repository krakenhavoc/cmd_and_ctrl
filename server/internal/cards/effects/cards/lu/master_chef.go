package cardslu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Master Chef — Legendary Enchantment — Background {2}{G} (Commander Legends:
// Battle for Baldur's Gate, #5948):
//
//	"Commander creatures you own have "This creature enters with an additional
//	 +1/+1 counter on it" and "Other creatures you control enter with an
//	 additional +1/+1 counter on them.""
//
// Two granted replacement effects (ADR 0093, AbilityGrant.Replacements): CR
// 614 entry replacements that add a counter to the entry event, so Hardened
// Scales and Doubling Season apply to it (Bloodspore Thrinax's second
// ability). The first is the commander's own entry, the second another creature
// entering under the commander's controller's control, "you" being whoever
// controls the commander. Choose a Background is deck construction (ADR 0144).
//
// No simplification.
const masterChefGrant = "master-chef/counters"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "9eab87ff-82a0-4655-b6ea-00b27e882449",
		Name:         "Master Chef",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: masterChefGrant,
			Replacements: []game.ReplacementEffect{
				{
					Watches: []game.EventKind{game.EventZoneMove},
					AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
						return ev.Kind == game.RepEventMove && ev.NewZone == game.ZoneBattlefield && src != nil && ev.CardID == src.InstanceID
					},
					Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
						ev.AddCounterAtETB(game.CounterPlusOne, 1)
						return nil
					},
					Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID { return src.Controller },
					Label:      "Master Chef: this creature enters with an additional +1/+1 counter",
				},
				{
					Watches:   []game.EventKind{game.EventZoneMove},
					AppliesTo: effects.BnsB13OtherCreatureYouControlEnters,
					Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
						ev.AddCounterAtETB(game.CounterPlusOne, 1)
						return nil
					},
					Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID { return src.Controller },
					Label:      "Master Chef: other creatures you control enter with an additional +1/+1 counter",
				},
			},
			Text: "This creature enters with an additional +1/+1 counter on it. Other creatures you control enter with an additional +1/+1 counter on them.",
		}},
		Static: []game.StaticAbility{effects.GrantToCommanderCreaturesYouOwn(masterChefGrant)},
	})
}
