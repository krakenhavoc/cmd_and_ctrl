package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// arenaLabel is the row Arena and Magus of the Arena share.
const arenaLabel = "{3}, {T}: Tap target creature you control and target creature of an opponent's choice they control. Those creatures fight each other."

// Arena — Land:
//
//	"{3}, {T}: Tap target creature you control and target creature of
//	 an opponent's choice they control. Those creatures fight each
//	 other."
//
// The second target is an opponent's choice (#2055, ADR 0019's
// 2026-10-16 amendment): as the ability is activated its controller
// chooses their own creature first, then an opponent, and that opponent
// chooses a creature they control (the Magus of the Arena rulings). A
// new target chosen for the second clause must be a creature the chosen
// opponent controls; the choice of opponent can't be changed.
//
// The rulings on resolution: if either target is illegal, the ability
// taps the one that is left and no damage is dealt (CR 701.14b); with
// both illegal it doesn't resolve (CR 608.2b). Tapped creatures can be
// targeted, and the fight reads their power as the ability resolves.
// In a multiplayer game a different opponent may be chosen each time.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "12326d23-1fb5-4084-a23a-d6e2c90ea36c",
		Name:         "Arena",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{arenaRow()},
	})
}

// arenaRow is the ability Arena and Magus of the Arena both print.
func arenaRow() ActivatedAbility {
	return ActivatedAbility{
		Label: arenaLabel,
		Cost:  Plus(ManaCost("{3}"), TapCost()),
		Targets: Clauses(
			TargetCreature("target creature you control", YouControl()),
			TheyControl(OfAnOpponentsChoice(TargetCreature("target creature of an opponent's choice they control"))),
		),
		Effect: arenaTapAndFight,
	}
}

// arenaTapAndFight taps each still-legal target and, when both are,
// has them fight (CR 701.14a); a creature that is an illegal target by
// then deals and is dealt nothing (CR 701.14b).
func arenaTapAndFight(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var fighters []uuid.UUID
	for slot := 0; slot < 2; slot++ {
		t, ok := ctx.ClauseTarget(slot)
		if !ok || t.Kind != game.TargetCard {
			continue
		}
		if err := (TapTarget{Target: t.ID}).Apply(ctx); err != nil {
			return err
		}
		fighters = append(fighters, t.ID)
	}
	if len(fighters) != 2 {
		return nil
	}
	return b10Fight(ctx, fighters[0], fighters[1])
}
