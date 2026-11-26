package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// General Kudro of Drannith — Legendary Creature — Human Soldier {1}{W}{B},
// 3/3 (Ikoria: Lair of Behemoths, #5512):
//
//	"Other Humans you control get +1/+1.
//	 Whenever General Kudro or another Human you control enters, exile target
//	 card from an opponent's graveyard.
//	 {2}, Sacrifice two Humans: Destroy target creature with power 4 or
//	 greater."
//
// The anthem is Jirina Kudro's layer 7c modify over the other Humans you
// control. The enters trigger is one trigger for the General and every Human
// you control (Breathless Knight's shape, the Human read through the layered
// view as it enters), its target a card in a graveyard you don't own (Ruin
// Rat's), removed when there is none (CR 603.3d). The sacrifice is part of
// the cost (CR 602.2b) and may include the General himself; the target is
// chosen before it is paid, so a sacrificed Human can't be the target of an
// ability whose cost took it.
//
// No simplification.
func init() {
	trigger := effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		c, ok := effects.EnteredUnderYourControl(ev, source, g, false)
		return ok && c.IsCreature() && c.HasSubtype(source.TextSubtype("Human"))
	}, "General Kudro of Drannith — exile target card from an opponent's graveyard", effects.ExileTheFirstLegalTarget)
	trigger.Targets = effects.TargetCardInGraveyard("target card from an opponent's graveyard", effects.Not(effects.YouOwn()))
	effects.Register(effects.Spec{
		OracleID:     "a5a024a5-16a8-4083-9157-5905c5e96d76",
		Name:         "General Kudro of Drannith",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "Other Humans you control get +1/+1.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: effects.B16OtherCreaturesYouControlOfSubtype("Human"),
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power++
				c.Toughness++
			},
		}},
		Triggered: []game.TriggeredAbility{trigger},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, Sacrifice two Humans: Destroy target creature with power 4 or greater.",
			Purpose: game.Purpose{Answers: game.AnswerRemove | game.AnswerSacOutlet},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.SacrificeN(2, "two Humans", effects.HasSubtype("Human"))),
			Targets: effects.TargetCreature("target creature with power 4 or greater", effects.PowerGE(4)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.DestroyTheTargetPermanent(item, effects.NewContext(g, item))
			},
		}},
	})
}
