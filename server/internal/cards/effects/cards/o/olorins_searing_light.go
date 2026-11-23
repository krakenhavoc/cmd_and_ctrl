package cardso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Olórin's Searing Light — Instant {2}{R}{W} (Tales of Middle-earth
// Commander):
//
//	"Each opponent exiles a creature with the greatest power among
//	 creatures that player controls.
//	 Spell mastery — If there are two or more instant and/or sorcery cards
//	 in your graveyard, Olórin's Searing Light deals damage to each
//	 opponent equal to the power of the creature they exiled."
//
// Each opponent chooses, for themselves, among the creatures that tie for
// the greatest power on their own side (ChoosePermanents with that
// opponent as the chooser, one prompt per opponent in seat order), then the
// creature is exiled. Spell mastery is read once as the spell begins to
// resolve, when it is itself still on the stack (so it is not one of the two
// cards); the damage is the exiled creature's power as it last stood on the
// battlefield, dealt by the spell to the opponent who exiled it, and only if
// a creature really was exiled. An opponent with no creature exiles nothing
// and takes nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e6f2ba06-2c92-48b7-a95f-ea6bf35aa0fe",
		Name:         "Olórin's Searing Light",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			mastery := 0
			if effects.SpellMastery(ctx.Game, ctx.Controller()) {
				mastery = 1
			}
			for _, opp := range ctx.Opponents() {
				if err := (effects.ChoosePermanents{
					Player:   opp,
					Question: "Olórin's Searing Light — exile a creature with the greatest power among creatures you control",
					Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
						best, ids := -1, []uuid.UUID(nil)
						for _, c := range g.BattlefieldCardsForEffect() {
							if c.Controller != of || !c.IsCreature() {
								continue
							}
							switch p := c.CurrentPower(); {
							case p > best:
								best, ids = p, []uuid.UUID{c.InstanceID}
							case p == best:
								ids = append(ids, c.InstanceID)
							}
						}
						return ids, 1, 1
					},
					Keyed: s82GapB119OlorinExile,
					Carry: []uuid.UUID{opp},
					Nums:  []int{mastery},
				}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}

var s82GapB119OlorinExile = effects.RegisterPermanentsPick("prompt-run/olorins-searing-light-exile", effects.PermanentsDef{
	ThenRun: func(ctx *effects.Context, picked game.PromptedPicks, r game.RunDone) error {
		for _, id := range picked.Cards() {
			c, ok := ctx.Game.LookupCardForEffect(id)
			if !ok {
				continue
			}
			then := game.KeyedTail{}
			// Spell mastery: the damage waits for the exile and is gated on the
			// creature really having reached exile (the CR 614 window may cancel
			// or redirect the move), by the creature's power as it last stood.
			if len(r.Nums) > 0 && r.Nums[0] == 1 && len(r.Carry) > 0 {
				then = ctx.ItemTail(s82GapB119OlorinDamage)
				then.N = c.CurrentPower()
				then.Carry = []uuid.UUID{r.Carry[0]}
			}
			if err := (effects.ExileTarget{Target: id, ThenTail: then}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	},
})

// s82GapB119OlorinDamage is the spell mastery damage: the exiled creature's
// power (the tail's N) to the opponent who exiled it (the tail's first ID).
var s82GapB119OlorinDamage = effects.CardTail("olorins-searing-light/mastery-damage", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 || o.Tail.N <= 0 || len(o.Tail.Carry) == 0 {
		return nil
	}
	return effects.DealDamage{Source: ctx.Source(), Target: o.Tail.Carry[0], Amount: o.Tail.N}.Apply(ctx)
})
