package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Armored Armadillo — Creature — Armadillo {W}, 0/4 (Outlaws of Thunder
// Junction, #4468):
//
//	"Ward {1} (Whenever this creature becomes the target of a spell or
//	 ability an opponent controls, counter it unless that player pays
//	 {1}.)
//	 {3}{W}: This creature gets +X/+0 until end of turn, where X is its
//	 toughness."
//
// Ward is ward.go's triggered ability. X is the creature's toughness as
// the ability resolves (CR 608.2h), after every layer, so a toughness
// pump or a +1/+1 counter it already has counts; the +X/+0 is then
// locked in for the turn. A source that left, or came back as a new
// object, gets nothing (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "dca25245-523a-4587-a059-a14ee2ef908d",
		Name:         "Armored Armadillo",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{Ward(WardMana("{1}"), "Armored Armadillo — ward {1}")},
		Activated: []ActivatedAbility{{
			Label:   "{3}{W}: This creature gets +X/+0 until end of turn, where X is its toughness.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{3}{W}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !sourceIsStillThisPermanent(g, item) {
					return nil
				}
				c, ok := g.LookupCardForEffect(item.SourceCardID)
				if !ok || c.CurrentToughness() <= 0 {
					return nil
				}
				return BoostUntilEOT{
					Target: item.SourceCardID,
					Power:  c.CurrentToughness(),
					Label:  "Armored Armadillo — +X/+0, X its toughness",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
