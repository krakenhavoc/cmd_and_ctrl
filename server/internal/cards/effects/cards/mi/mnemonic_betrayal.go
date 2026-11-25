package cardsmi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mnemonic Betrayal — Sorcery {1}{U}{B} (Guilds of Ravnica, #5525):
//
//	"Exile all opponents' graveyards. You may cast spells from among those cards
//	 this turn, and mana of any type can be spent to cast them. At the beginning
//	 of the next end step, if any of those cards remain exiled, return them to
//	 their owners' graveyards.
//	 Exile Mnemonic Betrayal."
//
// Shadow of the Enemy's exile of every card in the opponents' graveyards with a
// cast-only, any-type permission for the caster (a land among them is not played,
// CR 305.1 and the permission's CastOnly), on the default this-turn window. The
// delayed trigger is Glimpse the Impossible's: scheduled for the next end step
// over the exiled cards, it puts those still in exile into their owners'
// graveyards in one move. A card that was cast has left exile and is not
// returned. "Exile Mnemonic Betrayal" is All Suns' Dawn's tail, so the spell goes
// to exile rather than the graveyard.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "aa41e0e6-d04e-48d6-8450-02b972dd0bb1",
		Name:         "Mnemonic Betrayal",
		Completeness: effects.CompletenessFull,
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				var ids []uuid.UUID
				for _, opp := range ctx.Opponents() {
					ids = append(ids, effects.B42CardsInGraveyardOf(ctx.Game, opp, nil)...)
				}
				if len(ids) == 0 {
					return nil
				}
				grant := effects.ExileWithPermission{GrantTo: ctx.Controller(), CastOnly: true, AnyType: true}.Permission()
				if err := effects.ExileAllWithPermission(ctx.Game, ids, grant); err != nil {
					return err
				}
				return effects.ScheduleDelayedTrigger{
					Label: "Mnemonic Betrayal — return the cards still exiled to their owners' graveyards",
					Cards: ids,
					Body:  mnemonicBetrayalReturn,
				}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error { return effects.ExileTarget{Target: ctx.Source()}.Apply(ctx) }),
		),
	})
}

// mnemonicBetrayalReturn is the delayed half: the listed cards still in exile
// go to their owners' graveyards together. An on-disk key: never renamed, never
// reused.
var mnemonicBetrayalReturn = game.SimpleDelayedBody("exile/mnemonic-betrayal-return", func(g *game.Game, item *game.StackItem) error {
	var left []uuid.UUID
	for _, t := range item.Targets {
		if t.Kind != game.TargetCard {
			continue
		}
		if z := g.FindCardZoneForEffect(t.ID); z != nil && z.Kind == game.ZoneExile {
			left = append(left, t.ID)
		}
	}
	if len(left) == 0 {
		return nil
	}
	return g.PutCardsIntoGraveyardKeyedForEffect(left, game.KeyedTail{})
})
