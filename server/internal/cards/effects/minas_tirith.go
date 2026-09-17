package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Minas Tirith — Legendary Land (EDHREC rank 410):
//
//	"Minas Tirith enters tapped unless you control a legendary
//	 creature.
//	 {T}: Add {W}.
//	 {1}{W}, {T}: Draw a card. Activate only if you attacked with two
//	 or more creatures this turn."
//
// The white sibling of Rivendell and Mines of Moria: the same
// legendary-creature entry and a plain {W}. The draw's "Activate only
// if you attacked with two or more creatures this turn" is its
// activation condition (CR 602.1b, #743), read off this turn's attack
// declarations: two DIFFERENT creatures you declared as attackers. A
// creature that attacks in two combats of one turn counts once, which
// is what "two or more creatures" says.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7b0d7e62-0287-454a-8702-b0bfa7b41245",
		Name:         "Minas Tirith",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTappedUnless(b08ControlsLegendaryCreature)},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W}",
			Label:    "Add {W}",
		}},
		Activated: []ActivatedAbility{{
			Label:     "{1}{W}, {T}: Draw a card. Activate only if you attacked with two or more creatures this turn.",
			Cost:      Plus(ManaCost("{1}{W}"), TapCost()),
			Condition: minasTirithAttackedWithTwo,
			Effect:    b36DrawOne,
		}},
	})
}

// minasTirithAttackedWithTwo is Minas Tirith's condition: at least two
// distinct creatures `controller` declared as attackers this turn.
func minasTirithAttackedWithTwo(g *game.Game, controller, _ uuid.UUID) bool {
	seen := map[uuid.UUID]bool{}
	for _, ev := range g.EventsThisTurn() {
		if !attackDeclaredByYou(ev, controller) || ev.CardID == uuid.Nil {
			continue
		}
		seen[ev.CardID] = true
		if len(seen) >= 2 {
			return true
		}
	}
	return false
}
