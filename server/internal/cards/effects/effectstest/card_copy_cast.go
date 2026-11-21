package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// card_copy_cast_test.go — #3258, ADR 0066's 2026-10-13 card-copy
// amendment: Mizzix's Mastery and Arcane Bombardment over CopyAndCast.

const (
	mizzixsMasteryOracle   = "40362fe0-a1a9-4d76-8c35-eac474b91af5"
	ArcaneBombardmentOracl = "1e36a42f-58e2-49bc-8c14-7f75e1d8ef98"
	ShockOracleForCopies   = "a9d288b8-cdc1-4e55-a0c9-d6edfc95e65d"
	DivinationOracle       = "273b339c-964b-4a18-8eb5-ceb8abcdfd9e"
)

// gyCard puts a catalog card with a real mana cost into p's graveyard.
func GyCard(p *game.Player, name, typeLine, cost, oracle string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost,
		OracleID: oracle, Owner: p.ID, Controller: p.ID,
	})
	return id
}

// cardCopies lists every card copy (Card.CardCopy) in every zone but
// the stack, the battlefield and the libraries.
func CardCopies(g *game.Game) []game.Card {
	var out []game.Card
	scan := func(z *game.Zone) {
		for _, c := range z.Cards {
			if c.CardCopy {
				out = append(out, c)
			}
		}
	}
	scan(g.Exile)
	for _, p := range g.Seats {
		scan(p.Hand)
		scan(p.Graveyard)
	}
	return out
}

// copyOf finds the copy whose printed name is name, or fails.
func CopyOf(t *testing.T, g *game.Game, name string) game.Card {
	t.Helper()
	for _, c := range CardCopies(g) {
		if c.Name == name {
			return c
		}
	}
	t.Fatalf("no copy of %q in exile or a graveyard", name)
	return game.Card{}
}

// castMizzix casts Mizzix's Mastery from the active player's hand.
func CastMizzix(t *testing.T, g *game.Game, overload bool, targets []game.TargetRef) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Mizzix's Mastery", TypeLine: "Sorcery", ManaCost: "{3}{R}",
		OracleID: mizzixsMasteryOracle, Owner: active.ID, Controller: active.ID,
	})
	ToMainForCost(t, g)
	params := game.CastSpellParams{Targets: targets}
	if overload {
		params.AlternativeCost = "overload"
		params.Targets = nil
	}
	if err := g.CastSpell(active.ID, id, params); err != nil {
		t.Fatalf("CastSpell Mizzix's Mastery: %v", err)
	}
	return id
}

func CastCopy(g *game.Game, who, copyID uuid.UUID, strict bool, targets ...game.TargetRef) error {
	return g.CastSpell(who, copyID, game.CastSpellParams{FromZone: "exile", Strict: strict, Targets: targets})
}

func PlayerTarget(id uuid.UUID) game.TargetRef {
	return game.TargetRef{Kind: game.TargetPlayer, ID: id}
}
