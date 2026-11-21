package cardshu

import (
	"sort"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ichormoon Gauntlet — Artifact {2}{U}:
//
//	"Planeswalkers you control have "[0]: Proliferate" and "[−12]: Take
//	 an extra turn after this one."
//	 Whenever you cast a noncreature spell, choose a counter on target
//	 permanent. Put an additional counter of that kind on that
//	 permanent."
//
// The two quoted abilities are two ADR 0093 bundles granted to the
// class with GrantAbilitiesToYourPlaneswalkers (Kiora of Salt and
// Sand's shape, #2797). Each row is the walker's: it pays the loyalty
// (CR 606.6) and shares the walker's one loyalty activation a turn with
// its printed rows (CR 606.3). The recipient set is read live.
//
// The trigger targets any permanent (one with no counters is a legal
// target, and then nothing happens). At resolution the kinds of counter
// on it are read: with one kind the extra counter is put at once, with
// several its controller is asked which (CR 608.2), and the counter is
// put by the trigger's controller so placement replacements apply as
// usual.
//
// No simplification. Guffster's Party deck request (#4184).
func init() {
	const name = "Ichormoon Gauntlet"
	effects.Register(effects.Spec{
		OracleID:     "9fc7aba7-eda2-4134-9c10-31ee35201b72",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{
			{
				Key: ichormoonProliferateGrant,
				Activated: []effects.ActivatedAbility{{
					Label: "0: Proliferate.",
					Cost:  effects.LoyaltyCost(0),
					Effect: func(g *game.Game, item *game.StackItem) error {
						return effects.Proliferate{}.Apply(effects.NewContext(g, item))
					},
				}},
				Text: "[0]: Proliferate.",
			},
			{
				Key: ichormoonExtraTurnGrant,
				Activated: []effects.ActivatedAbility{{
					Label: "−12: Take an extra turn after this one.",
					Cost:  effects.LoyaltyCost(-12),
					Effect: func(g *game.Game, item *game.StackItem) error {
						return effects.TakeExtraTurn{}.Apply(effects.NewContext(g, item))
					},
				}},
				Text: "[−12]: Take an extra turn after this one.",
			},
		},
		Static: []game.StaticAbility{effects.GrantAbilitiesToYourPlaneswalkers(ichormoonProliferateGrant, ichormoonExtraTurnGrant)},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WheneverYouCast(effects.Noncreature(), name+" — an additional counter of a kind already on target permanent", ichormoonAddCounter),
				effects.TargetPermanent("target permanent")),
		},
	})
}

const (
	ichormoonProliferateGrant = "ichormoon-gauntlet/proliferate"
	ichormoonExtraTurnGrant   = "ichormoon-gauntlet/extra-turn"
)

// ichormoonAddCounter chooses a kind of counter on the target and puts
// one more of it there.
func ichormoonAddCounter(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	t, ok := effects.FirstLegalTarget(ctx)
	if !ok {
		return nil
	}
	c, ok := g.LookupCardForEffect(t.ID)
	if !ok {
		return nil
	}
	var kinds []string
	for kind, n := range c.Counters {
		if n > 0 {
			kinds = append(kinds, kind)
		}
	}
	switch len(kinds) {
	case 0:
		return nil
	case 1:
		return g.AddCounterByForEffect(item.Controller, t.ID, kinds[0], 1)
	}
	sort.Strings(kinds)
	return effects.PutChosenCounterOn(ctx, item.Controller, t.ID,
		"Ichormoon Gauntlet — choose a counter on that permanent", kinds)
}
