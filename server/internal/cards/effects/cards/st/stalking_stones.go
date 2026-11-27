package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Stalking Stones — Land (Duel Decks: Elspeth vs. Tezzeret, #5578):
//
//	"{T}: Add {C}.
//	 {6}: This land becomes a 3/3 Elemental artifact creature that's still a
//	 land. (This effect lasts indefinitely.)"
//
// The mana ability is Academy Ruins'. The second is a layer-4 type change
// with the base size set in layer 7b (CR 613.1d, 613.4b), no end: it lasts
// as long as the land is this object (CR 611.2a, 400.7). It adds Artifact
// and Creature and the Elemental subtype, and keeps Land. A Stalking Stones
// that is already a creature just has the same values set again.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f3658894-3d3d-4cd4-b0ac-c53e1d08747c",
		Name:         "Stalking Stones",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{6}: This land becomes a 3/3 Elemental artifact creature that's still a land.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    effects.ManaCost("{6}"),
			Effect:  S82OldB474StalkingStonesAnimate,
		}},
	})
}

// S82OldB474StalkingStonesAnimate is the indefinite "this land becomes a 3/3
// Elemental artifact creature that's still a land", pinned to this object.
func S82OldB474StalkingStonesAnimate(g *game.Game, item *game.StackItem) error {
	self := item.SourceCardID
	if !effects.OnBattlefield(g, self) || effects.SourceIsNewObject(g, item) {
		return nil
	}
	return effects.ScopedEffectFor{
		Target: self,
		Mods: []game.Mod{
			game.AddTypesMod("Artifact", "Creature"),
			game.AddSubtypesMod(g.TextReaderForEffect(item).TextSubtypes("Elemental")...),
			game.SetBasePowerMod(3),
			game.SetBaseToughnessMod(3),
		},
		Duration: g.PinnedTo(game.IndefiniteDuration(), self),
		Label:    "Stalking Stones — becomes a 3/3 Elemental artifact creature",
	}.Apply(effects.NewContext(g, item))
}
