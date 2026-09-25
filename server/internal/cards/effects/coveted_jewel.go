package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Coveted Jewel — Artifact {6}:
//
//	"When this artifact enters, draw three cards.
//	 {T}: Add three mana of any one color.
//	 Whenever one or more creatures an opponent controls attack you
//	 and aren't blocked, that player draws three cards and gains
//	 control of this artifact. Untap it."
//
// The mana ability is Gilded Lotus's OneColorOfAmount(3). The steal
// trigger needs "attacked and wasn't blocked", and the engine has no
// event for a creature's blocked status once blockers are declared
// (EventAttack fires at the declaration, before blocks exist) — a
// gap GainControl itself does not have, but the trigger condition
// does.
//
// Caveat: the unblocked-attack steal trigger isn't implemented — this
// never changes controller. The ETB draw and the mana ability work.
func init() {
	Register(Spec{
		OracleID:     "98492d7d-3b9e-4ae1-ac45-1b508d6d2670",
		Name:         "Coveted Jewel",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The trigger that gives an unblocked attacker's controller three cards and this artifact isn't implemented — there's no event yet for an attacker that went unblocked. The ETB draw and the mana ability work."},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Coveted Jewel — draw three cards", Do(DrawCards{N: 3})),
		},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: OneColorOfAmount(3),
			Label:    "Add three mana of any one color",
		}},
	})
}
