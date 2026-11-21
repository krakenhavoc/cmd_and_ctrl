package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Phyrexian Ironworks — Artifact {2}{R}:
//
//	"Whenever you attack, you get {E} (an energy counter).
//	 {T}, Pay {E}{E}{E}: Create a 3/3 colorless Phyrexian Golem artifact
//	 creature token. Activate only as a sorcery."
//
// "Whenever you attack" is once per attack declaration, however many
// creatures attack (OncePerBatch over EventAttack, as Adeline). The
// Golem is written here rather than in the token table: the table's
// "3/3 colorless Golem artifact" is an enchantment Golem, and this one
// is a Phyrexian Golem.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b3c87cb9-a37a-428e-95d1-e6dd90b31097",
		Name:         "Phyrexian Ironworks",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OncePerBatch(effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.AttackDeclaredByYou(ev, source.Controller)
			}, "Phyrexian Ironworks — you get {E}", effects.Do(effects.GetEnergy{N: 1}))),
		},
		Activated: []effects.ActivatedAbility{{
			Label:        "{T}, Pay {E}{E}{E}: Create a 3/3 colorless Phyrexian Golem artifact creature token. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.TapCost(), effects.PayEnergy(3)),
			SorcerySpeed: true,
			Purpose:      game.Purpose{Tokens: 1},
			Effect: effects.Do(effects.CreateToken{N: 1, Template: game.Card{
				Name: "Phyrexian Golem", TypeLine: "Token Artifact Creature — Phyrexian Golem", Power: 3, Toughness: 3,
			}}),
		}},
	})
}
