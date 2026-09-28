package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// hexproof_bypass.go — constructors for Spec.HexproofBypasses and
// Spec.WardSuppressions (#1560, ADR 0038's amendment of 2026-09-27):
//
//	HexproofBypasses: []game.HexproofBypass{
//		AsThoughNoHexproof(BySpellsAndAbilities, And(Creature(), OpponentControls())), // Nowhere to Run
//		AsThoughNoHexproof(BySpellsAndAbilitiesYouControl, OpponentControls()),       // Kaya's permanents
//		PlayersAsThoughNoHexproof(BySpellsAndAbilitiesYouControl, Opponent()),        // Kaya's players
//	},
//	WardSuppressions: []game.WardSuppression{
//		WardDoesNotTrigger(And(Creature(), OpponentControls())), // Nowhere to Run
//	},
//
// Every predicate is asked with the STATIC'S CONTROLLER as its
// `caster`, so OpponentControls() reads "your opponents control" from
// the permanent that prints the line — the same reading a Targets:
// clause gives it from the spell's controller. It is read live, on
// every targeting check, so a creature that changes control or enters
// after the static is covered exactly while it matches (a static
// ability, CR 611.3a, has no locked set).
//
// "As though it didn't have hexproof" is the whole grammar. A card
// that says "lose hexproof" (Shadowspear, Arcane Lighthouse) removes
// the ability in layer 6 and is RemoveKeywordsMod, not this.

// HexproofBypassSources is WHOSE spells and abilities may ignore the
// hexproof — the one printed difference between Nowhere to Run and
// Glaring Spotlight.
type HexproofBypassSources int

const (
	// BySpellsAndAbilities is Nowhere to Run's unrestricted "can be
	// the targets of spells and abilities": any player's, so the
	// creature's other opponents benefit too.
	BySpellsAndAbilities HexproofBypassSources = iota
	// BySpellsAndAbilitiesYouControl is "…of spells and abilities you
	// control" (Glaring Spotlight, Kaya, Bane of the Dead).
	BySpellsAndAbilitiesYouControl
)

// AsThoughNoHexproof is "<permanents matching every predicate> can be
// the targets of <sources> as though they didn't have hexproof".
func AsThoughNoHexproof(by HexproofBypassSources, preds ...CardPredicate) game.HexproofBypass {
	match := And(preds...)
	return game.HexproofBypass{
		Permanent: func(g *game.Game, target *game.Card, source *game.Card) bool {
			return match(g, source.Controller, *target)
		},
		YoursOnly: by == BySpellsAndAbilitiesYouControl,
	}
}

// PlayersAsThoughNoHexproof is the player half (CR 702.11d): "<players
// matching every predicate> can be the targets of <sources> as though
// they didn't have hexproof" — Kaya, Bane of the Dead's "your
// opponents".
func PlayersAsThoughNoHexproof(by HexproofBypassSources, preds ...PlayerPredicate) game.HexproofBypass {
	return game.HexproofBypass{
		Player: func(g *game.Game, target *game.Player, source *game.Card) bool {
			for _, p := range preds {
				if p != nil && !p(g, source.Controller, target) {
					return false
				}
			}
			return true
		},
		YoursOnly: by == BySpellsAndAbilitiesYouControl,
	}
}

// WardDoesNotTrigger is "ward abilities of <permanents matching every
// predicate> don't trigger" (CR 702.21). It stops the ward TRIGGER,
// which is all a ward ability is, and nothing else: a triggered
// ability that merely resembles ward without being one (Diffusion
// Sliver's) is not built on WardGranted and is not suppressed, which
// is the printed text.
func WardDoesNotTrigger(preds ...CardPredicate) game.WardSuppression {
	match := And(preds...)
	return game.WardSuppression{
		Warded: func(g *game.Game, warded *game.Card, source *game.Card) bool {
			return match(g, source.Controller, *warded)
		},
	}
}
