package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "032f59c3-9f47-4db3-a43c-065a99ed7efe",
		Name:         name,
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: summonPrimalOdinZantetsuken,
			Triggered: []game.TriggeredAbility{
				WheneverThisDealsCombatDamageToAPlayer(name+" — Zantetsuken: that player loses the game",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						return LoseTheGame{Player: ctx.Trigger().Event.Target}.Apply(ctx)
					}),
			},
			Text: "Whenever this creature deals combat damage to a player, that player loses the game.",
		}},
		Triggered: []game.TriggeredAbility{
			ChapterTriggerTargeting(1, SagaChapterLabel(name, 1, "Gungnir — destroy target creature an opponent controls"),
				TargetCreature("target creature an opponent controls", OpponentControls()),
				destroyFirstLegalCardTarget),
			ChapterTrigger(2, SagaChapterLabel(name, 2, "Zantetsuken — this creature gains \"combat damage to a player: that player loses the game\""),
				func(g *game.Game, item *game.StackItem) error {
					if sourceIsNewObject(g, item) {
						return nil
					}
					self := item.SourceCardID
					return GrantAbilitiesFor{
						Target:   self,
						Keys:     []string{summonPrimalOdinZantetsuken},
						Duration: g.PinnedTo(game.IndefiniteDuration(), self),
						Label:    name + " — Zantetsuken",
					}.Apply(NewContext(g, item))
				}),
			ChapterTrigger(3, SagaChapterLabel(name, 3, "Hall of Sorrow — draw two cards; each player loses 2 life"),
				Do(DrawCards{N: 2}, Instruction(eachPlayerLosesTwoLife))),
		},
	})
}

// summonPrimalOdinZantetsuken is chapter II's granted bundle. An on-disk
// key: never renamed, never reused.
const summonPrimalOdinZantetsuken = "summon-primal-odin/zantetsuken"
