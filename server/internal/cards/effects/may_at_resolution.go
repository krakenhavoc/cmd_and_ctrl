package effects

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// may_at_resolution.go — an activated ability whose whole effect is "you
// may …", asked as it resolves (ADR 0197 §10, #2731).
//
//	Activated: []ActivatedAbility{
//	    YouMay(Boast("{1}: You may have Arni's base power become …",
//	        ManaCost("{1}"), arniSetBasePower),
//	        MayAtResolution{Question: "…?", Helps: arniSetWouldRaise}),
//	}
//
// CR 608.2d: the choice in an ability's effect is made as it resolves. An
// optional trigger already asks its "you may" then (ADR 0197): the item
// carries the question, the resolution asks a trigger_prompt, "yes" runs
// the effect, and an ADR 0127 standing answer can answer it. An
// activated row now declares the same question and takes the same path,
// with its own standing-answer key (`<catalog key>|activated|<label>`).
//
// This is for a "you may" that IS the whole effect. One partway through
// ("Mill two cards. Then you may sacrifice this land") is a MayChoice
// inside the effect, asked after the part before it has happened.

// MayAtResolution is an activated ability's resolution-time "you may".
type MayAtResolution struct {
	// Question is the prompt's header, written the way the card is:
	// "Arni Brokenbrow — have its base power become …?".
	Question string
	// Helps, when set, reports as the ability resolves whether "yes"
	// would leave its controller better off on this board. When it
	// answers false the prompt warns the player and the bot declines;
	// the answer is still the player's. Nil means the row has no
	// opinion, which is right for nearly every "you may".
	Helps func(ctx *Context) bool
}

// question is the declared question, or "" for a nil declaration.
func (m *MayAtResolution) question() string {
	if m == nil {
		return ""
	}
	return m.Question
}

// YouMay declares ab's whole effect a resolution-time "you may".
func YouMay(ab ActivatedAbility, m MayAtResolution) ActivatedAbility {
	ab.May = &m
	return ab
}

// checkMayAtResolution refuses a declaration the engine would not ask:
// one with no question, or on an ability with no effect to run on a
// "yes". Called by Register.
func checkMayAtResolution(name string, abilities []ActivatedAbility) {
	for i, a := range abilities {
		if a.May == nil {
			continue
		}
		if a.May.Question == "" || a.Effect == nil {
			panic(fmt.Sprintf("effects.Register(%s): activated ability %d declares a \"you may\" with no question or no effect", name, i))
		}
	}
}

// rowMayHelps is game.CatalogActivatedMayHelps: the Helps of the
// catalog row a stamped activation names, or nil. Only a card's own row
// is looked up; a lent or granted row declares nothing here.
func rowMayHelps(ref game.AbilityRef) func(*game.Game, *game.StackItem) bool {
	spec, ok := Lookup(ref.Key)
	if !ok {
		return nil
	}
	for _, a := range spec.Activated {
		if a.Label != ref.Name || a.May == nil || a.May.Helps == nil {
			continue
		}
		helps := a.May.Helps
		return func(g *game.Game, item *game.StackItem) bool {
			return helps(NewContext(g, item))
		}
	}
	return nil
}

func init() { game.CatalogActivatedMayHelps = rowMayHelps }
