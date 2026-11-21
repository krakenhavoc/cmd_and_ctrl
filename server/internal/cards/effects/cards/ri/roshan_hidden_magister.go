package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Roshan, Hidden Magister — Legendary Creature — Human Assassin {3}{B},
// 4/4 (Assassin's Creed, #5948):
//
//	"Other creatures you control are Assassins in addition to their other
//	 types. The same is true for creature spells you control and creature
//	 cards you own that aren't on the battlefield.
//	 Face-down creatures you control have menace.
//	 Whenever a permanent you control is turned face up, you draw a card and
//	 you lose 1 life."
//
// Leyline of Transformation's three layer-4 statics with the type fixed at
// Assassin, the first excluding Roshan ("other"): creatures you control, creature spells you
// control and creature cards you own off the battlefield (#3862). "In
// addition": the type is appended, the others kept. Face-down creatures you
// control have menace is a layer-6 keyword grant over the face-down
// permanents (a face-down creature has no abilities of its own, CR 708.2a,
// so the grant is the only source of menace). The trigger is Cryptid
// Inspector's turned-face-up read over every permanent you control, Roshan
// included: it is face up and the same object by the time the event is
// emitted (CR 708.8).
//
// No simplification.
func init() {
	grant := func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
		if !effects.HasFold(c.Subtypes, "Assassin") {
			c.Subtypes = append(c.Subtypes, "Assassin")
		}
	}
	effects.Register(effects.Spec{
		OracleID:     "f60b21c1-d192-4a02-9b1f-a08e9e08a84a",
		Name:         "Roshan, Hidden Magister",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			{
				Label:     "Other creatures you control are Assassins in addition to their other types.",
				Layer:     game.Layer4Type,
				AppliesTo: effects.TribeFilter{Others: true, YoursOnly: true}.Matches,
				Apply:     grant,
			},
			{
				Label:        "Creature spells you control are Assassins in addition to their other types.",
				Layer:        game.Layer4Type,
				AffectsZones: effects.SpellsOnTheStack(),
				AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return target.Controller == source.Controller && target.IsCreature()
				},
				Apply: grant,
			},
			{
				Label:        "Creature cards you own that aren't on the battlefield are Assassins in addition to their other types.",
				Layer:        game.Layer4Type,
				AffectsZones: effects.CardsOffTheBattlefield(),
				AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return target.Owner == source.Controller && target.IsCreature()
				},
				Apply: grant,
			},
			func() game.StaticAbility {
				s := effects.KeywordGrant(func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return target.IsCreature() && target.FaceDownIsPermanent() && target.Controller == source.Controller
				}, "menace")
				s.Label = "Face-down creatures you control have menace."
				return s
			}(),
		},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventTurnedFaceUp, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.Controller == source.Controller
			}, "Roshan, Hidden Magister — you draw a card and you lose 1 life",
				effects.Do(effects.DrawCards{N: 1}, effects.YouLoseLife(1))),
		},
	})
}
