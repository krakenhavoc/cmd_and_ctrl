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

// fandanielLoseLife is "loses 2 life for each instant and sorcery card in
// its controller's graveyard", counted when it applies.
func fandanielLoseLife(ctx *Context, controller, victim uuid.UUID) error {
	if loss := 2 * fandanielInstantsAndSorceries(ctx.Game, controller); loss > 0 {
		return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), victim, -loss)
	}
	return nil
}

// fandanielOpponentsStep asks the first opponent of `remaining` and hands
// the rest on. Carry is the opponents still to be asked, in turn order.
func fandanielOpponentsStep(ctx *Context, controller uuid.UUID, remaining []uuid.UUID) error {
	for len(remaining) > 0 {
		victim := remaining[0]
		rest := remaining[1:]
		p := ctx.Game.PlayerByIDForEffect(victim)
		if p == nil || p.Eliminated {
			remaining = rest
			continue
		}
		candidates := fandanielNontokenCreatures(ctx.Game, victim)
		if len(candidates) == 0 {
			if err := fandanielLoseLife(ctx, controller, victim); err != nil {
				return err
			}
			remaining = rest
			continue
		}
		return PickOption{
			Player:   victim,
			Question: "Fandaniel, Telophoroi Ascian — sacrifice a nontoken creature, or lose 2 life for each instant and sorcery card in its controller's graveyard?",
			Options: []game.ChoiceOption{
				{Label: "Don't sacrifice — lose life", LifeCost: 2 * fandanielInstantsAndSorceries(ctx.Game, controller)},
				{Label: "Sacrifice a nontoken creature"},
			},
			ThenKey: fandanielChosen,
			Carry:   rest,
		}.Apply(ctx)
	}
	return nil
}

// fandanielChosen applies one opponent's answer and asks the next, keyed
// (#4253): Carry is the opponents still to be asked. The sacrifice they
// choose is SacrificeChoice, whose continuation is the keyed tail
// fandanielSacrificedNext.
var fandanielChosen game.OptionPickThen

// Registered in init: the body asks the next opponent, which names the key
// again, and a package-level initialiser would be an initialisation cycle.
func init() {
	fandanielChosen = OptionPickThen("option-pick/fandaniel-telophoroi-ascian-sacrifice-or-lose-life", fandanielAnswered)
}

func fandanielAnswered(ctx *Context, r game.OptionPicked) error {
	victim, rest := r.Chooser, append([]uuid.UUID(nil), r.Carry...)
	controller := ctx.Controller()
	if r.Index == 1 {
		return SacrificeChoice{
			Player:     victim,
			Candidates: fandanielNontokenCreatures(ctx.Game, victim),
			Question:   "Fandaniel, Telophoroi Ascian — sacrifice a nontoken creature",
			ThenTail:   withCarry(ctx.Tail(fandanielSacrificedNext), rest...),
		}.Apply(ctx)
	}
	if err := fandanielLoseLife(ctx, controller, victim); err != nil {
		return err
	}
	return fandanielOpponentsStep(ctx, controller, rest)
}

// fandanielSacrificedNext asks the next opponent once the sacrifice has
// gone, keyed (#4253): Carry is the opponents still to be asked, and the
// tail's Player is Fandaniel's controller. An on-disk key: never renamed,
// never reused.
var fandanielSacrificedNext = CardTail("then/fandaniel-telophoroi-ascian-next-opponent", func(ctx *Context, o game.TailOutcome) error {
	return fandanielOpponentsStep(ctx, ctx.Controller(), append([]uuid.UUID(nil), o.Tail.Carry...))
})
