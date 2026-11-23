package cardssh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Simic Guildmage — Creature — Elf Wizard {G/U}{G/U}, 2/2 (Ravnica
// Remastered, #5948):
//
//	"{1}{G}: Move a +1/+1 counter from target creature onto another target
//	 creature with the same controller.
//	 {1}{U}: Attach target Aura attached to a permanent to another
//	 permanent with the same controller."
//
// The first ability's second target is related to the first
// (ControlledByTargetsController, as Breaking of the Fellowship's). Moving a
// counter is putting it on the second creature and removing it from the first,
// and if either can't be done neither is (CR 122.5): the counter lands first,
// through the ordinary counter window, and comes off the first creature only
// if it did (Simic Fluxmage's order).
//
// The second ability targets only the Aura (Autumn-Tail's shape). The other
// permanent is chosen as it resolves, from the permanents other than the
// Aura's current host that its controller also controls and the Aura can
// legally enchant (CR 701.3b, game.AttachLegalForEffect); with none, or an
// Aura that has since left or moved to a host of another controller, nothing
// happens.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8b70d296-2756-4bda-b51e-f93be395fce7",
		Name:         "Simic Guildmage",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{1}{G}: Move a +1/+1 counter from target creature onto another target creature with the same controller.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    effects.ManaCost("{1}{G}"),
				Targets: effects.Clauses(
					effects.TargetCreature("target creature"),
					effects.Distinct(effects.ControlledByTargetsController(effects.TargetCreature("another target creature with the same controller"), 0)),
				),
				Effect: simicGuildmageMoveCounter,
			},
			{
				Label: "{1}{U}: Attach target Aura attached to a permanent to another permanent with the same controller.",
				Cost:  effects.ManaCost("{1}{U}"),
				Targets: effects.BattlefieldSpec("permanent", "target Aura attached to a permanent",
					func(g *game.Game, _ uuid.UUID, c game.Card) bool {
						if !c.IsAura() || c.AttachedTo.Kind != game.TargetCard {
							return false
						}
						_, ok := g.LookupCardForEffect(c.AttachedTo.ID)
						return ok && effects.OnBattlefield(g, c.AttachedTo.ID)
					}),
				Effect: simicGuildmageMoveAura,
			},
		},
	})
}

// simicGuildmageMoveCounter is the first ability: both targets must still be
// legal and the first must still hold a +1/+1 counter (CR 122.5).
func simicGuildmageMoveCounter(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	from, okFrom := ctx.ClauseTarget(0)
	to, okTo := ctx.ClauseTarget(1)
	if !okFrom || !okTo || from.Kind != game.TargetCard || to.Kind != game.TargetCard {
		return nil
	}
	src, ok := g.LookupCardForEffect(from.ID)
	if !ok || src.Counters[game.CounterPlusOne] <= 0 {
		return nil
	}
	then := ctx.Tail(simicGuildmageTakeOne)
	then.Carry = []uuid.UUID{from.ID}
	return g.AddCounterByKeyedForEffect(item.Controller, to.ID, game.CounterPlusOne, 1, then)
}

// simicGuildmageTakeOne is "remove a +1/+1 counter from the first creature",
// if the one onto the second landed (the outcome's Amount). An on-disk key:
// never renamed, never reused.
var simicGuildmageTakeOne = game.RegisterTail("simic-guildmage/take-one", func(g *game.Game, o game.TailOutcome) error {
	if o.Amount <= 0 || len(o.Tail.Carry) == 0 {
		return nil
	}
	if c, ok := g.LookupCardForEffect(o.Tail.Carry[0]); !ok || c.Counters[game.CounterPlusOne] <= 0 {
		return nil
	}
	return g.AddCounterForEffect(o.Tail.Carry[0], game.CounterPlusOne, -1)
})

// simicGuildmageMoveAura asks the controller which permanent to move the
// targeted Aura to, among those with the same controller as its host.
func simicGuildmageMoveAura(g *game.Game, item *game.StackItem) error {
	auraID := effects.FirstLegalCardTarget(effects.NewContext(g, item))
	if auraID == uuid.Nil {
		return nil
	}
	aura, ok := g.LookupCardForEffect(auraID)
	if !ok || aura.AttachedTo.Kind != game.TargetCard {
		return nil
	}
	current, ok := g.LookupCardForEffect(aura.AttachedTo.ID)
	if !ok || !effects.OnBattlefield(g, current.InstanceID) {
		return nil
	}
	var hosts []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.InstanceID != current.InstanceID && c.Controller == current.Controller && g.AttachLegalForEffect(auraID, c.InstanceID) {
			hosts = append(hosts, c.InstanceID)
		}
	}
	if len(hosts) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   item.SourceCardID,
		Question: "Simic Guildmage — attach " + aura.Name + " to another permanent with the same controller",
		Cards:    hosts,
		Min:      1,
		Max:      1,
		ThenKey:  simicGuildmageAttachPick,
		LiveZone: game.ZoneBattlefield,
		Carry:    []uuid.UUID{auraID},
	})
	return nil
}

// simicGuildmageAttachPick is the Aura (the one carried ID) attached to the
// permanent picked.
var simicGuildmageAttachPick = game.RegisterChooseCardsThen("choose-cards/simic-guildmage-attach", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Picked) == 0 || len(r.Carry) == 0 {
		return nil
	}
	return g.AttachForEffect(r.Carry[0], game.TargetRef{Kind: game.TargetCard, ID: r.Picked[0]})
})
