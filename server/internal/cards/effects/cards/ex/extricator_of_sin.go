package cardsex

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// extricatorOfSinOracle is shared by both faces: the front registers under
// it, the back under extricatorOfSinOracle + "#1".
const extricatorOfSinOracle = "08b3328c-1d96-4a05-ae8b-f1b654084faa"

// Extricator of Sin // Extricator of Flesh — a transforming creature (Eldritch
// Moon, S82 OLD, #5541):
//
//	Extricator of Sin {2}{W}, Creature — Human Cleric, 0/3
//	  "When this creature enters, you may sacrifice another permanent. If you
//	   do, create a 3/2 colorless Eldrazi Horror creature token.
//	   Delirium — At the beginning of your upkeep, if there are four or more
//	   card types among cards in your graveyard, transform this creature."
//
//	Extricator of Flesh, Creature — Eldrazi Horror, 3/5
//	  "Eldrazi you control have vigilance.
//	   {2}, {T}, Sacrifice a non-Eldrazi creature: Create a 3/2 colorless
//	   Eldrazi Horror creature token."
//
// The enters trigger asks at resolution (Beetle-Headed Merchants' "you may
// sacrifice, if you do" shape, CR 701.21). The transform is an intervening
// "if" (CR 603.4): checked as upkeep begins and again as the trigger
// resolves, turning over in place only if this is still the same permanent
// (CR 701.27f). "Eldrazi you control" is every permanent with the subtype,
// read through the source's text (CR 612.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     extricatorOfSinOracle,
		Name:         "Extricator of Sin",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Extricator of Sin — you may sacrifice another permanent; if you do, create a 3/2 colorless Eldrazi Horror creature token",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MaySacrificeOne(g, item, "Extricator of Sin — sacrifice another permanent to create a 3/2 Eldrazi Horror?",
						effects.Permanent(), true, extricatorOfSinSacrificed)
				}),
			effects.On(game.EventBeginUpkeep, effects.AllOf(effects.ByYou, effects.YouHaveDeliriumWhen),
				"Extricator of Sin — Delirium: transform this creature",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if !effects.DrcDeliriumMet(g, ctx.Controller()) {
						return nil // CR 603.4: checked again as it resolves
					}
					return effects.TransformThis{}.Apply(ctx)
				}),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     extricatorOfSinOracle + "#1",
		Name:         "Extricator of Flesh",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{effects.Labelled(effects.KeywordGrant(
			func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.Controller == source.Controller && target.HasSubtype(source.TextSubtype("Eldrazi"))
			}, "vigilance"), "Eldrazi you control have vigilance.")},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}, Sacrifice a non-Eldrazi creature: Create a 3/2 colorless Eldrazi Horror creature token.",
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost(), game.AbilityCost{SacrificeOther: effects.SacrificeSpec("a non-Eldrazi creature", effects.Creature(), effects.Not(effects.HasSubtype("Eldrazi")))}),
			Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1, Answers: game.AnswerSacOutlet | game.AnswerMakesBlocker},
			Effect:  effects.Do(effects.CreateToken{Template: effects.TokenCard("3/2 colorless Eldrazi Horror"), N: 1}),
		}},
	})
}

var extricatorOfSinSacrificed = effects.RegisterSacrificedOne("prompt-run/extricator-of-sin-horror", func(ctx *effects.Context, _ uuid.UUID) error {
	return effects.CreateToken{Template: effects.TokenCard("3/2 colorless Eldrazi Horror"), N: 1}.Apply(ctx)
})
