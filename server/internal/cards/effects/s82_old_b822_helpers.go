package effects

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b822_helpers.go — Homelands, "+1/+1 and other counters" (S82 OLD
// batch 822, #5635). A file of its own so concurrent batches never collide.

// OldB822CombatPartners is "each creature blocking or blocked by this
// creature", read as the trigger resolves: the source as it is now, or as it
// last existed on the battlefield (CR 608.2h).
func OldB822CombatPartners(ctx *Context) []uuid.UUID {
	self, ok := ctx.SourcePermanent()
	if !ok {
		return nil
	}
	src := ctx.Source()
	var out []uuid.UUID
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.InstanceID == src || !c.IsCreature() {
			continue
		}
		blocking := self.Attacking && c.IsBlockingAttacker(src)
		blocked := c.AttackingTarget != uuid.Nil && slices.Contains(self.Blocking, c.InstanceID)
		if blocking || blocked {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// OldB822EachCombatPartnerGetsACounter is "put a <kind> counter on each
// creature blocking or blocked by this creature" (Greater Werewolf).
func OldB822EachCombatPartnerGetsACounter(kind string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		member := ctx.AsGroupMember()
		for _, id := range OldB822CombatPartners(ctx) {
			if err := (AddCounter{Target: id, Kind: kind, N: 1}).Apply(member); err != nil {
				return err
			}
		}
		return nil
	}
}

// OldB822TargetGetsACounter is "put a <kind> counter on target creature".
func OldB822TargetGetsACounter(kind string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		id, ok := B16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		return AddCounter{Target: id, Kind: kind, N: 1}.Apply(ctx)
	}
}

// OldB822AttackedOrBlockedThisCombat is "this creature attacked or blocked
// this combat": the engine's per-combat records of the declarations.
func OldB822AttackedOrBlockedThisCombat(g *game.Game, c *game.Card) bool {
	return c != nil && (g.DeclaredAsAttackerThisCombat(c) || g.DeclaredAsBlockerThisCombat(c))
}

// oldB822ClockworkCounter is the +1/+0 counter of the Homelands Clockwork
// creatures (CR 122.1a).
const oldB822ClockworkCounter = "+1/+0"

// OldB822ClockworkSpec is the shared body of Clockwork Steed and Clockwork
// Swarm: they differ only in the creatures that can't block them.
//
//	"This creature enters with four +1/+0 counters on it.
//	This creature can't be blocked by <bar>.
//	At end of combat, if this creature attacked or blocked this combat, remove
//	a +1/+0 counter from it.
//	{X}, {T}: Put up to X +1/+0 counters on this creature. This ability can't
//	cause the total number of +1/+0 counters on this creature to be greater
//	than four. Activate only during your upkeep."
//
// The end-of-combat trigger is one registered row with its intervening "if"
// (CR 603.4) judged as it triggers and again as it resolves. "Up to X" is a
// number the controller picks, from zero to the most that fits under four.
func OldB822ClockworkSpec(oracleID, name string, bar CardPredicate, barLabel string) Spec {
	const label = " — remove a +1/+0 counter from it"
	cond := CondBoard{Name: "this creature attacked or blocked this combat", Test: func(ctx *Context) bool {
		c, ok := ctx.Game.LookupCardForEffect(ctx.Source())
		return ok && OldB822AttackedOrBlockedThisCombat(ctx.Game, &c)
	}}
	end := AtEachStep(game.StepEndCombat, name+label, RemoveACounterFromThis(oldB822ClockworkCounter))
	steps := end.AppliesTo
	end.AppliesTo = func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
		return steps(ev, source, lki, g) && OldB822AttackedOrBlockedThisCombat(g, source)
	}
	return Spec{
		OracleID:     oracleID,
		Name:         name,
		Completeness: CompletenessFull,
		XMatters:     true,
		Replacements: []game.ReplacementEffect{
			B10EntersWithCounters(oldB822ClockworkCounter, 4, name+": enters with four +1/+0 counters on it"),
		},
		BlockRules: []game.BlockRule{CantBeBlockedBy(OnSelf(), bar, barLabel)},
		Triggered:  []game.TriggeredAbility{InterveningIf(cond, end)},
		Activated: []ActivatedAbility{{
			Label:     "{X}, {T}: Put up to X +1/+0 counters on this creature. This ability can't cause the total number of +1/+0 counters on this creature to be greater than four. Activate only during your upkeep.",
			Purpose:   game.Purpose{Answers: game.AnswerValue},
			Cost:      Plus(ManaCost("{X}"), TapCost()),
			Condition: DuringYourUpkeep(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				c, ok := g.LookupCardForEffect(item.SourceCardID)
				if !ok || !OnBattlefield(g, item.SourceCardID) {
					return nil
				}
				most := min(ctx.X(), 4-c.Counters[oldB822ClockworkCounter])
				if most <= 0 {
					return nil
				}
				return ChooseNumber{
					Min:      0,
					Max:      most,
					Question: name + " — how many +1/+0 counters to put on it?",
					Unit:     game.PayAmountCounters,
					Then:     oldB822ClockworkPut,
				}.Apply(ctx)
			},
		}},
	}
}

