package effects

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// escalate.go — CR 702.120a: "Escalate [cost]. For each mode you
// choose beyond the first as you cast this spell, you pay an
// additional [cost]." (#2126)
//
// The cost is declared once on the spell's ModeSpec and owed
// (len(modes) - 1) times. It rides the cast's one CR 601.2f payment
// plan, so everything an additional cost already does applies per
// extra mode: mana joins the total beside Thalia's tax, discards ride
// discard_ids and are paid with the spell on the stack (a discard
// payoff triggers above it), and "tap an untapped creature you control"
// rides teamwork_ids. The view, the bot enumerator and CastSpell all
// read the same per-mode-count arithmetic, so a mode count the caster
// cannot pay is neither offered nor accepted.
//
//	Modes: Escalating(ChooseOneOrMore(…), EscalateMana("{G}")),          // Collective Resistance
//	Modes: Escalating(ChooseOneOrMore(…), EscalateDiscard(1)),           // Collective Brutality
//	Modes: Escalating(ChooseOneOrMore(…), EscalateTapCreature()),        // Collective Effort

// Escalating attaches an escalate cost to a modal spell's choice.
func Escalating(ms *game.ModeSpec, cost *game.EscalateCost) *game.ModeSpec {
	ms.Escalate = cost
	return ms
}

// EscalateMana is "Escalate {cost}".
func EscalateMana(mana string) *game.EscalateCost {
	return &game.EscalateCost{ManaCost: mana, Label: "Escalate " + mana}
}

// EscalateDiscard is "Escalate—Discard a card" (n cards per extra
// mode).
func EscalateDiscard(n int) *game.EscalateCost {
	return &game.EscalateCost{DiscardCards: n, Label: "Escalate—" + discardLabel(n)}
}

// EscalateTapCreature is "Escalate—Tap an untapped creature you
// control."
func EscalateTapCreature() *game.EscalateCost {
	return &game.EscalateCost{TapCreatures: 1, Label: "Escalate—Tap an untapped creature you control"}
}

// checkEscalate holds a ModeSpec.Escalate to the shapes the cast plan
// can pay: mana, discards and creature taps, once each per extra mode.
// It also keeps TapCreatures out of every other slot, since only
// escalate reads it.
func checkEscalate(spec Spec) {
	if ac := spec.AdditionalCost; ac != nil && ac.TapCreatures != 0 {
		panic(fmt.Sprintf("effects.Register: %q puts TapCreatures in the mandatory additional cost — it is an escalate component (#2126)", spec.Name))
	}
	for _, oc := range spec.OptionalCosts {
		if oc.TapCreatures != 0 {
			panic(fmt.Sprintf("effects.Register: %q puts TapCreatures in optional cost %q — it is an escalate component (#2126)", spec.Name, oc.Key))
		}
	}
	for _, a := range spec.Activated {
		if a.Modes != nil && a.Modes.Escalate != nil {
			panic(fmt.Sprintf("effects.Register: %q declares escalate on an activated ability — CR 702.120a is a static ability of a modal spell", spec.Name))
		}
	}
	for _, t := range spec.Triggered {
		if t.Modes != nil && t.Modes.Escalate != nil {
			panic(fmt.Sprintf("effects.Register: %q declares escalate on a trigger — CR 702.120a is a static ability of a modal spell", spec.Name))
		}
	}
	if spec.Modes == nil || spec.Modes.Escalate == nil {
		return
	}
	e, ms := spec.Modes.Escalate, spec.Modes
	switch {
	case ms.Max < 2:
		panic(fmt.Sprintf("effects.Register: %q declares escalate on a spell that can choose only one mode", spec.Name))
	case e.Empty():
		panic(fmt.Sprintf("effects.Register: %q declares an escalate cost that demands nothing", spec.Name))
	case e.DiscardCards < 0 || e.TapCreatures < 0:
		panic(fmt.Sprintf("effects.Register: %q declares a negative escalate component", spec.Name))
	case e.Label == "":
		panic(fmt.Sprintf("effects.Register: %q declares an escalate cost with no label — the picker shows it", spec.Name))
	}
	if e.ManaCost != "" {
		if _, err := game.ParseCost(e.ManaCost); err != nil {
			panic(fmt.Sprintf("effects.Register: %q declares an unparseable escalate cost %q: %v", spec.Name, e.ManaCost, err))
		}
	}
	// Teamwork and escalate's taps share one wire list; no printed card
	// has both and the validator reads a single reading of it.
	for _, oc := range spec.OptionalCosts {
		if e.TapCreatures != 0 && oc.Teamwork != 0 {
			panic(fmt.Sprintf("effects.Register: %q pairs escalate taps with teamwork — they share teamwork_ids", spec.Name))
		}
	}
}
