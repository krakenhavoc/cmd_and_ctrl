package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Harmonic Prodigy — Creature — Human Wizard {1}{R}, 1/3:
//
//	"Prowess (Whenever you cast a noncreature spell, this creature gets
//	 +1/+1 until end of turn.)
//	 If a triggered ability of a Shaman or another Wizard you control
//	 triggers, that ability triggers an additional time."
//
// Prowess is the canonical engine keyword (game/prowess.go), declared
// through PrintedKeywords like Pinnacle Monk. The second line is Roaming Throne's doubler shape (DoublesAbilitiesOf
// with SourceMatch): the source must be a Shaman (the Prodigy itself
// qualifies only if some effect makes it one) or a Wizard other than
// the Prodigy, and the default controller check keeps it to creatures
// you control. Types are read from the source's last-known
// characteristics, so a changeling counts and a dies-trigger is judged
// by what the creature was on the battlefield. Two Prodigies each
// double the other's triggers, giving three copies, as the rules say.
//
// No simplification.
func init() {
	d := DoublesAbilitiesOf(nil, DoublesAbilitiesOfOptions{
		SourceMatch: func(_ *game.Game, q game.TriggerDoublingQuery) bool {
			src := characteristicCard(q.Source.InstanceID, q.Source, q.SourceLKI)
			if src.HasSubtype("Shaman") {
				return true
			}
			return q.Source.InstanceID != q.Doubler.InstanceID && src.HasSubtype("Wizard")
		},
	})
	d.Label = "Harmonic Prodigy"
	Register(Spec{
		OracleID:        "2e2ace5b-4018-43af-8e72-ebafec1a7739",
		Name:            "Harmonic Prodigy",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordProwess},
		TriggerDoublers: []game.TriggerDoubler{d},
	})
}
