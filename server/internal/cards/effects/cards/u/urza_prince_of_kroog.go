package cardsu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Urza, Prince of Kroog — Legendary Creature — Human Artificer {2}{W}{U},
// 2/3 (The Brothers' War):
//
//	"Artifact creatures you control get +2/+2.
//	 {6}: Create a token that's a copy of target artifact you control,
//	 except it's a 1/1 Soldier creature in addition to its other types."
//
// The first line is a layer 7c anthem over the controller's artifact
// creatures, Urza himself not included unless he is one (he isn't). The
// second is a token copy of the target's copiable values as it is on
// resolution (CR 707.2) with the exception applied (CR 707.9): base power
// and toughness 1/1 (a printed 1/1, not a `*` stand-in), and the creature
// card type and the Soldier subtype added to its own. A copy of a
// creature keeps that creature's other subtypes. Not a target any more:
// the ability does nothing (CR 608.2b).
//
// No simplification.
func init() {
	anthem := effects.B16Anthem(func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return target.Controller == source.Controller && target.IsArtifact() && target.IsCreature()
	}, 2, 2)
	anthem.Label = "Artifact creatures you control get +2/+2."
	effects.Register(effects.Spec{
		OracleID:     "6a448862-e8d8-4a08-8e3d-5460afa29b4d",
		Name:         "Urza, Prince of Kroog",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{anthem},
		Activated: []effects.ActivatedAbility{{
			Label:   "{6}: Create a token that's a copy of target artifact you control, except it's a 1/1 Soldier creature in addition to its other types.",
			Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1},
			Cost:    effects.ManaCost("{6}"),
			Targets: effects.TargetPermanent("target artifact you control", effects.Artifact(), effects.YouControl()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id := effects.FirstLegalBattlefieldTarget(ctx)
				if id == uuid.Nil {
					return nil
				}
				return effects.CreateTokenCopy{
					Controller: item.Controller,
					Copy:       id,
					N:          1,
					Except:     tokenCopyIsASoldier,
				}.Apply(ctx)
			},
		}},
	})
}

// tokenCopyIsASoldier is "except it's a 1/1 Soldier creature in addition
// to its other types" (CR 707.9b): base 1/1, the creature type and the
// Soldier subtype added.
func tokenCopyIsASoldier(tok *game.Card) {
	tok.Power, tok.Toughness = 1, 1
	tok.VariableToughness = false
	tok.PrintedPTKnown = true
	tok.TypeLine = effects.SocB19WithCreatureAndSubtypeInAddition(tok.TypeLine, "Soldier")
}
