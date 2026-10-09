package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Master of Barbs — Creature — Lizard Bard {1}{R}, 2/1:
//
//	"Menace
//	 Whenever one or more opponents are dealt noncombat damage,
//	 creatures you control get +1/+0 until end of turn."
//
// "One or more" is OncePerBatch: damage that hits three opponents in one
// resolution is a single trigger. The creatures are the ones you control
// when the trigger resolves (CR 611.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e8c36337-135c-460d-a1da-cc01743dda71",
		Name:            "Master of Barbs",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return !ev.Combat && ev.Amount > 0 && ev.Target != source.Controller &&
					g.PlayerByIDForEffect(ev.Target) != nil
			}, "Master of Barbs — creatures you control get +1/+0 until end of turn",
				Do(BoostUntilEOT{Match: And(Creature(), YouControl()), Power: 1, Label: "Master of Barbs — +1/+0"}))),
		},
	})
}
