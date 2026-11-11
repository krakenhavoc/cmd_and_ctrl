package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Havoc Eater — Creature — Elemental {5}{R}{R}, 3/3:
//
//	"Flying
//	 When this creature enters, for each opponent, goad up to one target
//	 creature that opponent controls. Put X +1/+1 counters on this
//	 creature, where X is the total power of creatures goaded this way."
//
// "For each opponent, up to one target creature that player controls" is
// perOpponentClauses. Each pick still legal is goaded by the controller
// (GoadTarget's shape, goadAndScheduleClear: until the controller's next
// turn, CR 701.15a). X is the total power of those creatures read after
// the goad, as they stand (CR 208.1, 613.4), so a creature that left in
// response, or went illegal, adds nothing and a pick of negative power
// counts as it is (X never goes below zero). The counters go on Havoc
// Eater if it is still the same object on the battlefield.
//
// No simplification.
func init() {
	trigger := WhenThisEnters("Havoc Eater — for each opponent, goad up to one target creature that player controls, then put X +1/+1 counters on this creature",
		havocEaterEnters)
	trigger.Targets = perOpponentClauses(game.MaxPlayers-1, "creature", Creature())
	Register(Spec{
		OracleID:        "769c511c-37ce-4a9d-85f8-1ba91c2daf82",
		Name:            "Havoc Eater",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{trigger},
	})
}

func havocEaterEnters(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var goaded []uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard && onBattlefield(g, t.ID) {
			goaded = append(goaded, t.ID)
		}
	}
	if len(goaded) == 0 {
		return nil
	}
	if err := goadAndScheduleClear(ctx, goaded); err != nil {
		return err
	}
	total := 0
	for _, id := range goaded {
		if c, ok := g.LookupCardForEffect(id); ok {
			total += c.PowerForComparison()
		}
	}
	src := item.SourceCardID
	if total <= 0 || ctx.isNewSourceObject(src) || !onBattlefield(g, src) {
		return nil
	}
	return AddCounter{Target: src, Kind: game.CounterPlusOne, N: total}.Apply(ctx)
}
