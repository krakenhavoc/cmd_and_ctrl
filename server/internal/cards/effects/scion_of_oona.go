package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Scion of Oona — Creature — Faerie Soldier {2}{U}, 1/1 (Wilds of
// Eldraine Commander, #5368):
//
//	"Flash
//	 Flying
//	 Other Faerie creatures you control get +1/+1.
//	 Other Faeries you control have shroud."
//
// The anthem is the tribal lord's (TribalAnthem): layer 7c, "Faerie" read
// through the effective view so a changeling counts. The shroud line says
// "Faeries", not "Faerie creatures", so its scope is every other Faerie
// PERMANENT you control (layer 6), not only creatures. The Scion has
// neither for itself.
//
// No simplification.
func init() {
	anthem := TribeFilter{Tribes: []string{"Faerie"}, Others: true, YoursOnly: true}
	shroud := KeywordGrant(func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return source != nil && target.InstanceID != source.InstanceID && target.Controller == source.Controller &&
			target.HasSubtype(source.TextSubtype("Faerie"))
	}, "shroud")
	shroud.Label = "Other Faeries you control have shroud."
	Register(Spec{
		OracleID:        "b34601d2-2aeb-4649-9d73-d0895611bd69",
		Name:            "Scion of Oona",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Static:          []game.StaticAbility{TribalAnthem(anthem, 1, 1), shroud},
	})
}
