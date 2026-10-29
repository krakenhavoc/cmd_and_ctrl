package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Necklace of Girion — Legendary Artifact {2}{G} (The Hobbit Eternal, HOC):
//
//	"Whenever you cast a green spell and whenever a Forest you control
//	 enters, put a +1/+1 counter on target creature you control.
//	 {T}: Add {G}."
//
// Two triggers that share the counter (Hunting Moa's body). The spell's
// colour is read through the Necklace's text (CR 612.2, as WheneverYouCast
// does); the Forest is judged by its effective subtypes, as Roiling Canopy
// judges it, so a Forest that changed type is still one. The target is a
// creature you control, chosen as each trigger goes on the stack and
// re-checked at resolution (CR 608.2b); with none the trigger is not put on
// the stack.
//
// No simplification.
func init() {
	const counter = "put a +1/+1 counter on target creature you control"
	green := WheneverYouCast(OfColor("G"), "Necklace of Girion — "+counter+" (you cast a green spell)",
		putPlusOneCounterOnEachLegalTarget)
	green.Targets = TargetCreature("target creature you control", YouControl())
	forest := On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		c, ok := enteredUnderYourControl(ev, source, g, false)
		return ok && c.IsLand() && c.HasSubtype(source.TextSubtype("Forest"))
	}, "Necklace of Girion — "+counter+" (a Forest entered)", putPlusOneCounterOnEachLegalTarget)
	forest.Targets = TargetCreature("target creature you control", YouControl())
	Register(Spec{
		OracleID:     "9f970385-e7c6-40bf-999e-7015693c2250",
		Name:         "Necklace of Girion",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{G}",
			Label:    "Add {G}",
		}},
		Triggered: []game.TriggeredAbility{green, forest},
	})
}
