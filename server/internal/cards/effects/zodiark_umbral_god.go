package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Zodiark, Umbral God — Legendary Creature — God {B}{B}{B}{B}{B}, 5/5
// (Final Fantasy, #4462):
//
//	"Indestructible
//	 When Zodiark enters, each player sacrifices half the non-God
//	 creatures they control of their choice, rounded down.
//	 Whenever a player sacrifices another creature, put a +1/+1 counter
//	 on Zodiark."
//
// The enters trigger is Pox Plague's half-sacrifice over non-God creatures:
// each seat's count is read as the ability starts, the seats choose in APNAP
// order and the chosen creatures leave together. The counter trigger
// watches EventSacrifice, which fires for every player's sacrifice and
// before the creature leaves, once per creature, so Zodiark's own enters
// trigger feeds it.
//
// No simplification.
func init() {
	const counterLabel = "Zodiark — put a +1/+1 counter on Zodiark"
	Register(Spec{
		OracleID:        "f1190fb9-72a2-48f8-bc2d-61240f1e9119",
		Name:            "Zodiark, Umbral God",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"indestructible"},
		Triggered: []game.TriggeredAbility{
			TriggerEdict(WhenThisEnters("Zodiark — each player sacrifices half the non-God creatures they control",
				Do(Instruction(eachPlayerSacrificesHalfTheirNonGodCreatures)))),
			On(game.EventSacrifice, anotherCreatureWasSacrificed, counterLabel,
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
