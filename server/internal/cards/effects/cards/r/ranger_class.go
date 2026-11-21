package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ranger Class — Enchantment — Class {1}{G} (Adventures in the Forgotten
// Realms):
//
//	"(Gain the next level as a sorcery to add its ability.)
//	 When this Class enters, create a 2/2 green Wolf creature token.
//	 {1}{G}: Level 2
//	 Whenever you attack, put a +1/+1 counter on target attacking creature.
//	 {3}{G}: Level 3
//	 You may look at the top card of your library any time.
//	 You may cast creature spells from the top of your library."
//
// One line per level (ADR 0071):
//
//   - Level 1 is an entry trigger making the table's Wolf token, ungated
//     (CR 716.3). It fires as the Class enters, at level 1 (CR 716.2b).
//   - Level 2 is Battlesong Berserker's "whenever you attack" (one trigger
//     per declaration, OncePerBatch), gated at level 2, targeting any
//     attacking creature. A target that stopped attacking by resolution
//     gets nothing (CR 608.2b).
//   - Level 3 is Vizier of the Menagerie's private look at the top card and
//     its creature-spells-only permission to cast from it at the printed
//     cost, both gated at level 3 (Spec.LibraryTopVisibleWhen and
//     GatedCastPermissions). Ordinary timing applies (CR 401.5).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:              "d1586808-efc2-469a-8384-ba4acf231e69",
		Name:                  "Ranger Class",
		Completeness:          effects.CompletenessFull,
		LibraryTopVisible:     game.LibraryTopOwner,
		LibraryTopVisibleWhen: effects.Level(3),
		GatedCastPermissions: []game.CastPermissionGate{{
			Permission: effects.PlayFromTopOfYourLibrary(game.PermissionFilter{CreatureOnly: true},
				"Cast a creature spell from the top of your library (Ranger Class)"),
			ActiveWhen: effects.Level(3),
		}},
		Activated: []effects.ActivatedAbility{
			effects.LevelUp(2, effects.ManaCost("{1}{G}")),
			effects.LevelUp(3, effects.ManaCost("{3}{G}")),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Ranger Class — create a 2/2 green Wolf creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 green Wolf"), N: 1})),
			effects.AtLevel(2, effects.OncePerBatch(effects.Targeting(
				effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return effects.AttackDeclaredByYou(ev, source.Controller)
				}, "Ranger Class — put a +1/+1 counter on target attacking creature",
					effects.PutPlusOneCounterOnEachLegalTarget),
				effects.TargetCreature("target attacking creature", effects.AttackingCreature())))),
		},
	})
}
