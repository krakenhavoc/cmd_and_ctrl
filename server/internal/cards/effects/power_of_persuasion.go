package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Power of Persuasion — Sorcery {2}{U} (Adventures in the Forgotten Realms,
// S82):
//
//	"Choose target creature an opponent controls, then roll a d20.
//	 1—9 | Return it to its owner's hand.
//	 10—19 | Its owner puts it on their choice of the top or bottom of
//	 their library.
//	 20 | Gain control of it until the end of your next turn."
//
// With the target gone or illegal as the spell resolves it does nothing and
// rolls nothing (CR 608.2b). The creature is carried to the tail as data, and
// a creature that has left the battlefield by then is left alone. Row 2 is
// the owner's own choice of lane (PutInLibraryInAnyOrder, as Endless Detour);
// row 3 is a layer-2 control change pinned to the creature, lasting through
// the end of the caster's next turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5c615c99-77b7-4bea-9286-1e8d257153fb",
		Name:         "Power of Persuasion",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature an opponent controls", OpponentControls()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				return RollDice{Sides: 20, N: 1, Then: powerOfPersuasionRolled, TailCarry: []uuid.UUID{t.ID}}.Apply(ctx)
			}
			return nil
		},
	})
}

// powerOfPersuasionRolled is the results table, after the d20. An on-disk key.
var powerOfPersuasionRolled = CardTail("dice/power-of-persuasion-rolled", func(ctx *Context, o game.TailOutcome) error {
	if len(o.Rolls) == 0 || len(o.Tail.Carry) == 0 {
		return nil
	}
	g, id := ctx.Game, o.Tail.Carry[0]
	if !onBattlefield(g, id) {
		return nil
	}
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		return nil
	}
	switch r := o.Rolls[0]; {
	case r <= 9:
		return BounceToHand{Target: id}.Apply(ctx)
	case r <= 19:
		return PutInLibraryInAnyOrder{
			Chooser:   c.Owner,
			Cards:     []uuid.UUID{id},
			Placement: game.LibraryPlaceTopOrBottom,
			Label:     "Power of Persuasion — put it on the top or bottom of your library",
		}.Apply(ctx)
	}
	return GainControl{
		Target:   id,
		Duration: g.UntilEndOfYourNextTurnDuration(ctx.Controller()),
		Label:    "Power of Persuasion — gain control until the end of your next turn",
	}.Apply(ctx)
})
