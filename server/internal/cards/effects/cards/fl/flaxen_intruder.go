package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flaxen Intruder // Welcome Home — an adventure card (Throne of Eldraine):
//
//	Flaxen Intruder {G}, Creature — Human Berserker 1/2:
//	"Whenever this creature deals combat damage to a player, you may
//	 sacrifice it. When you do, destroy target artifact or enchantment."
//
//	Welcome Home {5}{G}{G}, Sorcery — Adventure:
//	"Create three 2/2 green Bear creature tokens."
//
// The adventure layout is Brazen Borrower's: the creature under the bare
// oracle ID, the Adventure under "<oracle_id>#1". The creature half is
// Gorbag of Minas Morgul's shape: a confirmation, the sacrifice of the
// creature itself, and a CR 603.12 reflexive trigger created only if it
// really was sacrificed, which chooses its target as it goes on the stack.
//
// No simplification.
func init() {
	const oracleID = "bacedc99-46d9-4757-8a27-8df77d7c2f02"

	effects.Register(effects.Spec{
		OracleID:     oracleID,
		Name:         "Flaxen Intruder",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer(
				"Flaxen Intruder — you may sacrifice it; when you do, destroy target artifact or enchantment",
				effects.Do(effects.MayChoice{
					Question: "Flaxen Intruder — sacrifice it to destroy target artifact or enchantment?",
					YesLabel: "Sacrifice",
					NoLabel:  "Decline",
					Then:     flaxenIntruderAnswer,
				})),
		},
	})

	effects.Register(effects.Spec{
		OracleID:     oracleID + "#1",
		Name:         "Welcome Home",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.CreateToken{Template: effects.TokenCard("2/2 green Bear"), N: 3}.Apply(ctx)
		},
	})
}

// flaxenIntruderAnswer is the "yes": sacrifice the creature itself. Keyed.
var flaxenIntruderAnswer = effects.MayYes("confirm/flaxen-intruder-sacrifice", func(ctx *effects.Context, _ game.ConfirmAnswer) error {
	if !effects.OnBattlefield(ctx.Game, ctx.Source()) {
		return nil
	}
	return effects.SacrificePermanent{Target: ctx.Source(), ThenTail: ctx.Tail(flaxenIntruderSacrificed)}.Apply(ctx)
})

// flaxenIntruderSacrificed is "when you do": only a creature that really was
// sacrificed. An on-disk key: never renamed, never reused.
var flaxenIntruderSacrificed = effects.CardTail("then/flaxen-intruder-sacrificed", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	return effects.WhenYouDo("Flaxen Intruder — destroy target artifact or enchantment", flaxenIntruderDestroyBody).Apply(ctx)
})

// flaxenIntruderDestroyBody is the reflexive trigger. An on-disk key: never
// renamed, never reused.
var flaxenIntruderDestroyBody = game.ReflexiveBody("flaxen-intruder/destroy", effects.SimpleBody(effects.DestroyFirstLegalCardTarget),
	effects.ConstTargets(func() *game.TargetSpec {
		return effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment()))
	}))
