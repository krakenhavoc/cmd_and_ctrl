package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Waterbender's Restoration — "As an additional cost to cast this
// spell, waterbend {X}. (While paying a waterbend cost, you can tap
// your artifacts and creatures to help. Each one pays for {1}.)
// Exile X target creatures you control. Return those cards to the
// battlefield under their owner's control at the beginning of the
// next end step."
//
// The canonical delayed blink: the creatures leave now and come back
// a step boundary later as new objects, which is what makes the card
// both an ETB engine and a fog against targeted removal.
//
// S22 sandbox simplification — **waterbend {X} is not costed.**
// There is no "tap permanents as a cost" component (`AbilityCost`
// has tap-this, sacrifice, mana and life; convoke and waterbend are
// the same missing piece, tracked together in the Aang triage), and
// X here is defined by that cost rather than by an {X} in the mana
// cost, so the engine has nothing to charge. The target clause is
// therefore "any number of target creatures you control" and the
// caster is expected to tap their own artifacts and creatures by
// hand — the same posture the permissive-by-default mana gate takes
// for every other cost in the sandbox. Left unpoliced this is
// strictly stronger than printed, which is why it is called out
// here rather than buried.
func init() {
	Register(Spec{
		OracleID: "285046f6-b3c4-4eb7-8712-9dffebabc762",
		Name:     "Waterbender's Restoration",
		Targets:  TargetCreature("X target creatures you control", YouControl()).WithCount(1, 0),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			exiled, err := exileTargetsForDelayedReturn(ctx)
			if err != nil {
				return err
			}
			if len(exiled) == 0 {
				return nil
			}
			return ScheduleDelayedTrigger{
				At:     game.StepEnd,
				Label:  "Waterbender's Restoration — return the exiled creatures",
				Cards:  exiled,
				Effect: returnExiledCardsToOwners,
			}.Apply(ctx)
		},
	})
}
