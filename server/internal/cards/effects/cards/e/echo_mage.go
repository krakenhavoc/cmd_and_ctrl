package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Echo Mage — Creature — Human Wizard {1}{U}{U}, 2/3 (Commander 2013, S82
// older sets):
//
//	"Level up {1}{U} ({1}{U}: Put a level counter on this. Level up only as
//	 a sorcery.)
//	 LEVEL 2-3
//	 2/4
//	 {U}{U}, {T}: Copy target instant or sorcery spell. You may choose new
//	 targets for the copy.
//	 LEVEL 4+
//	 2/5
//	 {U}{U}, {T}: Copy target instant or sorcery spell twice. You may choose
//	 new targets for the copies."
//
// A leveler (leveler.go, CR 702.87a, 711.2). Each band's activated ability is
// a bundle the Mage grants itself behind that band's level check (Guul Draz
// Assassin's shape), so it has at most one of them and none below level 2.
// The copies go on the stack above the original, each with new targets
// chosen by the Mage's controller (CR 707.10).
//
// No simplification.
const (
	EchoMageCopyOnce  = "echo-mage/copy-once"
	EchoMageCopyTwice = "echo-mage/copy-twice"
)

func echoMageGrant(key string, n int, text string) effects.AbilityGrant {
	return effects.AbilityGrant{
		Key: key,
		Activated: []effects.ActivatedAbility{{
			Label: text,
			Cost:  effects.Plus(effects.ManaCost("{U}{U}"), effects.TapCost()),
			Targets: effects.TargetSpell("target instant or sorcery spell",
				effects.Or(effects.Instant(), effects.Sorcery())),
			Effect: effects.OldB381CopyTargetSpellTimes(n),
		}},
		Text: text,
	}
}

func init() {
	once := "{U}{U}, {T}: Copy target instant or sorcery spell. You may choose new targets for the copy."
	twice := "{U}{U}, {T}: Copy target instant or sorcery spell twice. You may choose new targets for the copies."
	effects.Register(effects.Spec{
		OracleID:     "fb98f6b7-5986-4c5d-98fc-e5c4106f48bf",
		Name:         "Echo Mage",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{
			echoMageGrant(EchoMageCopyOnce, 1, once),
			echoMageGrant(EchoMageCopyTwice, 2, twice),
		},
		Activated: []effects.ActivatedAbility{effects.LevelUpKeyword("{1}{U}")},
		Static: effects.LevelBands(
			effects.LevelBand{From: 2, To: 3, Power: 2, Toughness: 4, Statics: []game.StaticAbility{
				effects.Labelled(effects.GrantAbilities(effects.SelfOnly, EchoMageCopyOnce), `This creature has "`+once+`"`),
			}},
			effects.LevelBand{From: 4, Power: 2, Toughness: 5, Statics: []game.StaticAbility{
				effects.Labelled(effects.GrantAbilities(effects.SelfOnly, EchoMageCopyTwice), `This creature has "`+twice+`"`),
			}},
		),
	})
}
