package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Myth Realized — Enchantment {W} (S82 GAP, #5948):
//
//	"Whenever you cast a noncreature spell, put a lore counter on this
//	 enchantment.
//	 {2}{W}: Put a lore counter on this enchantment.
//	 {W}: Until end of turn, this enchantment becomes a Monk Avatar creature in
//	 addition to its other types and gains "This creature's power and toughness
//	 are each equal to the number of lore counters on it.""
//
// Beorn's Hospitality's animate-with-a-granted-static shape, with the
// until-end-of-turn duration the default. The granted power/toughness static
// sits in layer 7b (the layer a granted static can occupy, ADR 0093 Decision
// 10) rather than a printed characteristic-defining ability's 7a.
func init() {
	const grantKey = "myth-realized/power-toughness-lore-counters"
	const text = "This creature's power and toughness are each equal to the number of lore counters on it."
	effects.Register(effects.Spec{
		OracleID:     "3a726e40-829c-47ec-81a7-58a04164fc3f",
		Name:         "Myth Realized",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: grantKey,
			Static: []game.StaticAbility{{
				Label:    text,
				Layer:    game.Layer7PT,
				SubLayer: game.SubLayer7B_Set,
				AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return source != nil && target.InstanceID == source.InstanceID
				},
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
					n := source.Counters[game.CounterLore]
					c.Power, c.Toughness = n, n
				},
			}},
			Text: text,
		}},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.B10NoncreatureSpellCastByYou,
				"Myth Realized — put a lore counter on this enchantment",
				mythRealizedLore),
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{2}{W}: Put a lore counter on this enchantment.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.ManaCost("{2}{W}"),
				Effect:  mythRealizedLore,
			},
			{
				Label:   "{W}: Until end of turn, this enchantment becomes a Monk Avatar creature in addition to its other types and gains \"" + text + "\"",
				Purpose: game.Purpose{Answers: game.AnswerAnimate},
				Cost:    effects.ManaCost("{W}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					self := item.SourceCardID
					if !effects.OnBattlefield(g, self) || effects.SourceIsNewObject(g, item) {
						return nil
					}
					return effects.GrantAbilitiesFor{
						Target: self,
						Keys:   []string{grantKey},
						Also: []game.Mod{
							game.AddTypesMod("Creature"),
							game.AddSubtypesMod(g.TextReaderForEffect(item).TextSubtypes("Monk", "Avatar")...),
						},
						Label: "Myth Realized — becomes a Monk Avatar creature until end of turn",
					}.Apply(effects.NewContext(g, item))
				},
			},
		},
	})
}

// mythRealizedLore is "put a lore counter on this enchantment".
func mythRealizedLore(g *game.Game, item *game.StackItem) error {
	return effects.AddCounter{Target: item.SourceCardID, Kind: game.CounterLore, N: 1}.Apply(effects.NewContext(g, item))
}
