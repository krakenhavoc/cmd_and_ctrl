package cardsa

import (
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Adagia, Windswept Bastion — Land — Planet:
//
//	This land enters tapped.
//	{T}: Add {W}.
//	Station (Tap another creature you control: Put charge counters
//	equal to its power on this Planet. Station only as a sorcery.)
//	12+ | {3}{W}, {T}: Create a token that's a copy of target artifact
//	or enchantment you control, except it's legendary. Activate only
//	as a sorcery.
//
// The Planet proof card: a LAND with station (CR 721.1 — a Planet is
// a land with the keyword). Nothing in the land path conflicts with
// it. The land drop is a land drop, the mana ability is an ordinary
// {T} mana ability, and station's cost taps ANOTHER creature — never
// the land — so a Planet can tap for mana and be stationed in the
// same main phase. There is no P/T box, so it never becomes a
// creature.
//
// The 12+ line is a gated ACTIVATED ability (ActiveWhen on the
// ability, ADR 0071): below twelve counters it is absent from the
// menu and from the bots' move list, not greyed. "Except it's
// legendary" adds the supertype to the copy's type line, so the
// legend rule (CR 704.5j) applies to the copies — the printed
// drawback that stops Adagia minting a second copy of a legendary
// permanent's WORTH every turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "70d35dbd-1d91-4a2a-a643-6870d168f4f5",
		Name:         "Adagia, Windswept Bastion",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W}",
			Label:    "Add {W}",
		}},
		Activated: []effects.ActivatedAbility{
			effects.Station(),
			{
				Label:        "{3}{W}, {T}: Create a token that's a copy of target artifact or enchantment you control, except it's legendary. Activate only as a sorcery.",
				Cost:         effects.Plus(effects.ManaCost("{3}{W}"), effects.TapCost()),
				Targets:      effects.TargetPermanent("target artifact or enchantment you control", effects.Or(effects.Artifact(), effects.Enchantment()), effects.YouControl()),
				SorcerySpeed: true,
				ActiveWhen:   effects.AtChargeCounters(12),
				Effect:       adagiaLegendaryCopy,
			},
		},
	})
}

// adagiaLegendaryCopy is the 12+ ability's resolution: one token copy
// of the still-legal target, with "Legendary" added to its type line.
func adagiaLegendaryCopy(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		return effects.CreateTokenCopy{
			Controller: item.Controller,
			Copy:       t.ID,
			N:          1,
			Except: func(tok *game.Card) {
				tok.TypeLine = LegendaryTypeLine(tok.TypeLine)
			},
		}.Apply(ctx)
	}
	return nil
}

// legendaryTypeLine adds the Legendary supertype to a type line that
// does not already carry it — "except it's legendary". Supertypes
// lead the line, so it is prepended.
func LegendaryTypeLine(tl string) string {
	for _, w := range strings.Fields(tl) {
		if w == "Legendary" {
			return tl
		}
	}
	if tl == "" {
		return "Legendary"
	}
	return "Legendary " + tl
}
