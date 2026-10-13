package effects

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Akki Lavarunner // Tok-Tok, Volcano Born — {3}{R} Creature — Goblin
// Warrior 1/1 // Legendary Creature — Goblin Shaman 2/2, a flip card
// (#1892, ADR 0151):
//
//	Top:     "Haste
//	          Whenever this creature deals damage to an opponent, flip
//	          it."
//	Flipped: "Protection from red
//	          If a red source would deal damage to a player, it deals
//	          that much damage plus 1 to that player instead."
//
// CR 710.2's own example: Akki Lavarunner is nonlegendary and Tok-Tok
// is legendary, so the flip changes what "legendary creatures" sees.
// Its keywords change with the half that is up (ADR 0151 decision 3):
// flipped, it loses haste and has protection from red. Its colour and
// mana cost stay red and {3}{R} (CR 710.1c), so Tok-Tok is itself a red
// source.
//
// Tok-Tok's replacement is Torbran's shape with a wider net: ANY red
// source, whoever controls it, dealing damage to ANY player — Tok-Tok's
// controller included — read off the damage event's snapshot of the
// source (#1417, CR 608.2h).
//
// No simplification.
const akkiLavarunnerOracleID = "47795817-73e5-4af6-bd1e-d69b193e8e9e"

func init() {
	Register(Spec{
		OracleID:        akkiLavarunnerOracleID,
		Name:            "Akki Lavarunner",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, ThisDealtDamageToAnOpponent, "Akki Lavarunner — flip it", Do(FlipThis{})),
		},
	})
	Register(Spec{
		OracleID:        akkiLavarunnerOracleID + "#1",
		Name:            "Tok-Tok, Volcano Born",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"protection from red"},
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, _ *game.Card) bool {
				if ev.Kind != game.RepEventDamage || ev.DamageAmount <= 0 || ev.DamageTarget == uuid.Nil {
					return false
				}
				if g.PlayerByIDForEffect(ev.DamageTarget) == nil {
					return false
				}
				ch, ok := damageSourceCharacteristics(ev, g)
				return ok && slices.Contains(ch.Colors, "R")
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.DamageAmount++
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Tok-Tok, Volcano Born: +1 damage from a red source",
		}},
	})
}
