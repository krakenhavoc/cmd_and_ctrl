package cardsn

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Necromancer's Magemark — Enchantment — Aura {2}{B} (Guildpact):
//
//	"Enchant creature
//	 Creatures you control that are enchanted get +1/+1.
//	 If a creature you control that's enchanted would die, return it to its
//	 owner's hand instead."
//
// The anthem is Beastmaster's Magemark's. The second sentence is a CR 614
// replacement of the creature's move from the battlefield to a graveyard
// (CR 700.4), asked while the creature is still on the battlefield with
// its Aura attached. It never died, so its dies triggers do not fire. A
// token that is returned ceases to exist (CR 111.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dffa7c06-096a-47f8-9645-1ba7306ba5b0",
		Name:         "Necromancer's Magemark",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.OldB570MagemarkPump()},
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventZoneMove},
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
				if src == nil || ev.Kind != game.RepEventMove || ev.OldZone != game.ZoneBattlefield || ev.NewZone != game.ZoneGraveyard {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && effects.OldB570YourEnchantedCreature(&c, g, src)
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.NewZone = game.ZoneHand
				ev.NewZoneOwner = uuid.Nil
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label:    "Necromancer's Magemark: return it to its owner's hand instead of dying",
			RowLabel: "If a creature you control that's enchanted would die, return it to its owner's hand instead.",
		}},
	})
}
