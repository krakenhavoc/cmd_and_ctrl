package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// OldB804CantAttackIfAnotherCreature is "This creature can't attack if
// there's another creature on the battlefield" (Shauku, Endbringer): any other
// creature, whoever controls it, read live at declaration (CR 508.1c).
func OldB804CantAttackIfAnotherCreature() game.StaticAbility {
	return selfAttackTargetRestriction(game.AttackTargetRestriction{
		AnotherOnBattlefieldForbids: []game.PermanentQuery{{Types: []string{"creature"}}},
	})
}

// OldB804AttackedYouThisTurn passes a creature that was declared as an attacker
// this turn with the caster as the defending player (Jabari's Influence's
// "that attacked you this turn"). The record names the object, so a creature
// that left and came back is a new one with no history (CR 400.7), and it stays
// true after the creature has changed hands.
func OldB804AttackedYouThisTurn() CardPredicate {
	return func(g *game.Game, you uuid.UUID, c game.Card) bool {
		for _, a := range g.TurnTally.Attacks {
			if a.Attacker == c.InstanceID && a.Epoch == c.ObjectEpoch && a.Defender == you {
				return true
			}
		}
		return false
	}
}

// OldB804CreatureYouCastThisTurn passes a creature permanent that is the
// object a spell `you` cast this turn became (Cycle of Life's "target
// creature you cast this turn"): its provenance says it was cast this turn and
// this turn's cast events say who cast it.
func OldB804CreatureYouCastThisTurn() CardPredicate {
	return func(g *game.Game, you uuid.UUID, c game.Card) bool {
		if !c.Provenance.CastThisTurn {
			return false
		}
		for _, ev := range g.EventsThisTurn() {
			if ev.Kind == game.EventCast && ev.CardID == c.InstanceID && ev.Actor == you {
				return true
			}
		}
		return false
	}
}

// oldB804CycleOfLifeCounterBody is Cycle of Life's delayed trigger: "put a
// +1/+1 counter on that creature", only if it is still the object it was when
// the ability resolved (CR 400.7, 603.7c). An on-disk key.
var oldB804CycleOfLifeCounterBody = game.DelayedBody("old-b804/cycle-of-life-counter",
	func(g *game.Game, item *game.StackItem, p game.EffectParams) error {
		for _, ref := range p.Objects {
			c, ok := BattlefieldCardForEffect(g, ref.ID)
			if !ok || c.ObjectEpoch != ref.Epoch {
				continue
			}
			if err := (AddCounter{Target: ref.ID, Kind: game.CounterPlusOne, N: 1}).Apply(NewContext(g, item)); err != nil {
				return err
			}
		}
		return nil
	})

// OldB804CycleOfLifeResolve is Cycle of Life's ability: the target gets base
// power and toughness 0/1 until the controller's next upkeep, and a delayed
// trigger at that upkeep puts a +1/+1 counter on it.
func OldB804CycleOfLifeResolve(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	t, ok := ctx.ClauseTarget(0)
	if !ok || t.Kind != game.TargetCard {
		return nil
	}
	ref, ok := g.PermanentRefForEffect(t.ID)
	if !ok {
		return nil
	}
	if err := (ScopedEffectFor{
		Target:   t.ID,
		Mods:     game.SetBasePTMods(0, 1),
		Duration: DurationUntilYourNextUpkeep(ctx, ctx.Controller()),
		Label:    "Cycle of Life — base power and toughness 0/1 until your next upkeep",
	}).Apply(ctx); err != nil {
		return err
	}
	return ScheduleDelayedTrigger{
		At:                 game.StepUpkeep,
		ControllerTurnOnly: true,
		Label:              "Cycle of Life — put a +1/+1 counter on that creature",
		Cards:              []uuid.UUID{t.ID},
		Body:               oldB804CycleOfLifeCounterBody,
		Params:             game.EffectParams{Objects: []game.ObjectRef{ref}},
	}.Apply(ctx)
}

var oldB804NonartifactCreature = SacrificePayment(1, "nonartifact creature", "nonartifact creatures",
	game.PermanentQuery{Types: []string{"creature"}, WithoutTypes: []string{"artifact"}})

// OldB804RavenousVampireOffer is Ravenous Vampire's upkeep trigger: "you may
// sacrifice a nonartifact creature. If you do, put a +1/+1 counter on this
// creature. If you don't, tap this creature." A controller with nothing to
// sacrifice is not asked and doesn't.
func OldB804RavenousVampireOffer(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	cost, action := oldB804NonartifactCreature.Prompt()
	if !g.PayActionCanPayForEffect(ctx.Controller(), action) {
		return oldB804VampireTap(ctx)
	}
	return UpkeepPayUnless{
		Chooser:  ctx.Controller(),
		Cost:     cost,
		Action:   action,
		Question: "Ravenous Vampire — sacrifice a nonartifact creature to put a +1/+1 counter on it? (If you don't, tap it.)",
		Else:     oldB804VampireBranches,
	}.Apply(ctx)
}

func oldB804VampireTap(ctx *Context) error {
	if !SourceIsStillThisPermanent(ctx.Game, ctx.Item) {
		return nil
	}
	return TapTarget{Target: ctx.Item.SourceCardID}.Apply(ctx)
}

// oldB804VampireBranches registers both answers under one key. An on-disk key:
// never renamed, never reused.
var oldB804VampireBranches = PayUnlessBranches("pay-unless/s82-old-b804-ravenous-vampire",
	func(ctx *Context, _ game.PayUnlessAnswer) error { return PutCounterOnSelf(ctx.Game, ctx.Item) },
	func(ctx *Context, _ game.PayUnlessAnswer) error { return oldB804VampireTap(ctx) })

// OldB804FungusCounter is the fungus counter Mindbender Spores puts on a
// creature it blocks. Counter names are free-form strings (game/counter_types.go).
const OldB804FungusCounter = "fungus"
