package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Summon: Primal Odin — Enchantment Creature — Saga Knight {4}{B}{B},
// 5/3 (Final Fantasy, #4462):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I — Gungnir — Destroy target creature an opponent controls.
//	 II — Zantetsuken — This creature gains 'Whenever this creature
//	 deals combat damage to a player, that player loses the game.'
//	 III — Hall of Sorrow — Draw two cards. Each player loses 2 life."
//
// A Saga creature. Chapter II is Urza's Saga's self-grant (ADR 0093):
// the bundle is pinned to this object with no end, so it lasts until
// the Saga leaves (CR 611.2a, 400.7), and a Saga that left and came
// back before the chapter resolved gains nothing. The granted trigger
// is Frodo's: the damaged player loses the game. Chapter III draws,
// then every player, you included, loses 2 life (Defacing Duskmage's
// prepare spell).
//
// No simplification.
func init() {
	const name = "Summon: Primal Odin"
	effects.Register(effects.Spec{
		OracleID:     "032f59c3-9f47-4db3-a43c-065a99ed7efe",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: summonPrimalOdinZantetsuken,
			Triggered: []game.TriggeredAbility{
				effects.WheneverThisDealsCombatDamageToAPlayer(name+" — Zantetsuken: that player loses the game",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						return effects.LoseTheGame{Player: ctx.Trigger().Event.Target}.Apply(ctx)
					}),
			},
			Text: "Whenever this creature deals combat damage to a player, that player loses the game.",
		}},
		Triggered: []game.TriggeredAbility{
			effects.ChapterTriggerTargeting(1, effects.SagaChapterLabel(name, 1, "Gungnir — destroy target creature an opponent controls"),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls()),
				effects.DestroyFirstLegalCardTarget),
			effects.ChapterTrigger(2, effects.SagaChapterLabel(name, 2, "Zantetsuken — this creature gains \"combat damage to a player: that player loses the game\""),
				func(g *game.Game, item *game.StackItem) error {
					if effects.SourceIsNewObject(g, item) {
						return nil
					}
					self := item.SourceCardID
					return effects.GrantAbilitiesFor{
						Target:   self,
						Keys:     []string{summonPrimalOdinZantetsuken},
						Duration: g.PinnedTo(game.IndefiniteDuration(), self),
						Label:    name + " — Zantetsuken",
					}.Apply(effects.NewContext(g, item))
				}),
			effects.ChapterTrigger(3, effects.SagaChapterLabel(name, 3, "Hall of Sorrow — draw two cards; each player loses 2 life"),
				effects.Do(effects.DrawCards{N: 2}, effects.Instruction(effects.EachPlayerLosesTwoLife))),
		},
	})
}

// summonPrimalOdinZantetsuken is chapter II's granted bundle. An on-disk
// key: never renamed, never reused.
const summonPrimalOdinZantetsuken = "summon-primal-odin/zantetsuken"
