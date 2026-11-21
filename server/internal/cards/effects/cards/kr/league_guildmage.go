package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// League Guildmage — Creature — Human Wizard {U}{R}, 2/2 (Ravnica
// Allegiance guild-mage cycle, Mystery Booster):
//
//	"{3}{U}, {T}: Draw a card.
//	 {X}{R}, {T}: Copy target instant or sorcery spell you control with mana
//	 value X. You may choose new targets for the copy."
//
// The draw is a plain tap ability. The copy is Izzet Guildmage's CopySpell
// with the new-targets offer (CR 707.10c), over an instant or sorcery spell
// you control whose mana value equals the X announced with the activation
// (CR 202.3, read off the spell as it sits on the stack); an activated
// ability on the stack cannot be targeted, only the spell. The copy is put
// onto the stack, not cast (CR 707.10).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "08bb1f33-b232-4382-a268-4804d4e72d29",
		Name:         "League Guildmage",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{3}{U}, {T}: Draw a card.",
				Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
				Cost:    effects.Plus(effects.ManaCost("{3}{U}"), effects.TapCost()),
				Effect:  effects.Do(effects.DrawCards{N: 1}),
			},
			{
				Label: "{X}{R}, {T}: Copy target instant or sorcery spell you control with mana value X. You may choose new targets for the copy.",
				Cost:  effects.Plus(effects.ManaCost("{X}{R}"), effects.TapCost()),
				Targets: effects.TargetSpell("target instant or sorcery spell you control with mana value X",
					effects.Or(effects.Instant(), effects.Sorcery()), effects.YouControl()).WithManaValueEqualsX(),
				Effect: effects.CopyTheFirstLegalTargetSpell,
			},
		},
	})
}
