package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Turnabout — Instant {2}{U}{U} (Ravnica Remastered):
//
//	"Choose artifact, creature, or land. Tap all untapped permanents of the
//	 chosen type target player controls, or untap all tapped permanents of
//	 that type that player controls."
//
// Both choices are made on resolution by the caster, so they are one prompt
// of the six printed outcomes (type × tap or untap), the first of which
// always works. The target player's permanents are read after the choice, so
// a player who has left leaves nothing to do (CR 608.2b); the chosen set is
// fixed when the answer lands.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "715d0777-4d36-463b-9102-6bb6bb00b7e9",
		Name:         "Turnabout",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target player"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return PickOption{
				Question: "Turnabout — choose artifact, creature, or land, and tap or untap them",
				Options:  turnaboutOptions,
				ThenKey:  turnaboutChosen,
				WithItem: true,
			}.Apply(ctx)
		},
	})
}

var turnaboutOptions = []game.ChoiceOption{
	{Label: "Tap all untapped artifacts the player controls"},
	{Label: "Untap all tapped artifacts the player controls"},
	{Label: "Tap all untapped creatures the player controls"},
	{Label: "Untap all tapped creatures the player controls"},
	{Label: "Tap all untapped lands the player controls"},
	{Label: "Untap all tapped lands the player controls"},
}

var turnaboutChosen = OptionPickThen("option-pick/turnabout-type-and-direction", func(ctx *Context, r game.OptionPicked) error {
	if r.Option == nil || r.Index < 0 || r.Index >= len(turnaboutOptions) {
		return nil
	}
	var who game.TargetRef
	found := false
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetPlayer {
			who, found = t, true
			break
		}
	}
	if !found {
		return nil
	}
	typeMatch := []CardPredicate{Artifact(), Creature(), Land()}[r.Index/2]
	tap := r.Index%2 == 0
	for _, c := range MatchingBattlefield(ctx, And(typeMatch, func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
		return c.Controller == who.ID && c.Tapped != tap
	})) {
		var err error
		if tap {
			err = TapTarget{Target: c.InstanceID}.Apply(ctx)
		} else {
			err = UntapTarget{Target: c.InstanceID}.Apply(ctx)
		}
		if err != nil {
			return err
		}
	}
	return nil
})
