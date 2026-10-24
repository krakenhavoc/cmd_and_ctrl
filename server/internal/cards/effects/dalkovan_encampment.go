package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dalkovan Encampment — Land (EDHREC rank 3484):
//
//	"This land enters tapped unless you control a Swamp or a
//	 Mountain.
//	 {T}: Add {W}.
//	 {2}{W}, {T}: Whenever you attack this turn, create two 1/1 red
//	 Warrior creature tokens that are tapped and attacking. Sacrifice
//	 them at the beginning of the next end step."
//
// The Mardu land that brings two Warriors to every attack it was
// primed for. The entry is the checkland shape
// (SelfEntersTappedUnless over a Swamp or a Mountain, either
// controlled by the land's controller); the mana ability is a plain
// {W}.
//
// The activated ability creates a delayed triggered ability (CR 603.7):
// "whenever you attack this turn" is DelayedOnEvent{Repeats,
// OncePerBatch} watching EventAttack for the controller's creatures. It
// is not tied to the land, so a land that leaves the battlefield after
// being activated still makes its Warriors; it fires once per attack
// declaration however many creatures were declared (CR 603.2c), and two
// activations are two delayed triggers and four Warriors. The Warriors
// enter tapped and attacking, and the controller chooses what each
// attacks (CR 508.4, CreateTokensAttackingYourChoice, the first
// declared attacker's target asked first); the tokens' IDs go to a
// delayed trigger that sacrifices them at the beginning of the next end
// step. No simplification.
func init() {
	Register(Spec{
		OracleID:     "33a90122-7280-4481-9b97-5879194cae40",
		Name:         "Dalkovan Encampment",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			SelfEntersTappedUnless(youControlLandTyped("Swamp", "Mountain")),
		},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W}",
			Label:    "Add {W}",
		}},
		Activated: []ActivatedAbility{{
			Label:   b33DalkovanEncampmentLabel,
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{2}{W}"), TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return DelayedOnEvent{
					Label:        b33DalkovanAttackLabel,
					On:           []game.EventKind{game.EventAttack},
					Condition:    youAttackCondition,
					Body:         dalkovanWarriorsBody,
					Repeats:      true,
					OncePerBatch: true,
				}.Apply(NewContext(g, item))
			},
		}},
	})
}

var (
	// youAttackCondition is "whenever you attack": an EventAttack for a
	// creature the trigger's controller controls.
	youAttackCondition = game.DelayedCondition("attack/you-attack",
		func(ev game.Event, dt *game.DelayedTrigger, _ *game.Game, _ game.EffectParams) bool {
			return ev.Kind == game.EventAttack && ev.Actor == dt.Controller
		})

	dalkovanWarriorsBody = game.SimpleDelayedBody("dalkovan-encampment/warriors", dalkovanWarriors)
)

// dalkovanWarriors is the delayed trigger's body: two tapped and
// attacking 1/1 red Warriors, sacrificed at the next end step.
func dalkovanWarriors(g *game.Game, item *game.StackItem) error {
	return CreateTokensAttackingYourChoice{
		Template: b33TappedAttackingRedWarrior(),
		N:        2,
		Tapped:   true,
		Prefer:   triggerAttackTarget(item),
		Label:    "Dalkovan Encampment",
		Tail:     game.KeyedTail{Ref: dalkovanWarriorsSacrifice},
	}.Apply(NewContext(g, item))
}

// dalkovanWarriorsSacrifice is the Warriors' tail, keyed (#4294): the end
// step trigger that sacrifices them.
var dalkovanWarriorsSacrifice = AttackingTokensThen("attacking-tokens/dalkovan-encampment-sacrifice-the-warriors",
	func(ctx *Context, created []uuid.UUID, _ game.KeyedTail) error {
		if len(created) == 0 {
			return nil
		}
		return ScheduleDelayedTrigger{
			Label: "Dalkovan Encampment — sacrifice the Warriors",
			Cards: created,
			Body:  sacrificeListedCardsBody,
		}.Apply(ctx)
	})
