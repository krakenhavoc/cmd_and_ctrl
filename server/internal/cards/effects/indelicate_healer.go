package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Indelicate Healer — Creature — Klingon Warrior Doctor {2}{W}, 3/3 (Star
// Trek Commander, #4974):
//
//	"Whenever this creature or another Klingon or Warrior you control
//	 enters, creatures you control gain lifelink until end of turn."
//
// One trigger per entering permanent: the Healer itself, or another
// Klingon or Warrior under your control, its creature types read through
// the layers as it enters. The two type words are read through the
// Healer's own text (CR 612.2), so a text-changing effect reaches them. The lifelink grant is a layer-6 effect whose
// set is locked as the ability resolves (CR 611.2c), so a creature that
// enters afterwards does not gain it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9d18253d-a38e-48a4-80bf-aa8fe5283584",
		Name:         "Indelicate Healer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, indelicateHealerEntry,
				"Indelicate Healer — creatures you control gain lifelink until end of turn",
				Do(GrantKeywordUntilEOT{
					Match:    And(Creature(), YouControl()),
					Keywords: []string{"lifelink"},
					Label:    "Indelicate Healer — lifelink",
				})),
		},
	})
}

// indelicateHealerEntry is "this creature or another Klingon or Warrior
// you control enters".
func indelicateHealerEntry(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := enteredUnderYourControl(ev, source, g, false)
	if !ok {
		return false
	}
	return c.InstanceID == source.InstanceID || c.HasSubtype(source.TextSubtype("Klingon")) || c.HasSubtype(source.TextSubtype("Warrior"))
}