// oldB822ClockworkPut puts the chosen number of +1/+0 counters on the source
// if it is still there.
var oldB822ClockworkPut = NumberThen("number/s82-old-b822-clockwork-put-counters",
	func(ctx *Context, r game.NumberAnswer) error {
		if r.Number <= 0 || !OnBattlefield(ctx.Game, ctx.Source()) {
			return nil
		}
		return AddCounter{Target: ctx.Source(), Kind: oldB822ClockworkCounter, N: r.Number}.Apply(ctx)
	})

// OldB822Retribution is "Choose two target creatures controlled by the same
// opponent. That player chooses and sacrifices one of those creatures. Put a
// -1/-1 counter on the other." With one target gone, the other is the one
// sacrificed (CR 608.2b) and no counter is put anywhere.
func OldB822Retribution(_ *game.StackItem, ctx *Context) error {
	var ids []uuid.UUID
	for slot := 0; slot < 2; slot++ {
		if t, ok := ctx.ClauseTarget(slot); ok && t.Kind == game.TargetCard {
			ids = append(ids, t.ID)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	first, found := ctx.Game.LookupCardForEffect(ids[0])
	if !found {
		return nil
	}
	victim := first.Controller
	if len(ids) == 1 {
		return SacrificePermanent{Target: ids[0]}.Apply(ctx)
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    victim,
		FromPlayer: victim,
		Source:     ctx.Source(),
		Question:   "Retribution — choose one of the two creatures to sacrifice; the other gets a -1/-1 counter",
		Cards:      ids,
		Min:        1,
		Max:        1,
		ThenKey:    oldB822RetributionPick,
		LiveZone:   game.ZoneBattlefield,
		Carry:      ids,
		Asked:      game.NewAskedItem(ctx.Item),
	})
	return nil
}

// oldB822RetributionPick sacrifices the chosen creature, then puts the -1/-1
// counter on the other.
var oldB822RetributionPick = game.RegisterChooseCardsThen("choose-cards/s82-old-b822-retribution", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Picked) != 1 || len(r.Carry) != 2 {
		return nil
	}
	ctx, _ := AskedContext(g, r.Asked, r.Chooser, r.Source, nil)
	other := r.Carry[0]
	if other == r.Picked[0] {
		other = r.Carry[1]
	}
	t := ctx.ItemTail(oldB822RetributionCounter)
	t.Carry = []uuid.UUID{other}
	return SacrificePermanent{Target: r.Picked[0], ThenTail: t}.Apply(ctx)
})

// oldB822RetributionCounter is the rest of Retribution once the sacrifice has
// been made, or has been replaced: the -1/-1 counter on the other creature,
// whether or not the first one really left (the text says "put", not "if you
// do").
var oldB822RetributionCounter = CardTail("then/s82-old-b822-retribution-counter", func(ctx *Context, o game.TailOutcome) error {
	if len(o.Tail.Carry) == 0 || !OnBattlefield(ctx.Game, o.Tail.Carry[0]) {
		return nil
	}
	return AddCounter{Target: o.Tail.Carry[0], Kind: game.CounterMinusOne, N: 1}.Apply(ctx)
})

// OldB822DamagedCreatureDies is "Whenever a creature dealt damage by this
// creature this turn dies, put a <kind> counter on this creature" (Sengir Bats,
// Baron Sengir): Sengir Vampire's trigger with the counter kind a parameter.
func OldB822DamagedCreatureDies(label, kind string) game.TriggeredAbility {
	return On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		_, died := DiedCreature(ev, g)
		return died && ev.CardID != source.InstanceID && DealtDamageByThisTurn(g, source.InstanceID, ev.CardID)
	}, label, Do(CounterOnThis{Kind: kind, N: 1}))
}
