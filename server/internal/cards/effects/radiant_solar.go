package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Radiant Solar — Creature — Angel {5}{W}, 3/6:
//
//	"Flying, lifelink
//	 Whenever this creature or another nontoken creature you control
//	 enters, venture into the dungeon.
//	 {W}, Discard this card: Venture into the dungeon and you gain 3
//	 life."
//
// The last ability is activated from the hand (Channel's shape: Zones
// hand, DiscardSelf), at instant speed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "21a86dcd-03e3-409b-b214-5feed07db584",
		Name:            "Radiant Solar",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "lifelink"},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, radiantSolarEntered, ventureLabel("Radiant Solar"), Venture),
		},
		Activated: []ActivatedAbility{{
			Label:   "{W}, Discard this card: Venture into the dungeon and you gain 3 life.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    game.AbilityCost{Mana: "{W}", DiscardSelf: true},
			Zones:   []game.ZoneKind{game.ZoneHand},
			Effect:  Venturing(GainLife{Amount: 3}),
		}},
	})
}

// radiantSolarEntered is "this creature or another nontoken creature
// you control enters".
func radiantSolarEntered(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
	if source == nil {
		return false
	}
	if ev.CardID == source.InstanceID {
		return true
	}
	if !AnotherCreatureEnteredUnderYourControl(ev, source, lki, g) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && !IsToken(c)
}
