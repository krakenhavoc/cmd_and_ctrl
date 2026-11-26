package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lord of Shatterskull Pass — {3}{R} Creature — Minotaur Shaman 3/3 (Rise of the Eldrazi, S82 OLD #5581):
//
//	Level up {1}{R}
//	LEVEL 1-5: 6/6
//	LEVEL 6+: 6/6, Whenever this creature attacks, it deals 6 damage to each creature defending player controls.
//
// A leveler: Level up is the keyword, each level symbol a band of LevelBands.
//
// The attack trigger is a granted ability held by the 6+ band, so it exists
// only at level 6 and above; the 6 damage is Gouged Zealot's sweep body.
const lordOfShatterskullPassGrant = "lord-of-shatterskull-pass/attack-sweep"

func init() {
	sweep := effects.On(game.EventAttack, effects.ThisAttacked,
		"Lord of Shatterskull Pass — 6 damage to each creature defending player controls",
		effects.S82OldB38DamageEachCreatureDefendingPlayerControls(6))
	sweep.Purpose = game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 6, OpponentsOnly: true}}
	effects.Register(effects.Spec{
		OracleID:     "d369a02b-c282-454a-9f89-480f4f5d838c",
		Name:         "Lord of Shatterskull Pass",
		Completeness: effects.CompletenessFull,
		Purpose:      sweep.Purpose,
		Grants: []effects.AbilityGrant{{
			Key:       lordOfShatterskullPassGrant,
			Triggered: []game.TriggeredAbility{sweep},
			Text:      "Whenever this creature attacks, it deals 6 damage to each creature defending player controls.",
		}},
		Activated: []effects.ActivatedAbility{effects.LevelUpKeyword("{1}{R}")},
		Static: effects.LevelBands(
			effects.LevelBand{From: 1, To: 5, Power: 6, Toughness: 6},
			effects.LevelBand{From: 6, Power: 6, Toughness: 6, Statics: []game.StaticAbility{
				effects.Labelled(effects.GrantAbilities(effects.SelfOnly, lordOfShatterskullPassGrant),
					`This creature has "Whenever this creature attacks, it deals 6 damage to each creature defending player controls."`),
			}},
		),
	})
}
