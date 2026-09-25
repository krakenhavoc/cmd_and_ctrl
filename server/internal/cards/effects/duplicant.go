package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Duplicant — Artifact Creature — Shapeshifter {6}, 2/4 (EDHREC rank
// 2911):
//
//	"Imprint — When this creature enters, you may exile target
//	 nontoken creature.
//	 As long as a card exiled with this creature is a creature card,
//	 this creature has the power, toughness, and creature types of
//	 the last creature card exiled with it. It's still a
//	 Shapeshifter."
//
// Colourless exile removal that wears what it ate. The imprint is an
// optional targeted ETB (Springbloom Druid's prompt-then-pick
// shape): the target is chosen when the trigger goes on the stack,
// re-checked at resolution (CR 608.2b), and exiled by ExileTarget.
// "Exiled with this creature" is not a field on any card, so it is
// read back off the event log: b27ExiledWith finds every card whose
// latest move into exile happened while Duplicant's imprint trigger
// was resolving, and only while that card is still in exile. Two
// statics then read the LAST such card that is a creature card —
// layer 4 replaces Duplicant's creature types with that card's plus
// Shapeshifter, layer 7b sets its power and toughness to that
// card's printed values (not a CDA: the ability is conditional, CR
// 604.3a, so counters and anthems still apply on top). With nothing
// exiled — or only a noncreature, which the target clause forbids
// anyway — Duplicant stays a 2/4 Shapeshifter. A Duplicant that
// leaves and returns is a new object with an empty record, as
// printed.
//
// Sandbox simplification, declared: the statics re-read the record
// on every layer recompute, and a recompute follows every
// battlefield change — but not a card leaving EXILE by itself, so a
// Duplicant whose imprint is pulled back out of exile keeps the
// borrowed body until the next thing enters or leaves the
// battlefield. A beat late, never stronger.
func init() {
	Register(Spec{
		OracleID:     "ea86abfa-6cab-4ef0-8463-34136fc25b59",
		Name:         "Duplicant",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"If the exiled card leaves exile, Duplicant keeps the copied power, toughness and types until the next permanent enters or leaves the battlefield."},
		Static:       b27DuplicantStatics(),
		Triggered: []game.TriggeredAbility{{
			Watches:        []game.EventKind{game.EventETB},
			AppliesTo:      b06SelfETB,
			OptionalPrompt: &game.TriggerOptionalPrompt{Question: "Duplicant — exile target nontoken creature?"},
			Targets:        TargetCreature("target nontoken creature", Not(IsTokenPredicate())),
			Key:            b27DuplicantLabel,
			Effect:         b27ExileChosenTarget,
		}},
	})
}
