package cardsp

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Pearl Collector — Creature — Human Warlock {2}{B}, 3/3 (Mystery
// Booster Commander, #5948):
//
//	"Deathtouch, Lifelink
//	 At the beginning of your second main phase, if you gained 4 or more
//	 life this turn, conjure a card named Mox Pearl into your hand. This
//	 ability triggers only once.
//	 {2}{W}: Another target creature perpetually gains lifelink."
//
// The conjured Mox Pearl is a real card, and "perpetually" writes the
// keyword onto the card itself so it keeps it in every zone (ADR 0099
// amendment 2026-11-24, Card.PerpetualKeywords).
//
// No simplification.
func init() {
	const label = "Pearl Collector — conjure a card named Mox Pearl into your hand"
	effects.Register(effects.Spec{
		OracleID:        "a5ce937f-425f-4c24-a9d3-1bf8a244632b",
		Name:            "Pearl Collector",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch", "lifelink"},
		Triggered: []game.TriggeredAbility{
			effects.TriggersOnlyOnce(effects.InterveningIfYou("you gained 4 or more life this turn", YouGainedFourOrMoreLifeThisTurn,
				effects.AtYourSecondMainPhase(label, effects.Do(effects.CheckConjure("Pearl Collector", effects.Conjure{Names: []string{"Mox Pearl"}, Zone: game.ZoneHand}))))),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{W}: Another target creature perpetually gains lifelink.",
			Cost:    effects.ManaCost("{2}{W}"),
			Targets: effects.Another(effects.TargetCreature("another target creature")),
			Effect:  effects.Do(effects.PerpetuallyGainKeyword{Keyword: "lifelink"}),
		}},
	})
}

// YouGainedFourOrMoreLifeThisTurn is "you gained 4 or more life this
// turn": this turn's life gained, added up.
func YouGainedFourOrMoreLifeThisTurn(g *game.Game, you uuid.UUID) bool {
	return g.TurnTallyFor(you).LifeGained >= 4
}
