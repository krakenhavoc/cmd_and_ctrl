package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Alicia Masters, Skilled Sculptor — Legendary Creature — Human Artificer
// {1}{R}, 0/4 (Marvel Super Heroes Commander, #4455):
//
//	"At the beginning of combat on your turn, if you've cast a noncreature
//	 spell this turn, create a Treasure token.
//	 Sense the Good — At the beginning of your end step, each player gains
//	 control of all creatures they own."
//
// The combat trigger is Iron Man, Master of Machines' intervening if (CR 603.4)
// over the turn's cast tally, checked as the step begins and again as the
// trigger resolves. Sense the Good registers, for every creature on the
// battlefield that someone else controls, an indefinite control effect for its
// owner (CR 611.2a, 613.1b): a new effect, so it is later than every control
// effect already on the creature and wins over them, and a creature an owner
// already controls needs none.
//
// No simplification.
func init() {
	const combat = "Alicia Masters, Skilled Sculptor — create a Treasure token"
	Register(Spec{
		OracleID:     "223504ba-174a-46f2-a4a2-5d663a82dfd3",
		Name:         "Alicia Masters, Skilled Sculptor",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			InterveningIf(CondBoard{
				Name: "you've cast a noncreature spell this turn",
				Test: func(ctx *Context) bool { return ctx.Game.CastTallyFor(ctx.Controller()).Noncreature > 0 },
			}, On(game.EventStepBegan, AllOf(StepBegan(game.StepBeginCombat, true), msc16YouCastANoncreatureSpellThisTurn),
				combat, Do(CreateToken{Template: TreasureToken(), N: 1}))),
			AtYourEndStep("Alicia Masters, Skilled Sculptor — Sense the Good: each player gains control of all creatures they own",
				Do(Instruction(msc16EachPlayerRegainsTheirCreatures))),
		},
	})
}

// msc16YouCastANoncreatureSpellThisTurn is the trigger-time half of an
// intervening "if you've cast a noncreature spell this turn".
func msc16YouCastANoncreatureSpellThisTurn(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return g.CastTallyFor(source.Controller).Noncreature > 0
}

// msc16EachPlayerRegainsTheirCreatures is "each player gains control of all
// creatures they own".
func msc16EachPlayerRegainsTheirCreatures(ctx *Context) error {
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if !c.IsCreature() || c.Controller == c.Owner {
			continue
		}
		if err := (GainControl{
			Target:     c.InstanceID,
			Controller: c.Owner,
			Duration:   game.IndefiniteDuration(),
			Label:      "Sense the Good — its owner gains control of it",
		}).Apply(ctx.asGroupMember()); err != nil {
			return err
		}
	}
	return nil
}
