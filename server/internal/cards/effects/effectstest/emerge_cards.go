package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// emerge_cards_test.go — ADR 0135 §4 (#2416): emerge (CR 702.119) against
// the real catalog. The offer is a one-permanent sacrifice whose price is
// reduced by that permanent's mana value (CR 702.119a), taken off the
// generic part only (CR 118.7a) after any increase (CR 601.2f); the
// sacrificed creature may tap for mana before it goes (owner decision 3);
// the enumerator prices each payment on its own (owner decision 5).

const (
	AbundantMawOracle        = "3676b745-001d-46e6-880f-2fe26476a38d"
	AdiposeOffspringOracle   = "1e21e57e-3bc9-41a8-9746-57b583a5ad63"
	CrabominationOracle      = "5c26c30b-bc4a-4d20-8406-ef9022256e0d"
	CrestingMosasaurusOracle = "bf245433-8e3b-47d7-b628-3cdae93c384d"
	DecimatorOracle          = "7f522ded-09fd-457e-9efe-0a6324925e4c"
	DrownyardBehemothOracle  = "bfe1a52a-f2f8-4c62-8502-037f2d1395b2"
	ElderDeepFiendOracle     = "4eafe717-4ba4-4901-8c67-11757230eb54"
	ItOfTheHorridSwarmOracle = "f4ad190a-6c82-4923-af54-b4cce8d6465f"
	LashweedLurkerOracle     = "a7dd75bf-ecda-406c-baa2-8c051f138809"
	MockeryOfNatureOracle    = "5da5240e-5d87-494f-aec9-64b7a3f0d935"
	TwistedRiddlekeeperOrcl  = "c7ad20a3-51bf-4a22-a09c-fff8cae22765"
	VexingScuttlerOracle     = "2a5605a8-606e-4866-9473-7a2240ac6b8c"
	WretchedGryffOracle      = "83d0c4bc-aab9-4c14-a7dc-a344fcce4fea"
)

// emergeTable is a main phase on seat 0's turn with an empty hand.
func EmergeTable(t *testing.T) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	g := NewCatalogGame(t)
	AdvanceToMain(t, g)
	me := g.Seats[0]
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	return g, me, g.Seats[1]
}

// emergeFodder puts a creature with mana cost `cost` on `owner`'s
// battlefield.
func EmergeFodder(g *game.Game, owner uuid.UUID, name, cost string, power, toughness int) uuid.UUID {
	return ApaPush(g, owner, owner, game.Card{Name: name, TypeLine: "Creature — Test", ManaCost: cost, Power: power, Toughness: toughness})
}

func EmergeLands(g *game.Game, owner uuid.UUID, n int, name string) []uuid.UUID {
	var out []uuid.UUID
	for i := 0; i < n; i++ {
		out = append(out, PushLandFor(g, owner, name, "Basic Land — "+name))
	}
	return out
}

// castEmerge casts `card` for its emerge cost sacrificing `sac`. Strict
// casts pay their mana through the auto-tapper; the others are permissive.
func CastEmerge(g *game.Game, p *game.Player, card, sac uuid.UUID, strict bool, targets ...game.TargetRef) error {
	return g.CastSpell(p.ID, card, game.CastSpellParams{
		AlternativeCost: effects.AltCostKeyEmerge, AltCostIDs: []uuid.UUID{sac},
		Strict: strict, AutoTap: strict, Targets: targets,
	})
}

// emergePrice is what the engine's one pricer charges for `card` emerged
// over `sac`.
func EmergePrice(t *testing.T, g *game.Game, p *game.Player, card, sac uuid.UUID) string {
	t.Helper()
	var out string
	g.WithWriteLock(func() {
		c, ok := g.LookupCardForEffect(card)
		if !ok {
			t.Fatal("the emerge card is gone")
		}
		price, err := g.PriceCastForEffect(p.ID, c, game.CastSpellParams{AlternativeCost: effects.AltCostKeyEmerge, AltCostIDs: []uuid.UUID{sac}})
		if err != nil {
			t.Fatalf("PriceCastForEffect: %v", err)
		}
		out = price.Total.String()
	})
	return out
}
