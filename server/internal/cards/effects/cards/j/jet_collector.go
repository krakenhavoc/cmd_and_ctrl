package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Jet Collector — Creature — Merfolk Scout {1}{G}, 2/2 (Mystery Booster
// Commander, #5948):
//
//	"At the beginning of your second main phase, if there are four or
//	 more cards in your graveyard, conjure a card named Mox Jet into your
//	 hand. This ability triggers only once.
//	 {X}{B}: Return target creature card with mana value X from your
//	 graveyard to the battlefield with a finality counter on it. Activate
//	 only as a sorcery."
//
// The "if" is an intervening if (CR 603.4), and a trigger that never
// happened does not spend "triggers only once" (#5245). The conjured Mox
// Jet is a real card (ADR 0099 amendment 2026-11-24). The activation is
// Grim Reaper's Scythe's return with Sorin's mana-value-X target.
//
// No simplification.
func init() {
	const label = "Jet Collector — conjure a card named Mox Jet into your hand"
	effects.Register(effects.Spec{
		OracleID:     "191f5f95-bfb8-4d3c-a223-c7654cc507d8",
		Name:         "Jet Collector",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggersOnlyOnce(effects.InterveningIfYou("there are four or more cards in your graveyard", FourOrMoreCardsInYourGraveyard,
				effects.AtYourSecondMainPhase(label, effects.Do(effects.CheckConjure("Jet Collector", effects.Conjure{Names: []string{"Mox Jet"}, Zone: game.ZoneHand}))))),
		},
		Activated: []effects.ActivatedAbility{{
			Label: "{X}{B}: Return target creature card with mana value X from your graveyard to the battlefield with a finality counter on it. Activate only as a sorcery.",
			Cost:  effects.ManaCost("{X}{B}"),
			Targets: effects.TargetCardInGraveyard("target creature card with mana value X from your graveyard",
				effects.Creature(), effects.YouOwn()).WithManaValueEqualsX(),
			SorcerySpeed: true,
			Effect:       effects.GrimReapersScytheReturn,
		}},
	})
}

// FourOrMoreCardsInYourGraveyard is "there are four or more cards in
// your graveyard".
func FourOrMoreCardsInYourGraveyard(g *game.Game, you uuid.UUID) bool {
	p := g.PlayerByIDForEffect(you)
	return p != nil && p.Graveyard != nil && len(p.Graveyard.Cards) >= 4
}
