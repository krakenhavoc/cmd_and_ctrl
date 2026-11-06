package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lionheart Glimmer — Enchantment Creature — Cat Glimmer {3}{W}{W}, 2/5
// (Duskmourn: House of Horror, #4466):
//
//	"Ward {2} (Whenever this creature becomes the target of a spell or
//	 ability an opponent controls, counter it unless that player pays
//	 {2}.)
//	 Whenever you attack, creatures you control get +1/+1 until end of
//	 turn."
//
// Ward is the shared trigger (CR 702.21a). "Whenever you attack" is
// Salvation Colossus's shape: one trigger per attack declaration
// (OncePerBatch over EventAttack), whether or not the Glimmer attacks,
// and the creatures pumped are those you control as it resolves (CR
// 611.2c), attacking or not.
//
// No simplification.
func init() {
	const label = "Lionheart Glimmer — creatures you control get +1/+1 until end of turn"
	Register(Spec{
		OracleID:     "bd63bfda-e20a-4476-8a61-cdb51355a227",
		Name:         "Lionheart Glimmer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{2}"), "Lionheart Glimmer — ward {2}"),
			OncePerBatch(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attackDeclaredByYou(ev, source.Controller)
			}, label, Do(BoostUntilEOT{Match: And(Creature(), YouControl()), Power: 1, Toughness: 1, Label: label}))),
		},
	})
}
