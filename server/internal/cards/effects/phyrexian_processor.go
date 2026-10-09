package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Phyrexian Processor — Artifact {4}:
//
//	"As this artifact enters, pay any amount of life.
//	 {4}, {T}: Create an X/X black Phyrexian Minion creature token,
//	 where X is the life paid as this artifact entered."
//
// The payment is the life form of the number prompt (ADR 0129's
// amendment of 2026-10-09, #1941): any amount from 0 to its controller's
// life total (CR 119.4), none at all when their life total can't change
// (CR 119.8). It is stored on the permanent (game.Card.ChosenNumber), so
// it lasts as long as this object does and a Processor that leaves and
// comes back pays again (CR 400.7). The token's size is read as the
// ability resolves, from the Processor's last-known information when it
// has left (CR 608.2h), so sacrificing it in response still makes the
// token. Zero life paid makes a 0/0 that the state-based actions put
// into the graveyard (CR 704.5f), as printed.
//
// The payment is queued from the AsEnters hook, the as-enters family's
// shape (PayAnyAmountOfLifeAsEnters): the Processor is briefly on the
// battlefield with nothing paid, while the prompt holds the table, and
// its ability cannot be activated in that window.
//
// A bot pays half its life, keeping at least 10.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "36c800cb-b1ca-4432-ad3c-4d8b90337f4c",
		Name:         "Phyrexian Processor",
		Completeness: CompletenessFull,
		AsEnters:     PayAnyAmountOfLifeAsEnters("Phyrexian Processor", game.PayAmountPower, phyrexianProcessorGoal),
		Activated: []ActivatedAbility{{
			Label:   "{4}, {T}: Create an X/X black Phyrexian Minion creature token, where X is the life paid as this artifact entered.",
			Cost:    game.AbilityCost{Mana: "{4}", Tap: true},
			Purpose: game.Purpose{Tokens: 1},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return CreateToken{
					Controller: item.Controller,
					Template:   phyrexianMinionToken(LifePaidAsEntered(ctx)),
					N:          1,
				}.Apply(ctx)
			},
		}},
	})
}

// phyrexianProcessorGoal is what a bot pays: half its life, keeping at
// least 10 (the floor the heuristic holds every life payment to).
func phyrexianProcessorGoal(life int) int {
	return max(0, min(life/2, life-10))
}

// phyrexianMinionToken is the X/X black Phyrexian Minion. Built by hand
// because its size is the life paid, as Corpse Cobble's Zombie is.
// PrintedPTKnown, because 0/0 is a real size here.
func phyrexianMinionToken(x int) game.Card {
	x = max(x, 0)
	return game.Card{
		Name:           "Phyrexian Minion",
		TypeLine:       "Token Creature — Phyrexian Minion",
		Power:          x,
		Toughness:      x,
		PrintedPTKnown: true,
		Colors:         []string{"B"},
	}
}
