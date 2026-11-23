package cardssh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sigil of Myrkul — Enchantment {2}{B} (Commander Legends: Battle for
// Baldur's Gate, #5948):
//
//	"At the beginning of combat on your turn, mill a card. When you do, if
//	 there are four or more creature cards in your graveyard, put a +1/+1
//	 counter on target creature you control and it gains deathtouch until end
//	 of turn."
//
// Primal Witchstalker's mill-then-reflexive shape: with an empty library
// nothing is milled and the reflexive trigger (CR 603.12) does not happen,
// which is read before the mill. The reflexive trigger has an intervening
// "if" (CR 603.4): the graveyard is counted when it would trigger, after the
// mill, and again as it resolves. It picks its creature as it goes on the
// stack; a target gone by resolution gets neither the counter nor deathtouch.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ae7e14cf-a275-4260-8d78-324429a042b7",
		Name:         "Sigil of Myrkul",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtBeginningOfYourCombat("Sigil of Myrkul — mill a card", effects.Do(
				effects.Instruction(func(ctx *effects.Context) error {
					p := ctx.Game.PlayerByIDForEffect(ctx.Controller())
					if p == nil || p.Library == nil || p.Library.Size() == 0 {
						return nil
					}
					ctx.RecordStep(sigilOfMyrkulMilled, nil, 1)
					return effects.MillCards{N: 1}.Apply(ctx)
				}),
				effects.Instruction(func(ctx *effects.Context) error {
					if ctx.StepRecord(sigilOfMyrkulMilled).N() == 0 || !sigilOfMyrkulFourCreatures(ctx.Game, ctx.Controller()) {
						return nil
					}
					return effects.WhenYouDo("Sigil of Myrkul — put a +1/+1 counter on target creature you control and it gains deathtouch until end of turn",
						sigilOfMyrkulBoostBody).Apply(ctx)
				}),
			)),
		},
	})
}

// sigilOfMyrkulMilled files that the library had a card to mill.
const sigilOfMyrkulMilled = "sigil-of-myrkul/milled"

// sigilOfMyrkulFourCreatures is "there are four or more creature cards in
// your graveyard".
func sigilOfMyrkulFourCreatures(g *game.Game, player uuid.UUID) bool {
	p := g.PlayerByIDForEffect(player)
	if p == nil || p.Graveyard == nil {
		return false
	}
	n := 0
	for _, c := range p.Graveyard.Cards {
		if c.HasCardType("creature") {
			n++
		}
	}
	return n >= 4
}

// sigilOfMyrkulBoostBody is the reflexive trigger. An on-disk key: never
// renamed, never reused.
var sigilOfMyrkulBoostBody = game.ReflexiveBody("sigil-of-myrkul/boost",
	func(g *game.Game, item *game.StackItem, _ game.EffectParams) error {
		ctx := effects.NewContext(g, item)
		if !sigilOfMyrkulFourCreatures(g, item.Controller) {
			return nil
		}
		id, ok := effects.B16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		if err := (effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
			return err
		}
		return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"deathtouch"}, Label: "Sigil of Myrkul — deathtouch"}.Apply(ctx)
	},
	effects.ConstTargets(func() *game.TargetSpec {
		return effects.TargetCreature("target creature you control", effects.YouControl())
	}))
