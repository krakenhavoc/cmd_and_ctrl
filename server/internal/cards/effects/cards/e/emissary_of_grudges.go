package cardse

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Emissary of Grudges — Creature — Efreet {5}{R}, 6/5 (Commander 2018, S82 OLD
// batch 193, #5526):
//
//	"Flying, haste
//	 As this creature enters, secretly choose an opponent.
//	 Reveal the player you chose: Choose new targets for target spell or
//	 ability if it's controlled by the chosen player and if it targets you or
//	 a permanent you control. Activate only once."
//
// The choice is Guardian Archon's: made as the permanent enters (CR 614.12),
// stored on it and hidden from every seat but its chooser until the reveal,
// which is the activation's cost (CR 602.2b), so only the chooser can pay it.
// "Activate only once" is spent by the activation. The two conditions are on
// the instruction, not on the target (CR 608.2c): any spell or ability may be
// targeted, and as the ability resolves nothing happens to it unless the chosen
// player controls it and at least one of its targets is you or a permanent you
// control. Then you choose new targets for it (CR 115.7c: every target may
// change, any may stay as it is), and it is not optional.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b415684e-40f0-43c1-b8d5-1a90b75c5a47",
		Name:            "Emissary of Grudges",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		AsEnters: func(card *game.Card, ctx *effects.Context) error {
			ctx.Game.QueueSecretPlayerChoiceForEffect(card.Controller, card.InstanceID,
				"Emissary of Grudges — secretly choose an opponent (only you see the answer until you reveal it)",
				effects.Opponents.Seats(ctx, card.Controller))
			return nil
		},
		Activated: []effects.ActivatedAbility{{
			Label:     "Reveal the player you chose: Choose new targets for target spell or ability if it's controlled by the chosen player and if it targets you or a permanent you control. Activate only once.",
			Cost:      effects.RevealSecretChoice(),
			Condition: effects.YouSecretlyChoseForThis,
			OnlyOnce:  true,
			Targets:   effects.TargetSpellOrAbility("target spell or ability"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				chosen, _ := g.SecretChosenPlayerOf(item.SourceCardID)
				if chosen == uuid.Nil {
					return nil
				}
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					victim := g.StackItemForEffect(t.ID)
					if victim == nil || victim.Controller != chosen || !emissaryTargetsYouOrYours(g, victim, item.Controller) {
						return nil
					}
				}
				return effects.RetargetTheTargetedItem(ctx, "Emissary of Grudges — choose new targets", game.RetargetChooseNew, false)
			},
		}},
	})
}

// emissaryTargetsYouOrYours reports whether any target of the stack item is
// `you` or a permanent `you` control.
func emissaryTargetsYouOrYours(g *game.Game, victim *game.StackItem, you uuid.UUID) bool {
	for _, t := range victim.Targets {
		switch t.Kind {
		case game.TargetPlayer:
			if t.ID == you {
				return true
			}
		case game.TargetCard:
			if c, ok := g.LookupCardForEffect(t.ID); ok && g.Battlefield.Contains(t.ID) && c.Controller == you {
				return true
			}
		}
	}
	return false
}
