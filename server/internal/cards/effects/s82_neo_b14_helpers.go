package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_neo_b14_helpers.go — shared pieces of Kamigawa: Neon Dynasty batch 14
// (#5493, S82).

// neoCreaturesAndVehiclesOf are the creatures and Vehicles `player` controls,
// through the layered view, in battlefield order.
func neoCreaturesAndVehiclesOf(g *game.Game, player uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == player && neoIsCreatureOrVehicle(c) {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

func neoIsCreatureOrVehicle(c game.Card) bool {
	return c.IsCreature() || c.HasSubtype("Vehicle")
}

// neoDistributeAmongPlayersCreaturesAndVehicles is "distribute N +1/+1
// counters among any number of creatures and/or Vehicles target player
// controls" (Invoke Justice). The creatures are not targets, only the
// player is, so the split is chosen as the effect resolves (CR 608.2d): the
// controller picks the permanents, each of which gets at least one, then
// names each share but the last, which takes what is left (matDistributeStep).
// The player is clause `slot` of the spell; a player who is no longer a legal
// target ends the distribution (CR 608.2b).
func neoDistributeAmongPlayersCreaturesAndVehicles(ctx *Context, slot, n int) error {
	t, ok := ctx.ClauseTarget(slot)
	if !ok || t.Kind != game.TargetPlayer || n <= 0 {
		return nil
	}
	cands := neoCreaturesAndVehiclesOf(ctx.Game, t.ID)
	switch len(cands) {
	case 0:
		return nil
	case 1:
		return matPutCounters(ctx, cands[0], n)
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:   ctx.Controller(),
		Source:    ctx.Source(),
		Question:  "Choose any number of that player's creatures and/or Vehicles to distribute +1/+1 counters among (each gets at least one)",
		Cards:     cands,
		Min:       0,
		Max:       min(n, len(cands)),
		ThenKey:   neoDistributePick,
		CarryNums: []int{n, slot},
		LiveZone:  game.ZoneBattlefield,
		Asked:     game.NewAskedItem(ctx.Item),
	})
	return nil
}

// neoDistributePick is the permanents chosen: it starts the shares. A pick
// that is no longer one of the player's creatures or Vehicles is dropped.
var neoDistributePick = askedPick("choose-cards/neo-distribute-counters-among-players-permanents", func(ctx *Context, r game.CardsChosen) error {
	if len(r.Nums) < 2 {
		return nil
	}
	t, ok := ctx.ClauseTarget(r.Nums[1])
	if !ok || t.Kind != game.TargetPlayer {
		return nil
	}
	var picked []uuid.UUID
	for _, id := range r.Picked {
		if c, ok := ctx.Game.LookupCardForEffect(id); ok && c.Controller == t.ID && neoIsCreatureOrVehicle(c) {
			picked = append(picked, id)
		}
	}
	return matDistributeStep(ctx, picked, r.Nums[0])
})

// neoCounterChoices are the counters Invoke the Ancients offers, in the
// printed order: "your choice of a reach counter, a vigilance counter, or a
// trample counter".
var neoCounterChoices = []string{game.CounterReach, game.CounterVigilance, game.CounterTrample}

// neoAskTokenCounter asks the controller for the counter on the first of
// `tokens`, then on the rest, one prompt each ("for each of them"). A token
// that has left the battlefield by the time its answer comes gets nothing.
func neoAskTokenCounter(ctx *Context, tokens []uuid.UUID) error {
	if len(tokens) == 0 {
		return nil
	}
	options := make([]game.ChoiceOption, len(neoCounterChoices))
	for i, kind := range neoCounterChoices {
		options[i] = game.ChoiceOption{Label: "A " + kind + " counter"}
	}
	return PickOption{
		Question: "Invoke the Ancients — put your choice of a reach counter, a vigilance counter, or a trample counter on the Spirit",
		Options:  options,
		ThenKey:  neoTokenCounterChosen,
		Carry:    tokens,
	}.Apply(ctx)
}

// neoTokenCounterChosen puts the chosen counter on the first carried token
// and goes on to the rest. An on-disk key: never renamed, never reused.
var neoTokenCounterChosen game.OptionPickThen

// Registered from init: the answer asks the next token's question through
// neoAskTokenCounter, which names it, so a var initialiser would be a cycle.
func init() {
	neoTokenCounterChosen = OptionPickThen("option-pick/neo-invoke-the-ancients-counter", neoTokenCounterChosenBody)
}

func neoTokenCounterChosenBody(ctx *Context, r game.OptionPicked) error {
	if len(r.Carry) == 0 {
		return nil
	}
	if r.Index >= 0 && r.Index < len(neoCounterChoices) {
		if z := ctx.Game.FindCardZoneForEffect(r.Carry[0]); z != nil && z.Kind == game.ZoneBattlefield {
			if err := (AddCounter{Target: r.Carry[0], Kind: neoCounterChoices[r.Index], N: 1}).Apply(ctx); err != nil {
				return err
			}
		}
	}
	return neoAskTokenCounter(ctx, r.Carry[1:])
}

// neoTokenCounterTail starts the per-token counter choices once the tokens
// have landed. An on-disk key.
var neoTokenCounterTail = CardTail("neo-invoke-the-ancients/counters", func(ctx *Context, o game.TailOutcome) error {
	return neoAskTokenCounter(ctx, o.IDs)
})

// neoSpellManaValue is what a card costs the total of "instant and/or sorcery
// spells with total mana value N or less": the mana value of the instant or
// sorcery face a cast would make (X counts as 0, CR 202.3e), the greatest of
// them for a card with more than one such face. -1 when no face is an instant
// or sorcery.
func neoSpellManaValue(c game.Card) int {
	best := -1
	for _, i := range c.CastableFaces() {
		face := c
		face.SetFace(i)
		if !face.IsInstant() && !face.IsSorcery() {
			continue
		}
		if mv, ok := face.ParsedManaValue(); ok && mv > best {
			best = mv
		}
	}
	return best
}

// neoTotalSpellManaValueAtMost is the set rule "with total mana value N or
// less" over a picked set of instants and sorceries.
func neoTotalSpellManaValueAtMost(limit int) game.SetRuleFunc {
	return func(picked []game.Card) bool {
		total := 0
		for _, c := range picked {
			total += max(neoSpellManaValue(c), 0)
		}
		return total <= limit
	}
}

// SacrificeAnArtifactWithManaValueX is "Sacrifice an artifact with mana value
// X" (Scrap Welder): SacrificeAnotherWithManaValueX without the "other than"
// clause. The activator announces X (DemandsX asks for it as for an {X}) and
// the artifact sacrificed must have mana value X (game.SacrificeManaValueX;
// {X} in its cost counts as 0, a token as 0). It is also the X the ability's
// target clause is bound to.
func neoSacrificeAnArtifactWithManaValueX(label string) game.AbilityCost {
	spec := sacrificeSpec(label, Artifact(), YouControl()).WithCount(1, 1)
	spec.ManaValueEqualsX = true
	return game.AbilityCost{SacrificeOther: spec}
}

// damageXToTheTargetAndGainX is "<this> deals X damage to <target> and you
// gain X life" (Death Grasp, March of Wretched Sorrow). The life is the
// announced X, not the damage dealt, so it is gained in full when the
// damage is prevented; a target that is gone stops the whole spell
// (CR 608.2b), so no life is gained then.
func damageXToTheTargetAndGainX(item *game.StackItem, ctx *Context) error {
	if len(item.Targets) == 0 {
		return nil
	}
	if err := (DealDamage{Source: ctx.Source(), Target: item.Targets[0].ID, Amount: ctx.X()}).Apply(ctx); err != nil {
		return err
	}
	return GainLife{Player: ctx.Controller(), Amount: ctx.X()}.Apply(ctx)
}
