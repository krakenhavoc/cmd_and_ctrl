package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Grand Warlord Radha — Legendary Creature — Elf Warrior {2}{R}{G}, 3/4
// (Dominaria, S82 OLD, #5530):
//
//	"Haste
//	 Whenever one or more creatures you control attack, add that much mana in
//	 any combination of {R} and/or {G}. Until end of turn, you don't lose this
//	 mana as steps and phases end."
//
// One trigger per attack declaration of yours, whether or not Radha is among
// the attackers (Ancestor Dragon's "whenever you attack" batch). "That much" is
// the number of creatures its controller controls that are attacking, counted
// as the trigger resolves; each mana is the player's pick of {R} or {G}, and
// it is kept through the turn's steps and phases (KeepManaUntilEndOfTurn).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0fb90df6-a62b-4a9f-ae79-db71d40466ce",
		Name:            "Grand Warlord Radha",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			effects.OncePerBatch(effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.AttackDeclaredByYou(ev, source.Controller)
			}, "Grand Warlord Radha — add {R} and/or {G} for each attacking creature; you don't lose this mana as steps and phases end",
				effects.S82OldB213RadhaMana)),
		},
	})
}
