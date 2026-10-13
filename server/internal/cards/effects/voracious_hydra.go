package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Voracious Hydra — Creature — Hydra {X}{G}{G}, 0/1 (EDHREC rank
// 2734):
//
//	"Trample
//	 This creature enters with X +1/+1 counters on it.
//	 When this creature enters, choose one —
//	 • Double the number of +1/+1 counters on this creature.
//	 • This creature fights target creature you don't control."
//
// Green's X-cost removal spell that leaves a body behind. Trample
// rides PrintedKeywords; the X counters are the printed CR 614.1c
// entry clause and ride the CR 614 pipeline as one (XCounters,
// #1002), so they are on the permanent before EventETB and the ETB
// trigger below doubles a Hydra that already has them.
//
// The modal ETB is a real "choose one" (TriggeredAbility.Modes, CR
// 603.3c): the controller picks the mode as the trigger goes on the
// stack, and the fight bullet's target is picked then too, so an
// opponent can respond to the choice. The fight bullet is not offered
// when no opponent controls a creature, which leaves the doubling as
// the only mode; the fight is b10Fight (both amounts read before
// either lands) and the doubling goes through AddCounter, so a
// counter doubler applies to it, as printed. A fight target that
// became illegal in response leaves the fight bullet doing nothing
// (CR 608.2b), and nothing doubles.
func init() {
	Register(Spec{
		OracleID:                   "ff8f5a4b-112a-425e-b489-7ee26d1d9fb3",
		Name:                       "Voracious Hydra",
		Completeness:               CompletenessFull,
		PrintedKeywords:            []string{"trample"},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{XCounters(game.CounterPlusOne)},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: b06SelfETB,
			Key:       b25VoraciousHydraLabel,
			Effect:    func(*game.Game, *game.StackItem) error { return nil },
			Modes: ChooseOne(
				ModeDoing("Double the number of +1/+1 counters on this creature.", nil,
					func(item *game.StackItem, ctx *Context, _ int) error {
						return b23DoublePlusOneCountersOn(ctx, item.SourceCardID)
					}),
				ModeDoing("This creature fights target creature you don't control.",
					TargetCreature("target creature you don't control", OpponentControls()),
					func(item *game.StackItem, ctx *Context, occ int) error {
						t, ok := ModeTarget(ctx, occ)
						if !ok {
							return nil
						}
						return b10Fight(ctx, item.SourceCardID, t.ID)
					}),
			),
		}},
	})
}
