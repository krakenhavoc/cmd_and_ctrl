package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Territory Forge — Artifact {4}{R}:
//
//	"When this artifact enters, if you cast it, exile target artifact
//	 or land.
//	 This artifact has all activated abilities of the exiled card."
//
// "If you cast it" is the intervening-if b16EnteredFromStack answers (a
// permanent put onto the battlefield, or stolen on the stack, does not
// trigger). The exile is linked (ExileTarget.Linked, ADR 0155 §8), so
// "the exiled card" is the card linked to this Forge object, and the
// borrow lends its abilities (ExiledWithLenders). A land lends its mana
// abilities, a basic one through its land type (CR 305.6). A token
// exiled this way ceases to exist and lends nothing.
//
// No simplification.
func init() {
	const clause = "This artifact has all activated abilities of the exiled card."
	borrow := effects.HasActivatedAbilitiesOf(game.AbilityBorrow{
		Label:   clause,
		Lenders: effects.ExiledWithLenders(effects.AnyLender),
	})
	borrow.DependsOnExile = true
	effects.Register(effects.Spec{
		OracleID:     "a14d89b6-23eb-4337-8dee-93c39b0efbf1",
		Name:         "Territory Forge",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{borrow},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
				return effects.B06SelfETB(ev, source, ch, g) && effects.B16EnteredFromStack(g, source.InstanceID)
			},
			Targets: effects.TargetPermanent("target artifact or land", effects.Or(effects.Artifact(), effects.Land())),
			Key:     "Territory Forge — exile target artifact or land",
			Effect:  effects.ExileLinkedTargetsEffect,
		}},
	})
}
