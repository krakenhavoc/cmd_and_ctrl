package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mana_riders_test.go — the S22 mana-ability-rider batch: the eight
// painlands, the six Talismans, Ancient Tomb, Mana Confluence and
// City of Brass.
//
// The distinction every test here is aimed at is COST vs RIDER,
// because getting it backwards is invisible in the happy path and
// wrong in exactly the direction this project refuses to ship:
//
//   - A rider (painlands, Ancient Tomb) is part of the ability's
//     effect. The source taps at any life total, the mana arrives,
//     and the damage happens whether the player likes it or not.
//   - A cost (Mana Confluence) is validated first. Too little life
//     and the activation is REJECTED with the land still untapped.
//
// A third shape, City of Brass, looks like a rider and isn't: its
// "whenever this land becomes tapped" clause is a real CR 603
// trigger, so it uses the stack and fires on taps that were never
// activations. Folding it into the mana ability would have made the
// card better than printed.

const (
	ShivanReefOracle          = "0fe16212-66c3-4e45-a641-7391e9b2e304"
	AdarkarWastesOracle       = "d5ad26cc-2bdb-46b7-b8bf-dd099d5fa09b"
	AncientTombOracle         = "23467047-6dba-4498-b783-1ebc4f74b8c2"
	ManaConfluenceOracle      = "d0ee5bdc-2b69-4b73-9a20-ffcc18783b29"
	CityOfBrassOracle         = "f25351e3-539b-4bbc-b92d-6480acf4d722"
	TalismanOfDominanceOracle = "4c0a0448-b9d6-43a0-8549-64066dac63f0"
)

// riderLatestManaPick returns the most recent PendingChoiceMana queued for
// a chooser, or nil.
func RiderLatestManaPick(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	var out *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceMana && c.Chooser == chooser {
			out = c
		}
	}
	return out
}

// riderAnswerManaPicks answers every colour pick owed by chooser with
// `color` and returns how many it answered. #730 made an unanswered
// mana_pick gate the table, so a test that taps an any-colour source
// and then passes priority or walks the cursor on has to drain them
// first — which is what a player does anyway, since until the pick is
// answered the mana is not in the pool.
func RiderAnswerManaPicks(t *testing.T, g *game.Game, chooser uuid.UUID, color string) int {
	t.Helper()
	n := 0
	for i := 0; i < 16; i++ {
		pick := RiderLatestManaPick(g, chooser)
		if pick == nil {
			return n
		}
		if err := g.ResolveManaChoice(pick.ID, chooser, color); err != nil {
			t.Fatalf("ResolveManaChoice: %v", err)
		}
		n++
	}
	t.Fatal("colour picks did not drain")
	return n
}

// riderGiveCommander drops a commander with the supplied mana cost into a
// seat's command zone, which is what game.commanderIdentityFor reads
// to narrow a pipe ability. Used to prove the painland duals do NOT
// narrow.
func RiderGiveCommander(g *game.Game, p *game.Player, manaCost string) {
	g.WithWriteLock(func() {
		p.Command.PushTop(game.Card{
			InstanceID:  uuid.New(),
			Name:        "Test Commander",
			TypeLine:    "Legendary Creature — Human",
			ManaCost:    manaCost,
			Owner:       p.ID,
			Controller:  p.ID,
			IsCommander: true,
		})
	})
}
