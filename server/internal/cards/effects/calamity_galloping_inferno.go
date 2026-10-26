package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Calamity, Galloping Inferno — Legendary Creature — Horse Mount
// {4}{R}{R}, 4/6:
//
//	"Haste
//	 Whenever Calamity attacks while saddled, choose a nonlegendary
//	 creature that saddled it this turn and create a tapped and
//	 attacking token that's a copy of it. Sacrifice that token at the
//	 beginning of the next end step. Repeat this process once.
//	 Saddle 1"
//
// Haste is a PrintedKeyword; Saddle 1 is Saddle(1).
//
// Two rounds of one shape (ADR 0186 slice 6). Each round reads
// SaddlersOfThis as it starts, by Calamity's object, and offers the
// nonlegendary creatures among them in a card-set prompt of exactly one
// ("choose" is not "target": hexproof and protection are irrelevant, and a
// saddler that has left the battlefield is simply not offered). The token
// is a copy of the chosen creature as it exists then, made tapped and
// attacking through CreateTokensAttackingYourChoice, so it was never
// declared as an attacker and triggers nothing that watches attacks
// (CR 508.4); its controller chooses what each token attacks (Calamity's
// own defender is the default). One delayed trigger per token sacrifices it
// at the beginning of the next end step (CR 603.7, Sandstorm Crasher's
// shape).
//
// "Repeat this process once" starts the second round from the first
// round's continuation, after the first token is made and its sacrifice is
// scheduled, so a creature chosen in the first round may be chosen again.
// With no nonlegendary saddler on the battlefield the first round ends the
// ability, since the second would offer the same empty set.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "eb37496e-374f-41b9-bd52-f52f37bdac31",
		Name:            "Calamity, Galloping Inferno",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Activated:       []ActivatedAbility{Saddle(1)},
		Triggered: []game.TriggeredAbility{
			AttacksWhileSaddled("Calamity, Galloping Inferno — twice, copy a nonlegendary creature that saddled it, tapped and attacking; sacrifice each copy at the next end step",
				func(g *game.Game, item *game.StackItem) error {
					return calamityRound(g, item, 2)
				}),
		},
	})
}

// calamityRound is one pass of the printed process; `left` rounds remain,
// this one included.
func calamityRound(g *game.Game, item *game.StackItem, left int) error {
	if left <= 0 {
		return nil
	}
	var candidates []uuid.UUID
	for _, id := range SaddlersOfThis(NewContext(g, item)) {
		if c, ok := g.LookupCardForEffect(id); ok && c.IsCreature() && !c.IsLegendary() {
			candidates = append(candidates, id)
		}
	}
	if len(candidates) == 0 {
		return nil
	}
	// #4294: the pick is a registered continuation handed the trigger's item
	// and the rounds left as data, so a table waiting on it is a restore
	// point. LiveZone keeps the closure form's battlefield re-check.
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:   item.Controller,
		Source:    item.SourceCardID,
		Question:  "Calamity, Galloping Inferno — choose a nonlegendary creature that saddled it",
		Cards:     candidates,
		Min:       1,
		Max:       1,
		ThenKey:   calamityPick,
		LiveZone:  game.ZoneBattlefield,
		CarryNums: []int{left},
		Asked:     game.NewAskedItem(item),
	})
	return nil
}

// calamityPickBody is the continuation of the pick: a tapped and attacking
// copy of the chosen saddler, whose controller chooses what it attacks,
// then the rest of the round (calamityNextRound).
func calamityPickBody(g *game.Game, r game.CardsChosen) error {
	item := r.Asked
	if item == nil || len(r.Nums) == 0 || len(r.Picked) == 0 {
		return nil
	}
	left := r.Nums[0]
	tmpl, ok := TokenCopyTemplate(g, r.Picked[0])
	if !ok {
		return calamityRound(g, item, left-1)
	}
	ctx := NewContext(g, item)
	return CreateTokensAttackingYourChoice{
		Template: tmpl,
		Copy:     true,
		N:        1,
		Tapped:   true,
		Prefer:   triggerAttackTarget(item),
		Label:    "Calamity, Galloping Inferno",
		Tail:     withN(ctx.ItemTail(calamityNextRound), left-1),
	}.Apply(ctx)
}

// calamityNextRound is "sacrifice the token at the next end step", then the
// next copy, as a registered tail (#4253 slice 6): the rounds left are the
// tail's N and the asking item rides it as data. An on-disk key.
var (
	calamityNextRound game.TailRef
	// calamityPick is the pick before the split (#4294). An on-disk key.
	calamityPick game.ChooseCardsThen
)

func init() {
	calamityPick = game.RegisterChooseCardsThen("choose-cards/calamity-galloping-inferno-copy", calamityPickBody)
	// Registered here, not in the var, because the body starts the next
	// round, which names this tail.
	calamityNextRound = AttackingTokensThen("attacking-tokens/calamity-sacrifice-then-next-round",
		func(ctx *Context, created []uuid.UUID, tail game.KeyedTail) error {
			item := tail.Asked.StackItem()
			if item == nil {
				return nil
			}
			if len(created) > 0 {
				if err := (ScheduleDelayedTrigger{
					Label: "Calamity, Galloping Inferno — sacrifice the token",
					Cards: created,
					Body:  sacrificeListedCardsBody,
				}).Apply(NewContext(ctx.Game, item)); err != nil {
					return err
				}
			}
			return calamityRound(ctx.Game, item, tail.N)
		})
}
