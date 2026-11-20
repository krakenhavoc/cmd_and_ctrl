package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lighthouse Chronologist — Creature — Human Wizard {1}{U}, 1/3 (Rise of the
// Eldrazi, #5581):
//
//	"Level up {U} ({U}: Put a level counter on this. Level up only as a
//	 sorcery.)
//	 LEVEL 4-6
//	 2/4
//	 LEVEL 7+
//	 3/5
//	 At the beginning of each end step, if it's not your turn, take an extra
//	 turn after this one."
//
// A leveler (leveler.go, CR 702.87a, 711.2a, 711.2b). The LEVEL 7+ trigger is
// a bundle the Chronologist grants itself behind the band's level check, so
// below level 7 it has no such ability. It triggers at each opponent's end
// step (every other player is an opponent at this table) and gives its
// controller one extra turn (extra_turns.go). The "if" can't stop holding
// before the ability resolves: the turn's active player doesn't change within
// its end step (CR 603.4).
//
// No simplification.
const lighthouseChronologistGrant = "lighthouse-chronologist/extra-turn"

func init() {
	Register(Spec{
		OracleID:     "b19ff0dd-2041-461f-aa62-6dc273e4bbe5",
		Name:         "Lighthouse Chronologist",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: lighthouseChronologistGrant,
			Triggered: []game.TriggeredAbility{
				On(game.EventBeginEndStep, ByAnOpponent,
					"Lighthouse Chronologist — take an extra turn after this one", Do(TakeExtraTurn{})),
			},
			Text: "At the beginning of each end step, if it's not your turn, take an extra turn after this one.",
		}},
		Activated: []ActivatedAbility{LevelUpKeyword("{U}")},
		Static: LevelBands(
			LevelBand{From: 4, To: 6, Power: 2, Toughness: 4},
			LevelBand{From: 7, Power: 3, Toughness: 5, Statics: []game.StaticAbility{
				Labelled(GrantAbilities(selfOnly, lighthouseChronologistGrant),
					`This creature has "At the beginning of each end step, if it's not your turn, take an extra turn after this one."`),
			}},
		),
	})
}
