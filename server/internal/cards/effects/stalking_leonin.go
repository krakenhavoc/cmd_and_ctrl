package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Stalking Leonin — Creature — Cat Archer {2}{W}, 3/3 (Murders at Karlov
// Manor Commander):
//
//	"When this creature enters, secretly choose an opponent.
//	 Reveal the player you chose: Exile target creature that's attacking
//	 you if it's controlled by the chosen player. Activate only once."
//
// The choice is stored on the permanent, visible to nobody until the
// reveal (the same secret choice A Killer Among Us uses). Revealing is the
// cost, so the ability can be activated only by the player who chose, and
// only while a choice exists. The target is any creature attacking you
// (CR 506.3: not one attacking your planeswalker); on resolution it is
// exiled only if the chosen player controls it. "Activate only once"
// spends the ability even when the target was not the chosen player's.
func init() {
	Register(Spec{
		OracleID:     "6ec98a34-5c04-423c-98e5-8cf3047135c1",
		Name:         "Stalking Leonin",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Stalking Leonin — secretly choose an opponent", func(g *game.Game, item *game.StackItem) error {
				g.QueueSecretPlayerChoiceForEffect(item.Controller, item.SourceCardID,
					"Stalking Leonin — secretly choose an opponent (only you see the answer until you reveal it)",
					NewContext(g, item).Opponents())
				return nil
			}),
		},
		Activated: []ActivatedAbility{{
			Label:     "Reveal the player you chose: Exile target creature that's attacking you if it's controlled by the chosen player. Activate only once.",
			Purpose:   game.Purpose{Answers: game.AnswerRemove},
			Cost:      RevealSecretChoice(),
			Condition: YouSecretlyChoseForThis,
			OnlyOnce:  true,
			Targets:   TargetCreature("target creature that's attacking you", AttackingYou()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				ts := ctx.LegalTargets()
				chosen, _ := g.SecretChosenPlayerOf(item.SourceCardID)
				if len(ts) == 0 || chosen == uuid.Nil {
					return nil
				}
				c, ok := g.LookupCardForEffect(ts[0].ID)
				if !ok || c.Controller != chosen {
					return nil
				}
				return ExileTarget{Target: ts[0].ID}.Apply(ctx)
			},
		}},
	})
}
