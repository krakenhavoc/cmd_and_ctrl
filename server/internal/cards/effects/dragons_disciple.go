package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const dragonSubtypeWord = "Dragon"

// Dragon's Disciple — Creature — Human Monk {1}{W}, 1/3 (Adventures in the
// Forgotten Realms):
//
//	"As this creature enters, you may reveal a Dragon card from your hand.
//	 If you do or if you control a Dragon, this creature enters with a
//	 +1/+1 counter on it.
//	 Dragons you control have ward {1}."
//
// The reveal is the reveal-lands' entry card choice (CR 614.1c): revealing a
// Dragon card puts the counter on the entry, and declining or having none
// puts it there only if the controller controls a Dragon as it enters. The
// grant is WardGranted over Dragons its controller controls (the Disciple is
// a Human, so it does not ward itself).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3e5d3574-cc92-434a-a210-fcda906391c8",
		Name:         "Dragon's Disciple",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{{
			RowLabel:        "Enters with a +1/+1 counter on it if you reveal a Dragon card from your hand or control a Dragon.",
			EntryOnly:       true,
			Watches:         []game.EventKind{game.EventZoneMove},
			SelfReplacement: true,
			Label:           "Dragon's Disciple",
			PromptQuestion:  "Dragon's Disciple — reveal a Dragon card from your hand?",
			EntryCardChoice: &game.EntryCardChoice{
				Matches: func(c game.Card) bool { return c.HasSubtypeWord(dragonSubtypeWord) },
				Min:     0,
				Max:     1,
				Then: func(_ *game.Game, ev *game.ReplacementEvent, _ []uuid.UUID) error {
					ev.AddCounterAtETB(game.CounterPlusOne, 1)
					return nil
				},
			},
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
				return ev.Kind == game.RepEventMove &&
					ev.NewZone == game.ZoneBattlefield &&
					src != nil && ev.CardID == src.InstanceID
			},
			Controller: EnteringPermanentChooser,
			Replace: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) error {
				if src != nil && controlsADragonForDisciple(g, src.Controller) {
					ev.AddCounterAtETB(game.CounterPlusOne, 1)
				}
				return nil
			},
		}},
		Triggered: []game.TriggeredAbility{
			WardGranted(WardMana("{1}"), "Dragon's Disciple — Dragons you control have ward {1}",
				func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return target.Controller == source.Controller && target.HasSubtype(source.TextSubtype(dragonSubtypeWord))
				}),
		},
	})
}

func controlsADragonForDisciple(g *game.Game, you uuid.UUID) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == you && c.IsCreature() && c.HasSubtype(dragonSubtypeWord) {
			return true
		}
	}
	return false
}
