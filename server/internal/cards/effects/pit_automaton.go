package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Pit Automaton — Artifact Creature — Construct {2}, 0/4 (Aetherdrift,
// #4464):
//
//	"Defender
//	 {T}: Add {C}{C}. Spend this mana only to activate abilities.
//	 {2}, {T}: When you next activate an exhaust ability that isn't a
//	 mana ability this turn, copy it. You may choose new targets for the
//	 copy."
//
// The mana is Boommobile's activate-only restriction (ManaRestrictActivate):
// it pays for any activated ability, a mana ability with a mana cost
// included, and refuses a spell.
//
// The second ability sets up Repeated Reverberation's CR 603.7 delayed
// trigger (DelayedOnEvent) watching your activations. It fires on the
// first exhaust ability you activate this turn, whatever permanent it is
// on, and is then used up; it ends with the turn if none comes. A mana
// ability never goes on the stack and is announced as a different event,
// so it neither fires nor uses up the trigger. The copy is made while the
// exhaust ability is still on the stack, keeps its X, and may get new
// targets; a copy is not activated, so it neither fires another Automaton
// nor spends the exhaust. An exhaust ability that has already left the
// stack when the trigger resolves is not copied.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "49bfecc3-6258-454e-a8fd-fc49098a0898",
		Name:            "Pit Automaton",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
		ManaAbilities: []ManaAbility{{
			Cost:         ManaAbilityCost{Tap: true},
			Produced:     "{C}{C}",
			Label:        "{T}: Add {C}{C}. Spend this mana only to activate abilities.",
			Restrictions: []string{game.ManaRestrictActivate},
		}},
		Activated: []ActivatedAbility{{
			Label:   "{2}, {T}: When you next activate an exhaust ability that isn't a mana ability this turn, copy it. You may choose new targets for the copy.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{2}"), TapCost()),
			Effect: Do(Instruction(func(ctx *Context) error {
				return DelayedOnEvent{
					Label:     "Pit Automaton — copy that exhaust ability",
					On:        []game.EventKind{game.EventActivateAbility},
					Condition: pitAutomatonCondition,
					Body:      pitAutomatonBody,
				}.Apply(ctx)
			})),
		}},
	})
}

var (
	// "When you next activate an exhaust ability that isn't a mana
	// ability this turn". Mana abilities are announced as
	// EventManaAbilityActivated, which this does not watch. On-disk
	// identity.
	pitAutomatonCondition = game.DelayedCondition("pit-automaton/next-exhaust-activation",
		func(ev game.Event, dt *game.DelayedTrigger, _ *game.Game, _ game.EffectParams) bool {
			return ev.Kind == game.EventActivateAbility && ev.Actor == dt.Controller && ev.Exhaust
		})

	// "Copy it. You may choose new targets for the copy." On-disk
	// identity.
	pitAutomatonBody = game.SimpleDelayedBody("pit-automaton/copy-exhaust", pitAutomatonCopy)
)

func pitAutomatonCopy(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ev := ctx.Trigger().Event
	if ev.StackItemID == uuid.Nil {
		return nil
	}
	return CopyAbility{
		ItemID:           ev.StackItemID,
		Controller:       item.Controller,
		Count:            1,
		ChooseNewTargets: true,
	}.Apply(ctx)
}
