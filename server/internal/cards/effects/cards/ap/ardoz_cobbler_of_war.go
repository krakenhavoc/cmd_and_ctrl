package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ardoz, Cobbler of War — Legendary Creature — Goblin Shaman {1}{R}, 1/1
// (Jumpstart 2022, #5948):
//
//	"Haste
//	 Whenever Ardoz or another creature you control enters, that creature
//	 gets +2/+0 until end of turn.
//	 {3}{R}: Create a 1/1 red Goblin creature token with haste. Activate
//	 only as a sorcery."
//
// Kor Celebrant's trigger: one per creature entering under your control,
// Ardoz included, with the entering creature as the trigger's object (Ogre
// Battledriver's read): the boost goes to it only while it is still on the
// battlefield as the ability resolves. The token Ardoz makes is a creature
// entering under your control, so it triggers Ardoz and gets +2/+0 too.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "bcd5e175-2ef0-48f4-b9bd-14383dd234e0",
		Name:            "Ardoz, Cobbler of War",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.CreatureEnteredUnderYourControl,
				"Ardoz, Cobbler of War — that creature gets +2/+0 until end of turn", ardozPump),
		},
		Activated: []effects.ActivatedAbility{{
			Label:        "{3}{R}: Create a 1/1 red Goblin creature token with haste. Activate only as a sorcery.",
			Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
			Cost:         effects.ManaCost("{3}{R}"),
			SorcerySpeed: true,
			Effect:       effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 red Goblin with haste"), N: 1}),
		}},
	})
}

// ardozPump is "that creature gets +2/+0 until end of turn" for the creature
// whose entry triggered the ability, while it is still on the battlefield.
func ardozPump(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	info, ok := ctx.TriggeringPermanent()
	if !ok || info.Left {
		return nil
	}
	id := ctx.Trigger().Object.Ref().ID
	return effects.BoostUntilEOT{Target: id, Power: 2, Label: "Ardoz, Cobbler of War — +2/+0"}.Apply(ctx)
}
