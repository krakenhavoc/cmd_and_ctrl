package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Armored Scrapgorger — Creature — Phyrexian Beast {1}{G}, 0/3 (Phyrexia:
// All Will Be One):
//
//	"This creature gets +3/+0 as long as it has three or more oil counters
//	 on it.
//	 {T}: Add one mana of any color.
//	 Whenever this creature becomes tapped, exile target card from a
//	 graveyard and put an oil counter on this creature."
//
// The bonus is read live. The mana ability is a plain tap for any colour.
// The trigger fires on every tap (the mana ability, an attack without
// vigilance, a tap effect), targets a card in any graveyard, and puts the
// counter on as the same ability resolves; with no card to target it is
// removed and no counter is put on (CR 603.3d).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0d3a4db2-df79-4946-8f87-b174cf9c293c",
		Name:         "Armored Scrapgorger",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
		Static: []game.StaticAbility{{
			Label:     "This creature gets +3/+0 as long as it has three or more oil counters on it.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
				if source.Counters[effects.S82OneOil] >= 3 {
					c.Power += 3
				}
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.OnAny([]game.EventKind{game.EventTapCard, game.EventAttack}, effects.ThisBecameTapped,
					"Armored Scrapgorger — exile target card from a graveyard and put an oil counter on it",
					effects.Do(effects.ExileChosenTarget{}, effects.CounterOnThis{Kind: effects.S82OneOil, N: 1})),
				effects.TargetCardInGraveyard("target card from a graveyard")),
		},
	})
}
