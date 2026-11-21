package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Splinter Twin — Enchantment — Aura {2}{R}{R} (Special Guests):
//
//	"Enchant creature
//	 Enchanted creature has "{T}: Create a token that's a copy of this
//	 creature, except it has haste. Exile that token at the beginning of
//	 the next end step.""
//
// The granted ability copies the creature it is on (the Aura is not part
// of its copiable values, CR 707.2), and the exile is a delayed trigger
// over the token it made (Saheeli Rai's shape).
func init() {
	const grant = "splinter-twin/copy"
	const text = "{T}: Create a token that's a copy of this creature, except it has haste. Exile that token at the beginning of the next end step."
	effects.Register(effects.Spec{
		OracleID:     "2df3187b-4b91-4de9-9b26-d48a30700cec",
		Name:         "Splinter Twin",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Grants: []effects.AbilityGrant{{
			Key: grant,
			Activated: []effects.ActivatedAbility{{
				Label:   text,
				Cost:    effects.TapCost(),
				Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1, Answers: game.AnswerValue},
				Effect:  splinterTwinCopy,
			}},
			Text: text,
		}},
		Static: []game.StaticAbility{effects.GrantAbilitiesToAttached(grant)},
	})
}

func splinterTwinCopy(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	cursor := effects.B25LastEventSeq(g)
	if err := (effects.CreateTokenCopy{
		Controller: item.Controller,
		Copy:       item.SourceCardID,
		N:          1,
		Except:     effects.TokenCopyGainsHaste,
	}).Apply(ctx); err != nil {
		return err
	}
	tokens := effects.B27TokensCreatedByAfter(g, item.Controller, cursor)
	if len(tokens) == 0 {
		return nil
	}
	return effects.ScheduleDelayedTrigger{
		At:         game.StepEnd,
		Controller: item.Controller,
		Label:      "Splinter Twin — exile the token",
		Cards:      tokens,
		Body:       effects.ExileListedCardsBody,
	}.Apply(ctx)
}
