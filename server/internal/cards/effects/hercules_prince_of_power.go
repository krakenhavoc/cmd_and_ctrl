package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hercules, Prince of Power — {2}{G} Legendary Creature — Demigod Warrior
// Hero 3/3 (Marvel Super Heroes, #4455):
//
//	"Power-up — {4}{G}: Put a +1/+1 counter on Hercules. He gains
//	 vigilance, indestructible, and haste until end of turn. (Activate
//	 each power-up ability only once. Reduce the cost by his mana cost if
//	 he entered this turn.)"
//
// Power-up is the engine's (CR 702.193): the turn he entered the ability
// costs {4}{G} less {2}{G}, which is {2}; after that it costs {4}{G}; and
// it can be activated once per object. The counter goes on, then the three
// keywords are granted to him for the turn, both only while he is still the
// same permanent on the battlefield.
//
// No simplification.
func init() {
	ab := PowerUp(
		"{4}{G}: Put a +1/+1 counter on Hercules. He gains vigilance, indestructible, and haste until end of turn.",
		ManaCost("{4}{G}"), func(g *game.Game, item *game.StackItem) error {
			if !b09SourceStillOnBattlefield(g, item) {
				return nil
			}
			if err := plusOneCountersOnThis(1)(g, item); err != nil {
				return err
			}
			return GrantKeywordUntilEOT{
				Target:   item.SourceCardID,
				Keywords: []string{"vigilance", "indestructible", "haste"},
				Label:    "Hercules, Prince of Power",
			}.Apply(NewContext(g, item))
		})
	ab.Purpose.Answers = game.AnswerPump
	Register(Spec{
		OracleID:     "c3a8cc79-3771-484d-84dd-1a97f8bd12c2",
		Name:         "Hercules, Prince of Power",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{ab},
	})
}
