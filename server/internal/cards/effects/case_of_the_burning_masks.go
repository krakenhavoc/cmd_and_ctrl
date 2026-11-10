package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Case of the Burning Masks — Enchantment — Case {1}{R}{R} (Murders at
// Karlov Manor, #4469):
//
//	"When this Case enters, it deals 3 damage to target creature an
//	 opponent controls.
//	 To solve — Three or more sources you controlled dealt damage this
//	 turn. (If unsolved, solve at the beginning of your end step.)
//	 Solved — Sacrifice this Case: Exile the top three cards of your
//	 library. Choose one of them. You may play that card this turn."
//
// The Case (ADR 0071) is Case of the Filched Falcon's skeleton. The solve
// condition counts the distinct source OBJECTS its controller controlled
// as they dealt damage this turn (Game.DamageSourcesDealtThisTurnForEffect,
// damage_sources_tally.go): a permanent that deals damage twice is one
// source, a source that died or changed hands still counted for who
// controlled it as it dealt the damage, and a card that changed zones is a
// new object. The condition is read at the end step and again on
// resolution (CR 603.4).
//
// The payoff is Heroes' Hangout's exile-then-choose: all three cards are
// exiled, the controller chooses one, and only that card may be played
// (a land is played, a spell cast, paying its costs) until end of turn.
// The sacrifice is a cost, so the exile still happens with the Case gone.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7c1eb874-97f3-463f-bf4e-4d82dcc16e69",
		Name:         "Case of the Burning Masks",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Case of the Burning Masks — it deals 3 damage to target creature an opponent controls", DamageToFirstTarget(3)),
				TargetCreature("target creature an opponent controls", OpponentControls())),
			ToSolve("Case of the Burning Masks — to solve: three or more sources you controlled dealt damage this turn", threeOrMoreSourcesDealtDamage),
		},
		Activated: []ActivatedAbility{{
			Label:      "Sacrifice this Case: Exile the top three cards of your library. Choose one of them. You may play that card this turn.",
			Purpose:    game.Purpose{Answers: game.AnswerValue},
			Cost:       SacrificeThis(),
			ActiveWhen: Solved(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				ids := g.LookAtTopOfLibraryForEffect(item.Controller, 3)
				if len(ids) == 0 {
					return nil
				}
				return g.ExileCardsKeyedForEffect(ids, ctx.ItemTail(burningMasksChoose))
			},
		}},
	})
}

// threeOrMoreSourcesDealtDamage is the Case's solve condition.
func threeOrMoreSourcesDealtDamage(g *game.Game, controller, _ uuid.UUID) bool {
	return g.DamageSourcesDealtThisTurnForEffect(controller) >= 3
}

// burningMasksChoose is "choose one of them" over the cards that reached
// exile. An on-disk key.
var burningMasksChoose = CardTail("then/case-of-the-burning-masks-choose", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  ctx.Controller(),
		Source:   ctx.Source(),
		Question: "Case of the Burning Masks — choose one of them (you may play it this turn)",
		Cards:    o.IDs,
		Min:      1,
		Max:      1,
		LiveZone: game.ZoneExile,
		ThenKey:  burningMasksGrant,
	})
	return nil
})

// burningMasksGrant gives the chosen card its window: its chooser may play
// it this turn. An on-disk key.
var burningMasksGrant = game.RegisterChooseCardsThen("choose-cards/case-of-the-burning-masks-play", func(g *game.Game, r game.CardsChosen) error {
	for _, id := range r.Picked {
		g.GrantCastPermissionOverCardForEffect(id, game.CastPermission{
			Player:   r.Chooser,
			Zone:     game.ZoneExile,
			Duration: g.UntilEndOfTurnDuration(),
			Label:    "Case of the Burning Masks — play it this turn",
		})
	}
	return nil
})
