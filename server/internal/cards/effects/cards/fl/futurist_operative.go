package cardsfl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Futurist Operative — Creature — Human Ninja {3}{U}, 3/4 (Kamigawa: Neon
// Dynasty, #5493):
//
//	"As long as this creature is tapped, it's a Human Citizen with base
//	 power and toughness 1/1 and can't be blocked.
//	 {2}{U}: Untap this creature."
//
// One printed sentence is three effects that each hold only while the
// Operative is tapped: a layer 4 replacement of its creature types
// (Human Citizen, words of its text, CR 612.2; the card types stay), a
// layer 7b base 1/1 (counters and pumps still apply on top), and a
// "can't be blocked" block rule read off the attacker when the block is
// checked (Thieves' Tools' CantBeBlockedWhile), so it is evasive exactly
// while it is tapped, an attacking Operative without vigilance included.
// The untap is Axgard Braggart's.
//
// No simplification.
func init() {
	tapped := func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return target.InstanceID == source.InstanceID && target.Tapped
	}
	effects.Register(effects.Spec{
		OracleID:     "9f53aac8-f21f-4928-af8f-65b40cb9149c",
		Name:         "Futurist Operative",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			{
				Label:     "As long as this creature is tapped, it's a Human Citizen.",
				Layer:     game.Layer4Type,
				AppliesTo: tapped,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
					c.SetCreatureTypes(source.TextSubtypes("Human", "Citizen"))
				},
			},
			{
				Label:     "As long as this creature is tapped, it has base power and toughness 1/1.",
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7B_Set,
				AppliesTo: tapped,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Power, c.Toughness = 1, 1
				},
			},
		},
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedWhile(effects.OnSelf(), func(_ *game.Game, _ uuid.UUID, attacker game.Card) bool { return attacker.Tapped }),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{U}: Untap this creature.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{2}{U}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.UntapTarget{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
