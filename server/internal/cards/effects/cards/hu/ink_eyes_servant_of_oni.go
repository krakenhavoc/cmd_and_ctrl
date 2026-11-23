package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ink-Eyes, Servant of Oni — Legendary Creature — Rat Ninja {4}{B}{B}, 5/4
// (Bloomburrow Commander):
//
//	"Ninjutsu {3}{B}{B}
//	 Whenever Ink-Eyes deals combat damage to a player, you may put target
//	 creature card from that player's graveyard onto the battlefield under
//	 your control.
//	 {1}{B}: Regenerate Ink-Eyes."
//
// The target clause names the damaged player's graveyard (Fire Nation
// Salvagers' TargetsFrom), chosen as the trigger goes on the stack; "you
// may" is asked as it resolves (Ogre-Head Helm's MayChoice), and the card
// comes back under your control, not its owner's. The regeneration is
// Albino Troll's.
//
// No simplification.
func init() {
	damage := effects.WheneverThisDealsCombatDamageToAPlayer(
		"Ink-Eyes — you may put target creature card from that player's graveyard onto the battlefield under your control",
		cmdr2S45InkEyesAsk)
	damage.TargetsFrom = func(tc game.TriggerContext, _ *game.Card, _ *game.Game) *game.TargetSpec {
		return effects.TargetCardInGraveyard("target creature card from that player's graveyard",
			effects.Creature(), effects.OwnedBy(tc.Event.Target))
	}
	effects.Register(effects.Spec{
		OracleID:     "5d520476-740a-4005-801e-472b24fa6497",
		Name:         "Ink-Eyes, Servant of Oni",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{damage},
		Activated: []effects.ActivatedAbility{
			effects.Ninjutsu("{3}{B}{B}"),
			{
				Label:   "{1}{B}: Regenerate Ink-Eyes.",
				Purpose: game.Purpose{Answers: game.AnswerSurvive},
				Cost:    effects.ManaCost("{1}{B}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.Regenerate{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
				},
			},
		},
	})
}

func cmdr2S45InkEyesAsk(g *game.Game, item *game.StackItem) error {
	return effects.MayChoice{
		Question: "Ink-Eyes — put target creature card from that player's graveyard onto the battlefield under your control?",
		YesLabel: "Put it onto the battlefield",
		NoLabel:  "Decline",
		Then:     cmdr2S45InkEyesYes,
	}.Apply(effects.NewContext(g, item))
}

// cmdr2S45InkEyesYes is the "yes": the chosen creature card comes back under
// your control. Keyed (on disk): never renamed, never reused.
var cmdr2S45InkEyesYes = effects.MayYes("confirm/cmdr2-s45-ink-eyes-reanimate", func(ctx *effects.Context, _ game.ConfirmAnswer) error {
	return effects.MoseoInfusion(ctx.Game, ctx.Item)
})
