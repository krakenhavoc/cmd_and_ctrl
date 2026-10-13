package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fandaniel, Telophoroi Ascian — Legendary Creature — Elder Wizard
// {4}{B}, 4/5:
//
//	"Whenever you cast an instant or sorcery spell, surveil 1.
//	 At the beginning of your end step, each opponent may sacrifice a
//	 nontoken creature of their choice. Each opponent who doesn't
//	 loses 2 life for each instant and sorcery card in your graveyard."
//
// The end-step half is Braids, Arisen Nightmare's chain: opponents are
// asked one at a time in turn order, each by a two-way option prompt
// (sacrifice, or take the life loss), and an opponent with no nontoken
// creature is never asked and simply loses the life. The graveyard is
// counted when each opponent's consequence applies; nothing in this
// ability changes the controller's graveyard, so the count is the same
// for everyone.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3909ff79-2103-407c-a812-a5a5a996fce2",
		Name:         "Fandaniel, Telophoroi Ascian",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Or(Instant(), Sorcery()), "Fandaniel — surveil 1", Do(Surveil{N: 1})),
			AtYourEndStep("Fandaniel — each opponent sacrifices a nontoken creature or loses life", func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return fandanielOpponentsStep(ctx, ctx.Controller(), ctx.Opponents())
			}),
		},
	})
}

// fandanielInstantsAndSorceries counts the instant and sorcery cards in
// `controller`'s graveyard.
func fandanielInstantsAndSorceries(g *game.Game, controller uuid.UUID) int {
	p := g.PlayerByIDForEffect(controller)
	if p == nil {
		return 0
	}
	n := 0
	for _, c := range p.Graveyard.Cards {
		if c.IsInstant() || c.IsSorcery() {
			n++
		}
	}
	return n
}

// fandanielNontokenCreatures lists the nontoken creatures `opponent`
// controls.
func fandanielNontokenCreatures(g *game.Game, opponent uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller == opponent && c.IsCreature() && !c.IsToken() {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

func fandanielOpponentsStep(ctx *Context, controller uuid.UUID, remaining []uuid.UUID) error {
	for len(remaining) > 0 {
		victim := remaining[0]
		rest := remaining[1:]
		p := ctx.Game.PlayerByIDForEffect(victim)
		if p == nil || p.Eliminated {
			remaining = rest
			continue
		}
		loss := 2 * fandanielInstantsAndSorceries(ctx.Game, controller)
		candidates := fandanielNontokenCreatures(ctx.Game, victim)
		if len(candidates) == 0 {
			if loss > 0 {
				if err := ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), victim, -loss); err != nil {
					return err
				}
			}
			remaining = rest
			continue
		}
		return PickOption{
			Player:   victim,
			Question: "Fandaniel, Telophoroi Ascian — sacrifice a nontoken creature, or lose 2 life for each instant and sorcery card in its controller's graveyard?",
			Options: []game.ChoiceOption{
				{Label: "Don't sacrifice — lose life", LifeCost: loss},
				{Label: "Sacrifice a nontoken creature"},
			},
			Then: func(ctx *Context, index int) error {
				if index == 1 {
					return SacrificeChoice{
						Player:     victim,
						Candidates: candidates,
						Question:   "Fandaniel, Telophoroi Ascian — sacrifice a nontoken creature",
						Then: func(ctx *Context) error {
							return fandanielOpponentsStep(ctx, controller, rest)
						},
					}.Apply(ctx)
				}
				if l := 2 * fandanielInstantsAndSorceries(ctx.Game, controller); l > 0 {
					if err := ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), victim, -l); err != nil {
						return err
					}
				}
				return fandanielOpponentsStep(ctx, controller, rest)
			},
		}.Apply(ctx)
	}
	return nil
}
