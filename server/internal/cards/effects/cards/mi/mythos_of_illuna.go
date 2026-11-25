package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mythos of Illuna — Sorcery {2}{U}{U} (Ikoria: Lair of Behemoths, #5512):
//
//	"Create a token that's a copy of target permanent. If {R}{G} was spent
//	 to cast this spell, instead create a token that's a copy of that
//	 permanent, except the token has "When this token enters, if it's a
//	 creature, it fights up to one target creature you don't control.""
//
// Aggressive Biomancy's shape: the copy carries a granted ability (the
// catalog's Grants), here an intervening-"if" enters trigger that only fires
// when the copy is a creature (CR 603.4), with its "up to one" target chosen
// as it goes on the stack. The mana spent is read as the spell resolves.
//
// No simplification.
func init() {
	itsACreature := func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
		return source != nil && source.HasCardType("creature")
	}
	fights := effects.Targeting(
		effects.InterveningIf(
			effects.CondBoard{Name: "it's a creature", Test: func(ctx *effects.Context) bool {
				c, ok := ctx.Game.LookupCardForEffect(ctx.Source())
				return ok && c.HasCardType("creature")
			}},
			effects.On(game.EventETB, effects.AllOf(effects.Self, itsACreature),
				"Mythos of Illuna — it fights up to one target creature you don't control", effects.ItFightsTheFirstTarget)),
		effects.TargetCreature("up to one target creature you don't control", effects.OpponentControls()).WithCount(0, 1))
	effects.Register(effects.Spec{
		OracleID:       "42b10f0f-19cd-45d2-a2db-d5463ddc5e5b",
		Name:           "Mythos of Illuna",
		Completeness:   effects.CompletenessFull,
		ReadsManaSpent: game.ReadsOwnPayment,
		WantsColors:    []string{"R", "G"},
		Targets:        effects.TargetPermanent("target permanent"),
		Grants: []effects.AbilityGrant{{
			Key:       mythosOfIllunaGrant,
			Triggered: []game.TriggeredAbility{fights},
			Text:      "When this token enters, if it's a creature, it fights up to one target creature you don't control.",
		}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			c := effects.CreateTokenCopy{Controller: ctx.Controller(), Copy: id, N: 1}
			if ctx.ManaSpentOfColor("R") > 0 && ctx.ManaSpentOfColor("G") > 0 {
				c.Except = func(tok *game.Card) {
					v := game.PrintedValues{TypeLine: tok.TypeLine, GrantedAbilities: tok.GrantedAbilities}
					v.GrantAbility(mythosOfIllunaGrant)
					tok.GrantedAbilities = v.GrantedAbilities
				}
			}
			return c.Apply(ctx)
		},
	})
}

const mythosOfIllunaGrant = "mythos-of-illuna/token-fights"
