package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "b3c87cb9-a37a-428e-95d1-e6dd90b31097",
		Name:         "Phyrexian Ironworks",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attackDeclaredByYou(ev, source.Controller)
			}, "Phyrexian Ironworks — you get {E}", Do(GetEnergy{N: 1}))),
		},
		Activated: []ActivatedAbility{{
			Label:        "{T}, Pay {E}{E}{E}: Create a 3/3 colorless Phyrexian Golem artifact creature token. Activate only as a sorcery.",
			Cost:         Plus(TapCost(), PayEnergy(3)),
			SorcerySpeed: true,
			Purpose:      game.Purpose{Tokens: 1},
			Effect: Do(CreateToken{N: 1, Template: game.Card{
				Name: "Phyrexian Golem", TypeLine: "Token Artifact Creature — Phyrexian Golem", Power: 3, Toughness: 3,
			}}),
		}},
	})
}
