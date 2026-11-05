package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Elas il-Kor, Sadistic Pilgrim — Legendary Creature — Phyrexian Kor
// Cleric {W}{B}, 2/2 (EDHREC rank 615):
//
//	"Deathtouch
//	 Whenever another creature you control enters, you gain 1 life.
//	 Whenever another creature you control dies, each opponent loses
//	 1 life."
//
// Soul Warden and Zulaport Cutthroat on one body. Two triggers, both
// "ANOTHER creature you control" — Elas's own entry and own death are
// excluded, unlike Zulaport's "this creature or another". The drain
// is life loss, not damage, so no prevention applies.
//
// Each row declares what it pays (#4990): the enters row 1 life per
// creature, as Hinterland Sanctifier's identical row does, and the
// dies row DeathPayoff (ADR 0126 §6), as Zulaport's does, so the bot
// keeps Elas in play and sacrifices around it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a4d33e21-d3fe-4d5b-903c-b75e859d6f7f",
		Name:            "Elas il-Kor, Sadistic Pilgrim",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			CreatureEnterPaying(WheneverAnotherCreatureEntersUnderYourControl("Elas il-Kor — you gain 1 life", Do(GainLife{Amount: 1})),
				game.CreatureEnterPayoff{LifeGain: 1}),
			TriggerWithPurpose(On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return anotherCreatureYouControlDied(ev, source, g)
			}, "Elas il-Kor — each opponent loses 1 life", Do(EachOpponentLosesLife{N: 1})), game.Purpose{DeathPayoff: true}),
		},
	})
}
