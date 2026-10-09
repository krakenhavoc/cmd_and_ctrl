package effects

import (
	"errors"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Presumed Dead — Instant {1}{B}:
//
//	"Until end of turn, target creature gets +2/+0 and gains "When this
//	 creature dies, return it to the battlefield under its owner's
//	 control and suspect it." (A suspected creature has menace and
//	 can't block.)"
//
// Fake Your Own Death's shape (ADR 0093 PR 4): the boost and the grant
// are one effect at one timestamp, and the granted trigger belongs to
// the creature. It returns the card UNTAPPED, as printed, and suspects
// the new object it became (CR 400.7). A token that dies is gone and
// returns nothing, and so is not suspected.
//
// One declared simplification, weaker than printed: if the creature
// asks a question as it re-enters (a copy effect, a shockland's
// payment) it comes back but is not suspected.
const presumedDeadReturn = "presumed-dead/return"

func init() {
	Register(Spec{
		OracleID:     "a2d3e3f0-7399-4eb9-a0cf-831710c3d431",
		Name:         "Presumed Dead",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"If the creature asks a question as it comes back, such as a copy effect, it returns but isn't suspected."},
		Targets:      TargetCreature("target creature"),
		Grants: []AbilityGrant{{
			Key: presumedDeadReturn,
			Triggered: []game.TriggeredAbility{
				WhenThisDies("Presumed Dead — return it and suspect it", presumedDeadReturnsAndSuspects),
			},
			Text: "When this creature dies, return it to the battlefield under its owner's control and suspect it.",
		}},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			target, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return GrantAbilitiesFor{
				Target: target,
				Keys:   []string{presumedDeadReturn},
				Also:   []game.Mod{game.ModifyPTMod(2, 0)},
				Label:  "Presumed Dead — +2/+0 and it returns when it dies",
			}.Apply(ctx)
		},
	})
}

// presumedDeadReturnsAndSuspects is the granted trigger's body:
// returnThisCreatureFromGraveyard's "still the object that died" check
// and untapped return, then the suspect on whatever entered.
func presumedDeadReturnsAndSuspects(g *game.Game, item *game.StackItem) error {
	z := g.FindCardZoneForEffect(item.SourceCardID)
	if z == nil || z.Kind != game.ZoneGraveyard || !sameGraveyardObject(z, item) {
		return nil
	}
	entered, err := g.ReturnFromGraveyardWithCountersForEffect(item.SourceCardID, uuid.Nil, false, nil)
	if err != nil {
		if errors.Is(err, game.ErrCardNotFound) {
			return nil
		}
		return err
	}
	if entered == uuid.Nil {
		return nil
	}
	// The game mutator, not the Suspect primitive: this ability's source
	// IS the card that came back, a new object since it died, and the
	// primitive's #1432 guard would read that as "the old object's
	// ability must not reach the new one" and do nothing. Here the new
	// object is exactly what the instruction names.
	g.SuspectForEffect(entered)
	return nil
}
