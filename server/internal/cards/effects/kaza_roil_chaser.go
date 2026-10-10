package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kaza, Roil Chaser — Legendary Creature — Human Wizard {U}{R}, 1/2:
//
//	"Flying, haste
//	 {T}: The next instant or sorcery spell you cast this turn costs {X}
//	 less to cast, where X is the number of Wizards you control as this
//	 ability resolves."
//
// X is counted when the ability resolves and stored in the promise
// (#1852), which is what "as this ability resolves" says; Wizards
// entering later change nothing. Kaza counts herself while she is on the
// battlefield. The reduction is generic mana only, like every reduction.
func init() {
	Register(Spec{
		OracleID:        "0b8517c0-0441-4fd9-99ec-2274e57c3cb6",
		Name:            "Kaza, Roil Chaser",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		Activated: []ActivatedAbility{{
			Label:   "{T}: The next instant or sorcery spell you cast this turn costs {X} less to cast, where X is the number of Wizards you control as this ability resolves.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    TapCost(),
			Effect:  kazaRoilChaserPromise,
		}},
	})
}

func kazaRoilChaserPromise(g *game.Game, item *game.StackItem) error {
	x := 0
	for i := range g.Battlefield.Cards {
		c := g.Battlefield.Cards[i]
		if c.Controller == item.Controller && c.HasSubtype("Wizard") {
			x++
		}
	}
	if x == 0 {
		return nil
	}
	return GrantNextSpellPromise{From: "Kaza, Roil Chaser", Promise: game.NextSpellPromise{
		Filter: game.PermissionFilter{InstantOrSorceryOnly: true},
		Reduce: x,
		Text:   "The next instant or sorcery spell you cast this turn costs {X} less to cast.",
	}}.Apply(NewContext(g, item))
}
