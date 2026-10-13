package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Whiplash Wordsmith // Vicious Verse — Creature — Vampire Sorcerer
// {3}{B/R}, 3/3 // Sorcery {B/R} (preparation card, CR 722):
//
//	"This creature enters prepared. (While it's prepared, you may cast a
//	 copy of its spell. Doing so unprepares it.)
//	 As long as an opponent was dealt noncombat damage this turn, this
//	 creature has flying and haste."
//
//	Vicious Verse — "Vicious Verse deals 1 damage to target opponent."
//
// "As long as an opponent was dealt noncombat damage this turn, this
// creature has flying and haste" is a self-only layer 6 static over the
// turn's damage tally (WasDealtNoncombatDamageThisTurn), woken on each
// damage event by DependsOnEvents; the turn's own advance clears it.
func init() {
	const id = "e1ffb884-a89e-4e2a-9beb-8bd5838b6a94"
	Register(Spec{
		OracleID:     id,
		Name:         "Whiplash Wordsmith",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersPrepared()},
		Static: []game.StaticAbility{{
			Layer:           game.Layer6Ability,
			Label:           "Has flying and haste as long as an opponent was dealt noncombat damage this turn.",
			DependsOnEvents: []game.EventKind{game.EventDealDamage},
			AppliesTo:       selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				for _, seat := range g.Seats {
					if seat.ID != source.Controller && g.WasDealtNoncombatDamageThisTurn(seat.ID) {
						appendKeywordsTo(c, []string{"flying", "haste"})
						return
					}
				}
			},
		}},
	})
	Register(Spec{
		OracleID:     id + "#1",
		Name:         "Vicious Verse",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target opponent", Opponent()),
		OnResolve:    viciousVerseResolve,
	})
}
