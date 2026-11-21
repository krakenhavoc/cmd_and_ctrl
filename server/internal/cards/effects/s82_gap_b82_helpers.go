package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b82_helpers.go — shared bodies for the GAP set, batch 82 (Jurassic
// World Collection).

// gapB82DamageDealtToOpponents is "the damage dealt to your opponents this
// turn" (Bloodthirst X, CR 702.54b): the total of every damage event this
// turn whose target is a player other than `controller`. Damage dealt, not
// life lost; a prevented point was never dealt.
func GapB82DamageDealtToOpponents(g *game.Game, controller uuid.UUID) int {
	total := 0
	for _, ev := range g.EventsThisTurn() {
		if ev.Kind != game.EventDealDamage || ev.Amount <= 0 || ev.Target == controller {
			continue
		}
		if g.PlayerByIDForEffect(ev.Target) == nil {
			continue
		}
		total += ev.Amount
	}
	return total
}

const (
	GapB82IndoraptorTakeLabel = "Take the damage"
	gapB82IndoraptorSacLabel  = "Sacrifice a nontoken creature"
)

// gapB82NontokenCreatures are the nontoken creatures `player` controls.
func gapB82NontokenCreatures(g *game.Game, player uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == player && c.IsCreature() && !c.IsToken() {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// gapB82IndoraptorEnrage is "choose an opponent at random. Indoraptor deals
// damage equal to its power to that player unless they sacrifice a nontoken
// creature of their choice": the chosen opponent decides; with no nontoken
// creature to sacrifice the damage is dealt without asking.
func GapB82IndoraptorEnrage(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var live []uuid.UUID
	for _, o := range ctx.Opponents() {
		if p := g.PlayerByIDForEffect(o); p != nil && !p.Eliminated {
			live = append(live, o)
		}
	}
	if len(live) == 0 {
		return nil
	}
	pick := g.ChooseAtRandomForEffect(RandomDraw(ctx), live, 1)
	if len(pick) == 0 {
		return nil
	}
	victim := pick[0]
	if len(gapB82NontokenCreatures(g, victim)) == 0 {
		return gapB82IndoraptorDamage(ctx, victim)
	}
	return PickOption{
		Player:   victim,
		Question: "Indoraptor — take damage equal to its power, or sacrifice a nontoken creature?",
		Options: []game.ChoiceOption{
			{Label: GapB82IndoraptorTakeLabel},
			{Label: gapB82IndoraptorSacLabel},
		},
		ThenKey: gapB82IndoraptorAnswer,
	}.Apply(ctx)
}

func gapB82IndoraptorDamage(ctx *Context, victim uuid.UUID) error {
	src, ok := ctx.Game.LookupCardForEffect(ctx.Source())
	if !ok {
		return nil
	}
	return DealDamage{Source: ctx.Source(), Target: victim, Amount: src.CurrentPower()}.Apply(ctx)
}

// gapB82IndoraptorAnswer is the opponent's answer. An on-disk key: never
// renamed, never reused. A refusal to answer takes the damage.
var gapB82IndoraptorAnswer = OptionPickThen("option-pick/gap-b82-indoraptor-answer", func(ctx *Context, r game.OptionPicked) error {
	victim := r.Chooser
	if r.Option != nil && r.Option.Label == gapB82IndoraptorSacLabel {
		return SacrificeChoice{
			Player:     victim,
			Candidates: gapB82NontokenCreatures(ctx.Game, victim),
			Question:   "Indoraptor — sacrifice a nontoken creature",
		}.Apply(ctx)
	}
	return gapB82IndoraptorDamage(ctx, victim)
})

// gapB82OpponentCreaturesWithPower4 counts the creatures with power 4 or
// greater `q.Controller`'s opponents control (Grim Giganotosaurus's cost
// reduction), read through the layered view.
func GapB82OpponentCreaturesWithPower4(q game.CostQuery) int {
	if q.Game == nil {
		return 0
	}
	n := 0
	for _, c := range q.Game.BattlefieldCardsForEffect() {
		if c.Controller != q.Controller && c.IsCreature() && c.CurrentPower() >= 4 {
			n++
		}
	}
	return n
}

var GapB82OwenKinds = []string{game.CounterReach, game.CounterMenace, game.CounterTrample, game.CounterHaste}

// gapB82OwenCounterChosen puts the chosen counter on the ability's target if
// it is still a legal one. An on-disk key: never renamed, never reused.
var GapB82OwenCounterChosen = OptionPickThen("option-pick/gap-b82-owen-grady-counter", func(ctx *Context, r game.OptionPicked) error {
	if r.Index < 0 || r.Index >= len(GapB82OwenKinds) {
		return nil
	}
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		return ctx.Game.AddCounterByForEffect(ctx.Controller(), t.ID, GapB82OwenKinds[r.Index], 1)
	}
	return nil
})

// GrantedAltCostProwl is the key of a granted prowl offer (Hunting
// Velociraptor). It is "granted-"-namespaced like every granted offer and is
// an on-disk identity: never renamed, never reused.
const GrantedAltCostProwl = "granted-prowl"

// gapB82SavageOrderTail is "It gains indestructible until your next turn" on
// the card Savage Order found. An on-disk key: never renamed, never reused.
var GapB82SavageOrderTail = CardTail("tail/gap-b82-savage-order-indestructible", func(ctx *Context, o game.TailOutcome) error {
	for _, id := range o.IDs {
		if err := (ScopedEffectFor{
			Target:   id,
			Mods:     []game.Mod{game.AddKeywordsMod("indestructible")},
			Duration: DurationUntilYourNextTurn(ctx, ctx.Controller()),
			Label:    "Savage Order — indestructible until your next turn",
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})

// gapB82LandDamagesCreatureBody is Swooping Pteranodon's delayed trigger: "at
// the beginning of the next end step, target land deals 3 damage to that
// creature". The scheduled Cards are the land then the creature; either gone
// from the battlefield and nothing happens (CR 608.2b). An on-disk key: never
// renamed, never reused.
var GapB82LandDamagesCreatureBody = game.SimpleDelayedBody("gap-b82/land-damages-creature", func(g *game.Game, item *game.StackItem) error {
	if len(item.Targets) < 2 {
		return nil
	}
	land, creature := item.Targets[0].ID, item.Targets[1].ID
	for _, id := range []uuid.UUID{land, creature} {
		if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneBattlefield {
			return nil
		}
	}
	return DealDamage{Source: land, Target: creature, Amount: 3}.Apply(NewContext(g, item))
})
