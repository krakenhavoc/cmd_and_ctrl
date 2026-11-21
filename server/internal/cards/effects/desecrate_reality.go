package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Desecrate Reality — Instant {7} (Commander Masters, #5948):
//
//	"For each opponent, exile up to one target permanent that player
//	 controls with an even mana value. (Zero is even.)
//	 Adamant — If at least three colorless mana was spent to cast this
//	 spell, return a permanent card with an odd mana value from your
//	 graveyard to the battlefield."
//
// One "up to one" clause per opponent in turn order, each bound to its own
// player (CR 608.2b). The adamant half (CR 207.2c) is a second resolution
// step, so it sees the graveyard after the exile; the card is chosen by the
// caster on resolution, not targeted. A waived payment reads as no mana
// spent, the weaker-than-printed direction (AdamantSpent).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c9cd57f6-1f69-4ea1-9672-ca944759c7e8",
		Name:         "Desecrate Reality",
		Completeness: CompletenessFull,
		// Adamant reads the mana spent to cast this spell (CR 207.2c).
		ReadsManaSpent: game.ReadsOwnPayment,
		// ADR 0126 §6: one permanent per opponent, not a sweep.
		Purpose: game.Purpose{Answers: game.AnswerRemove},
		Targets: perOpponentClauses(game.MaxPlayers-1, "permanent with an even mana value",
			func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.ManaValue()%2 == 0 }),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					if err := (ExileTarget{Target: t.ID}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			}),
			eldB1IfAdamantSpent("C", Instruction(desecrateRealityReturn)),
		),
	})
}

// desecrateRealityReturn asks for a permanent card with an odd mana value in
// the caster's graveyard and returns it to the battlefield.
func desecrateRealityReturn(ctx *Context) error {
	cards := gyCardIDs(ctx.Game, ctx.Controller(), func(c game.Card) bool {
		return !c.HasCardType("instant") && !c.HasCardType("sorcery") && c.ManaValue()%2 == 1
	})
	if len(cards) == 0 {
		return nil
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  ctx.Controller(),
		Source:   ctx.Source(),
		Question: "Desecrate Reality — return a permanent card with an odd mana value from your graveyard to the battlefield",
		Cards:    cards,
		Min:      1,
		Max:      1,
		LiveZone: game.ZoneGraveyard,
		ThenKey:  desecrateRealityPicked,
	})
	return nil
}

var desecrateRealityPicked = game.RegisterChooseCardsThen("choose-cards/desecrate-reality-return", func(g *game.Game, r game.CardsChosen) error {
	for _, id := range r.Picked {
		if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
			continue
		}
		if _, err := g.ReturnFromGraveyardWithCountersForEffect(id, uuid.Nil, false, nil); err != nil {
			return err
		}
	}
	return nil
})
