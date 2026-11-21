package cardso

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Overencumbered — Enchantment — Aura {1}{W} (Fallout):
//
//	"Enchant opponent
//	 When this Aura enters, enchanted opponent creates a Clue token, a Food
//	 token, and a Junk token.
//	 At the beginning of combat on enchanted opponent's turn, that player
//	 may pay {1} for each artifact they control. If they don't, creatures
//	 can't attack this combat."
//
// "Enchant opponent" is Archnemesis's player clause. The enters trigger
// makes the three tokens for the enchanted player, not for me (they are
// created, not investigated or foraged, so no Clue-watching card sees an
// investigate). The combat trigger counts the artifacts that player
// controls as it resolves; with none there is nothing to pay and nothing
// happens (paying {0} is always possible). Otherwise the enchanted player
// answers pay-or-not, and a "no", or a payment they cannot make, adds a rule
// for the rest of THIS combat phase only (CR 511.3; an additional combat is
// not covered): no creature can attack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "54400ec3-f2ab-4cdf-9731-e04281bf7e90",
		Name:         "Overencumbered",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("enchant opponent", effects.Opponent()),
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Overencumbered — enchanted opponent creates a Clue token, a Food token, and a Junk token",
				func(g *game.Game, item *game.StackItem) error {
					who := overencumberedEnchanted(g, item.SourceCardID)
					if who == uuid.Nil {
						return nil
					}
					return effects.Do(
						effects.CreateToken{Controller: who, Template: effects.ClueToken(), N: 1},
						effects.CreateToken{Controller: who, Template: effects.FoodToken(), N: 1},
						effects.CreateToken{Controller: who, Template: effects.JunkToken(), N: 1},
					)(g, item)
				}),
			effects.On(game.EventStepBegan, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Kind == game.EventStepBegan && ev.Step == game.StepBeginCombat &&
					source.AttachedTo.Kind == game.TargetPlayer && effects.IsActivePlayer(g, source.AttachedTo.ID)
			}, "Overencumbered — that player may pay {1} for each artifact they control; if they don't, creatures can't attack this combat",
				func(g *game.Game, item *game.StackItem) error {
					who := overencumberedEnchanted(g, item.SourceCardID)
					if who == uuid.Nil {
						return nil
					}
					n := 0
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller == who && c.IsArtifact() {
							n++
						}
					}
					if n == 0 {
						return nil
					}
					return effects.PayUnless{
						Chooser:  who,
						Cost:     "{" + strconv.Itoa(n) + "}",
						Question: "Overencumbered — pay {" + strconv.Itoa(n) + "}? If you don't, creatures can't attack this combat.",
						Else:     overencumberedNoAttack,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

// overencumberedEnchanted is the player the Aura is attached to.
func overencumberedEnchanted(g *game.Game, aura uuid.UUID) uuid.UUID {
	c, ok := g.LookupCardForEffect(aura)
	if !ok || c.AttachedTo.Kind != game.TargetPlayer {
		return uuid.Nil
	}
	return c.AttachedTo.ID
}

// overencumberedNoAttack is the decline: no creature can attack for the rest
// of this combat phase. An on-disk key: never renamed, never reused.
var overencumberedNoAttack = effects.PayUnlessElse("pay-unless/overencumbered-no-attack", func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
	d, ok := effects.DurationUntilEndOfCombat(ctx)
	if !ok {
		return nil
	}
	ctx.Game.RegisterScopedRuleEffectForEffect(ctx.Source(), game.ScopeCreatures, ctx.Controller(),
		[]game.Mod{game.AddRestrictionsMod(game.CantAttack)}, d, "Overencumbered — creatures can't attack this combat")
	return nil
})
