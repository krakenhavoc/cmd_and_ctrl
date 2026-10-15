package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Slicer, Hired Muscle // Slicer, High-Speed Antagonist (#3275, #3195,
// ADR 0152) — a More Than Meets the Eye card.
//
// Front face, Legendary Artifact Creature — Robot {4}{R}, 3/4:
//
//	"More Than Meets the Eye {2}{R}
//	 Double strike, haste
//	 At the beginning of each opponent's upkeep, you may have that player
//	 gain control of Slicer until end of turn. If you do, untap Slicer,
//	 goad it, and it can't be sacrificed this turn. If you don't, convert
//	 it."
//
// Back face, Legendary Artifact — Vehicle, 3/2:
//
//	"Living metal
//	 First strike, haste
//	 Whenever Slicer deals combat damage to a player, convert it at end
//	 of combat."
//
// "That player" is the upkeep's player, read off the trigger's event.
// The control change lasts until cleanup; the goad is the trigger
// controller's, so it lasts until their next turn (CR 701.15a) and
// Slicer must attack someone other than them. "Can't be sacrificed this
// turn" is CantBeSacrificedThisTurn (#3275), so the borrowing player
// can't feed it to a sacrifice cost or an edict before it goes home.
// "If you do" is the yes branch: with Slicer gone by then, neither the
// control change nor the convert finds it (CR 400.7).
//
// The back face's convert is a delayed trigger at end of combat with
// the turn-over count recorded when it is created (CR 701.28e), so two
// first-strike and regular hits turn it over once (ADR 0152 §3).
//
// No simplification.
const slicerOracleID = "fa716557-8e75-4617-8310-ab23a9ca1af1"

func init() {
	Register(Spec{
		OracleID:         slicerOracleID,
		Name:             "Slicer, Hired Muscle",
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{"double strike", "haste"},
		AlternativeCosts: []game.AlternativeCost{MoreThanMeetsTheEye("{2}{R}")},
		Triggered: []game.TriggeredAbility{
			AtEachOpponentsUpkeep("Slicer, Hired Muscle — you may have that player gain control of Slicer; if you don't, convert it",
				slicerUpkeep),
		},
	})
	Register(Spec{
		OracleID:        slicerOracleID + "#1",
		Name:            "Slicer, High-Speed Antagonist",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike", "haste"},
		Static:          []game.StaticAbility{LivingMetal()},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Slicer, High-Speed Antagonist — convert it at end of combat",
				func(g *game.Game, item *game.StackItem) error {
					return ScheduleDelayedTrigger{
						At:    game.StepEndCombat,
						Label: "Slicer, High-Speed Antagonist — convert Slicer at end of combat",
						Body:  transformersConvertThisBody,
					}.Apply(NewContext(g, item))
				}),
		},
	})
}

func slicerUpkeep(g *game.Game, item *game.StackItem) error {
	player := item.Trigger.Event.Actor
	slicer := item.SourceCardID
	return MayChoice{
		Question: "Slicer, Hired Muscle — have " + attackTargetName(g, player) + " gain control of Slicer until end of turn? (If you don't, convert it.)",
		YesLabel: "Hand it over",
		NoLabel:  "Convert it",
		OnYes: func(ctx *Context) error {
			if ctx.isNewSourceObject(slicer) || !onBattlefield(ctx.Game, slicer) {
				return nil
			}
			if err := (GainControl{
				Target:     slicer,
				Controller: player,
				Label:      "Slicer, Hired Muscle — gain control until end of turn",
			}).Apply(ctx); err != nil {
				return err
			}
			if err := (UntapTarget{Target: slicer}).Apply(ctx); err != nil {
				return err
			}
			if err := goadAndScheduleClear(ctx, []uuid.UUID{slicer}); err != nil {
				return err
			}
			return CantBeSacrificedThisTurn{
				Target: slicer,
				Label:  "Slicer, Hired Muscle — can't be sacrificed this turn",
			}.Apply(ctx)
		},
		OnNo: func(ctx *Context) error { return ConvertThis{}.Apply(ctx) },
	}.Apply(NewContext(g, item))
}
