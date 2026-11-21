package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Living History — Enchantment {1}{R} (Secrets of Strixhaven, #4456):
//
//	"When this enchantment enters, create a 2/2 red and white Spirit
//	 creature token.
//	 Whenever you attack, if a card left your graveyard this turn, target
//	 attacking creature gets +2/+0 until end of turn."
//
// "Whenever you attack" is Bast's shape: one trigger per declaration
// (OncePerBatch), with the attacking creature chosen as the trigger goes on
// the stack and checked again as it resolves (CR 603.3d, 608.2b). The "if"
// is an intervening if (CR 603.4): a card of yours left your graveyard this
// turn (cardLeftYourGraveyardThisTurn, a cast from it included), read as
// the attack is declared, so a turn without one puts nothing on the stack
// and no target is chosen, and read again as the trigger resolves. It reads
// the whole turn, so a card that left before the Enchantment entered counts.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a76adc88-e7db-43e2-8fbf-bae95c0838ef",
		Name:         "Living History",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Living History — create a 2/2 red and white Spirit creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 red and white Spirit"), N: 1})),
			effects.OncePerBatch(effects.Targeting(
				effects.InterveningIf(effects.CardLeftYourGraveyardThisTurnCond,
					effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
						return effects.AttackDeclaredByYou(ev, source.Controller) && effects.CardLeftYourGraveyardThisTurn(g, source.Controller)
					}, "Living History — target attacking creature gets +2/+0 until end of turn",
						livingHistoryPump)),
				effects.TargetCreature("target attacking creature", effects.AttackingCreature()))),
		},
	})
}

// livingHistoryPump gives the target attacking creature +2/+0 until end of
// turn, if it is still a legal target.
func livingHistoryPump(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		return effects.BoostUntilEOT{Target: t.ID, Power: 2, Label: "Living History — +2/+0 until end of turn"}.Apply(ctx)
	}
	return nil
}
