package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Myojin of Night's Reach — Legendary Creature — Spirit {5}{B}{B}{B}, 5/2
// (Foundations, #4465):
//
//	"Myojin of Night's Reach enters with a divinity counter on it if you
//	 cast it from your hand.
//	 Myojin of Night's Reach has indestructible as long as it has a
//	 divinity counter on it.
//	 Remove a divinity counter from Myojin of Night's Reach: Each
//	 opponent discards their hand."
//
// The counter is a CR 614.1c entry clause read off the resolving spell's
// cast record (CountersIfCastFrom, Patched Plaything's): cast from the
// command zone as a commander it is not cast from the hand (the ruling),
// and a reanimated Myojin has none. Indestructible is a layer 6 self
// grant that reads the counter live. The ability's cost is removing the
// counter, so it can be activated once per counter and in response to
// removal, and each opponent discards the whole hand as the ability
// resolves.
//
// No simplification.
const myojinDivinityCounter = "divinity"

func init() {
	hasDivinity := func(target *game.Card, g *game.Game, source *game.Card) bool {
		return selfOnly(target, g, source) && target.Counters[myojinDivinityCounter] > 0
	}
	indestructible := KeywordGrant(hasDivinity, "indestructible")
	indestructible.Label = "Has indestructible as long as it has a divinity counter on it."
	Register(Spec{
		OracleID:     "5ac65688-0a74-4dc4-b6cb-12bdf8aed01e",
		Name:         "Myojin of Night's Reach",
		Completeness: CompletenessFull,
		EntersWithCountersFromCast: []game.EntryCountersFromCast{
			CountersIfCastFrom(game.ZoneHand, myojinDivinityCounter, 1),
		},
		Static: []game.StaticAbility{indestructible},
		Activated: []ActivatedAbility{{
			Label: "Remove a divinity counter from Myojin of Night's Reach: Each opponent discards their hand.",
			Cost:  RemoveCountersFromThis(myojinDivinityCounter, 1),
			// ADR 0142: a hand attack answers nothing on the stack.
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, id := range ctx.Opponents() {
					if _, err := discardWholeHand(g, id); err != nil {
						return err
					}
				}
				return nil
			},
		}},
	})
}
