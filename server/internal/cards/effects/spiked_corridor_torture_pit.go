package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Spiked Corridor // Torture Pit — Enchantment — Room (ADR 0103):
//
//	Spiked Corridor {3}{R}: "When you unlock this door, create three
//	1/1 red Devil creature tokens with "When this token dies, it deals
//	1 damage to any target.""
//	Torture Pit {3}{R}: "If a source you control would deal noncombat
//	damage to an opponent, it deals that much damage plus 2 instead."
//
// The Devil carries its own trigger as a token template (ADR 0083), so
// it is a restore point. Torture Pit is Torbran's replacement narrowed
// the other way: any source you control, noncombat, and only a
// PLAYER who is an opponent (a permanent they control is not "an
// opponent").
func init() {
	Register(Room(RoomSpec{
		OracleID:     "e49b902f-a556-4dab-8928-93caf0a3f609",
		Name:         "Spiked Corridor // Torture Pit",
		Completeness: CompletenessFull,
		Left: Door{Triggered: []game.TriggeredAbility{
			WhenYouUnlockThisDoor(game.DoorLeft, "Spiked Corridor — create three 1/1 red Devil creature tokens",
				Do(CreateToken{Template: devilToken(), N: 3})),
		}},
		Right: Door{Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
				if ev.Kind != game.RepEventDamage || ev.DamageAmount <= 0 || ev.IsCombatDamage {
					return false
				}
				if !damageSourceControlledBy(ev, g, src.Controller) {
					return false
				}
				p := g.PlayerByIDForEffect(ev.DamageTarget)
				return p != nil && p.ID != src.Controller
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.DamageAmount += 2
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Torture Pit: +2 noncombat damage to an opponent",
		}}},
	}))
}

// devilToken is the 1/1 red Devil with "When this token dies, it deals 1
// damage to any target."
func devilToken() game.Card { return tokenFromCatalog(printedDevilToken) }

// printedDevilToken is the Devil as PRINTED. The damage's source is the
// token that died (CR 603.10a: a leaves-the-battlefield trigger).
func printedDevilToken() tokenTemplate {
	return tokenTemplate{
		Slug: "devil",
		Card: game.Card{
			Name:      "Devil",
			TypeLine:  "Token Creature — Devil",
			Power:     1,
			Toughness: 1,
			Colors:    []string{"R"},
		},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisDies("Devil — 1 damage to any target", sourceDealsOneToFirstTarget), TargetAny()),
		},
		Text: "When this token dies, it deals 1 damage to any target.",
	}
}
