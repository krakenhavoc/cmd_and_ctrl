package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tidal Flats — Enchantment {U} (Fallen Empires, S82 OLD batch 852, #5637):
//
//	"{U}{U}: For each attacking creature without flying, its controller may pay {1}.
//	 If that player doesn't, creatures you control blocking that creature gain first
//	 strike until end of turn."
//
// A chain of pay-unless prompts (tidalFlatsStart): the attacking
// creatures without flying are read as the ability resolves, and each one's
// controller is asked in turn whether to pay {1}. On a decline, or a payment that
// can't be made, the Flats' controller's creatures blocking that attacker gain
// first strike until end of turn; either answer moves on to the next attacker. No
// attackers means nothing is asked.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bf39d0f7-f689-45fd-b37b-12d1618fd6e7",
		Name:         "Tidal Flats",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}{U}: For each attacking creature without flying, its controller may pay {1}. If that player doesn't, creatures you control blocking that creature gain first strike until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{U}{U}"),
			Effect:  tidalFlatsStart,
		}},
	})
}

// tidalFlatsStart is the whole effect of Tidal Flats: for each
// attacking creature without flying, in battlefield order, its controller may pay
// {1}; if that player doesn't, creatures the Flats' controller controls that are
// blocking that creature gain first strike until end of turn. The attackers are
// read as the ability resolves and ride the chain as data (the prompt's Carry),
// each asked in turn, so a table waiting on an answer is a restore point.
func tidalFlatsStart(g *game.Game, item *game.StackItem) error {
	var attackers []uuid.UUID
	for _, id := range effects.AttackingCreatures(g) {
		c, ok := g.LookupCardForEffect(id)
		if !ok || game.HasKeyword(&c, "flying") {
			continue
		}
		attackers = append(attackers, id)
	}
	return tidalFlatsAsk(effects.NewContext(g, item), attackers)
}

// tidalFlatsAsk asks the controller of the first attacker left whether
// to pay {1}; both answers carry on with the rest. Nothing is asked of an
// attacker that has left combat or the battlefield by now.
func tidalFlatsAsk(ctx *effects.Context, attackers []uuid.UUID) error {
	for len(attackers) > 0 {
		c, ok := ctx.Game.LookupCardForEffect(attackers[0])
		if ok && c.AttackingTarget != uuid.Nil {
			return effects.PayUnless{
				Chooser:  c.Controller,
				Cost:     "{1}",
				Question: "Tidal Flats — pay {1}, or the creatures blocking " + c.Name + " gain first strike?",
				Else:     tidalFlatsAnswer,
				Carry:    append([]uuid.UUID(nil), attackers...),
			}.Apply(ctx)
		}
		attackers = attackers[1:]
	}
	return nil
}

// tidalFlatsAnswer is the keyed answer: a decline (or a payment the
// chooser could not make) gives first strike to the Flats' controller's blockers
// of the attacker just asked about; either answer moves on to the next. An
// on-disk key: never renamed, never reused.
var tidalFlatsAnswer game.PayUnlessElse

func init() {
	tidalFlatsAnswer = effects.PayUnlessBranches("pay-unless/s82-old-b852-tidal-flats",
		func(ctx *effects.Context, r game.PayUnlessAnswer) error {
			if len(r.Carry) == 0 {
				return nil
			}
			return tidalFlatsAsk(ctx, r.Carry[1:])
		},
		func(ctx *effects.Context, r game.PayUnlessAnswer) error {
			if len(r.Carry) == 0 {
				return nil
			}
			for _, id := range effects.BlockersOf(ctx.Game, r.Carry[0]) {
				c, ok := ctx.Game.LookupCardForEffect(id)
				if !ok || c.Controller != ctx.Controller() {
					continue
				}
				if err := (effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"first strike"},
					Label: "Tidal Flats — first strike"}).Apply(ctx.AsGroupMember()); err != nil {
					return err
				}
			}
			return tidalFlatsAsk(ctx, r.Carry[1:])
		})
}
