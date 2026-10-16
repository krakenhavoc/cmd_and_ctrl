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
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   item.SourceCardID,
		Question: "Calamity, Galloping Inferno — choose a nonlegendary creature that saddled it",
		Cards:    candidates,
		Min:      1,
		Max:      1,
		Zone:     game.ZoneBattlefield,
		Then: func(g *game.Game, picked []uuid.UUID) error {
			if len(picked) == 0 {
				return nil
			}
			next := func(g *game.Game) error { return calamityRound(g, item, left-1) }
			tmpl, ok := TokenCopyTemplate(g, picked[0])
			if !ok {
				return next(g)
			}
			return CreateTokensAttackingYourChoice{
				Template: tmpl,
				Copy:     true,
				N:        1,
				Tapped:   true,
				Prefer:   triggerAttackTarget(item),
				Label:    "Calamity, Galloping Inferno",
				Then: func(ctx *Context, created []uuid.UUID) error {
					if len(created) > 0 {
						if err := (ScheduleDelayedTrigger{
							Label: "Calamity, Galloping Inferno — sacrifice the token",
							Cards: created,
							Body:  sacrificeListedCardsBody,
						}).Apply(ctx); err != nil {
							return err
						}
					}
					return next(ctx.Game)
				},
			}.Apply(NewContext(g, item))
		},
	})
	return nil
}
