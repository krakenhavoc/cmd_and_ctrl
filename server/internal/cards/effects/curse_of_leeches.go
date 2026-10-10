package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Curse of Leeches // Leeching Lurker — {2}{B} Enchantment — Aura Curse //
// Creature — Leech Horror 4/4 (#2586, ADR 0132, ADR 0079's amendment of
// 2026-10-09):
//
//	Front: "Enchant player
//	        As this permanent transforms into Curse of Leeches, attach it
//	        to a player.
//	        At the beginning of enchanted player's upkeep, they lose 1
//	        life and you gain 1 life.
//	        Daybound (If a player casts no spells during their own turn,
//	        it becomes night next turn.)"
//	Back:  "Lifelink
//	        Nightbound (If a player casts at least two spells during their
//	        own turn, it becomes day next turn.)"
//
// Three ways onto the battlefield, one rule each:
//
//   - Cast as the Aura by day: an ordinary "enchant player" cast. The
//     player is a target (Spec.Targets), announced and re-checked.
//   - Cast at night, or on the battlefield when night falls: it is (or
//     becomes) Leeching Lurker. An Aura that turns into a creature is
//     no longer attached to anything (CR 704.5p's creature sentence,
//     attachmentLegalLocked), so the Curse lets go of its player the
//     moment it turns over.
//   - Turned back over when day returns: the front face's
//     AsTransformsInto hook asks the controller which player to enchant
//     (QueueAttachSourceToPlayerForEffect). Any player, themselves
//     included; it is a choice, not a target. If no player can be
//     enchanted, or the question is dropped unanswered, the Aura is
//     attached to nothing and CR 704.5m puts it in the graveyard.
//
// A Curse that ENTERS transformed never transformed, so its "as this
// transforms" clause does not run: that is the Lurker, which needs no
// player.
//
// The upkeep trigger reads the enchanted player off the event, as Curse
// of Fool's Wisdom does: the player whose upkeep it is, when that is the
// player the Curse is attached to. Their life loss and the controller's
// gain are not linked ("they lose 1 life and you gain 1 life"), so the
// gain happens whether or not the loss went through.
//
// No simplification.
func init() {
	const oracle = "44eb0caa-ba16-49e4-915c-bd5e1ce770e6"
	Register(Spec{
		OracleID:        oracle,
		Name:            "Curse of Leeches",
		Completeness:    CompletenessFull,
		Targets:         EnchantPlayer(),
		PrintedKeywords: []string{"daybound"},
		AsTransformsInto: func(card *game.Card, ctx *Context) error {
			ctx.Game.QueueAttachSourceToPlayerForEffect(card.InstanceID,
				"Curse of Leeches — attach it to a player")
			return nil
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginUpkeep, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return source.AttachedTo.Kind == game.TargetPlayer && source.AttachedTo.ID == ev.Actor
			}, "Curse of Leeches — they lose 1 life and you gain 1 life",
				func(g *game.Game, item *game.StackItem) error {
					victim := item.Trigger.Event.Actor
					if g.PlayerByIDForEffect(victim) != nil {
						if err := g.ChangePlayerLifeForEffect(item.SourceCardID, victim, -1); err != nil {
							return err
						}
					}
					return GainLife{Player: item.Controller, Amount: 1}.Apply(NewContext(g, item))
				}),
		},
	})
	Register(Spec{
		OracleID:        oracle + "#1",
		Name:            "Leeching Lurker",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink", "nightbound"},
	})
}
