package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Smirking Spelljacker — Creature — Djinn Wizard Rogue {4}{U}, 3/3
// (Outlaws of Thunder Junction Commander):
//
//	"Flash
//	 Flying
//	 When this creature enters, exile target spell an opponent controls.
//	 Whenever this creature attacks, if a card is exiled with it, you may
//	 cast the exiled card without paying its mana cost."
//
// The two abilities are linked (CR 607.2a). The enters trigger exiles the
// spell without countering it, so a spell that can't be countered is
// exiled all the same, linked to this OBJECT
// (game.ExileSpellLinkedKeyedForEffect): a Spelljacker that left and came
// back is a new object with no claim on the card (CR 400.7). The attack
// trigger carries an intervening "if" (CR 603.4), read as it triggers and
// again as it resolves, so a Spelljacker with nothing exiled with it does
// not trigger. The free cast is part of the resolution (CR 608.2g,
// Transcendent Dragon's CastExiledFreeKeepingForEffect): {0} with X as 0,
// the card's own timing ignored, and declining leaves it in exile to be
// offered again the next time it attacks.
//
// No simplification.
func init() {
	exile := effects.Targeting(
		effects.WhenThisEnters("Smirking Spelljacker — exile target spell an opponent controls",
			func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				ref, ok := ctx.SourceRef()
				if !ok {
					return nil
				}
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetCard {
						return g.ExileSpellLinkedKeyedForEffect(ref, t.ID, game.KeyedTail{})
					}
				}
				return nil
			}),
		effects.TargetSpell("target spell an opponent controls", effects.OpponentControls()))
	attack := effects.InterveningIf(effects.Cmdr2S36CondCardExiledWithThis{},
		effects.On(game.EventAttack, effects.Cmdr2S36ThisAttackedWithAnExiledCard,
			"Smirking Spelljacker — you may cast the exiled card without paying its mana cost",
			effects.Cmdr2S36CastExiledWithThisFree))
	effects.Register(effects.Spec{
		OracleID:        "b23c28ec-dc4d-4072-8fd8-6640a104335a",
		Name:            "Smirking Spelljacker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Triggered:       []game.TriggeredAbility{exile, attack},
	})
}
