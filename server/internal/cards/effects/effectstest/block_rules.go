package effectstest

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// block_rules_test.go — #750's card half: Spec.BlockRules and a token
// template's BlockRules slot, built from block_rules.go's constructors
// and read by the engine through CatalogBlockRules (ADR 0045 addendum,
// Decisions 11-13 and the 2026-09-24 amendment, Decisions 40-42).
//
// Every shape is pinned the same three ways, because ADR 0045 §3's
// invariant is that the three never disagree:
//
//   - the VERB: a legal block is accepted, an illegal one is refused
//     with its reason, the card's printed clause and the permanent
//     that carries the rule;
//   - the ENUMERATOR: legal.EnumerateFor offers exactly the blocks the
//     verb accepts;
//   - the VIEW: block_decision_seats (the #328 signal) says the seat
//     owes a decision only when some block is legal.

const (
	ProwlersHelmOracle       = "603ef53f-6191-4229-b85f-adad6f4503ce"
	LegolasGreenleafOracle   = "beacbb51-51fe-47f7-a612-02cd93fbffdd"
	ChampionOfLambholtOracle = "c549b0fd-1e08-4873-952e-a14dc45a0fd2"
	GingerbruteOracle        = "10b8d4c7-7553-4d76-b643-d98b80701e13"
	ThievesToolsOracle       = "7fe361ef-a168-4847-92a2-21c1661aac06"
	VorracBattlehornsOracle  = "89750d72-d1c0-4c8a-aa72-fdd48570aa92"
	RampagingCeratopsOracle  = "91d01119-9b3e-4629-8cf0-bbca2c19a7cc"
	CanopyCoverOracle        = "5b84101e-7e23-437d-835c-409bc061ecbb"
)

// brCreature seeds a non-catalog creature with keywords, able to
// attack and block this turn.
func BrCreature(g *game.Game, owner uuid.UUID, name, typeLine string, power, toughness int, keywords ...string) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
		Keywords: keywords,
	})
}

// brAttach attaches an Equipment or Aura without paying for the equip.
func BrAttach(t *testing.T, g *game.Game, attachment, host uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(attachment, game.TargetRef{Kind: game.TargetCard, ID: host}); err != nil {
			t.Fatalf("AttachForEffect: %v", err)
		}
	})
}

// brAttack declares `attackers` against seat 1 and parks the cursor on
// declare blockers.
func BrAttack(t *testing.T, g *game.Game, attackers ...uuid.UUID) {
	t.Helper()
	DeclareAttack(t, g, g.Seats[1].ID, attackers...)
	AdvanceTo(t, g, game.StepDeclareBlockers)
}

// brRefusal asserts `err` is an illegal-block refusal with `reason`
// and returns it.
func BrRefusal(t *testing.T, err error, reason game.BlockReason) *game.BlockRefusedError {
	t.Helper()
	if err == nil {
		t.Fatalf("the block was accepted; want a %q refusal", reason)
	}
	if !errors.Is(err, game.ErrIllegalBlock) {
		t.Fatalf("refusal does not wrap ErrIllegalBlock: %v", err)
	}
	var refused *game.BlockRefusedError
	if !errors.As(err, &refused) {
		t.Fatalf("refusal is %T, want *BlockRefusedError: %v", err, err)
	}
	if refused.Reason != reason {
		t.Fatalf("reason %q, want %q (%s)", refused.Reason, reason, refused.Sentence(uuid.Nil))
	}
	return refused
}

// brOffered is every (blocker, attacker) pair the enumerator offers
// `seat`, one entry per move: a single block is one pair, a grouped
// block (menace, a minimum) is all of its pairs in one entry.
func BrOffered(t *testing.T, g *game.Game, seat uuid.UUID) [][]game.BlockDeclaration {
	t.Helper()
	var out [][]game.BlockDeclaration
	for _, m := range legal.EnumerateFor(g, seat) {
		if m.Kind != legal.KindBlock {
			continue
		}
		var one struct {
			Blocker  string `json:"blocker"`
			Attacker string `json:"attacker"`
			Blocks   []struct {
				Blocker  string `json:"blocker"`
				Attacker string `json:"attacker"`
			} `json:"blocks"`
		}
		if err := json.Unmarshal(m.Params, &one); err != nil {
			t.Fatalf("block params %s: %v", m.Params, err)
		}
		var set []game.BlockDeclaration
		if one.Blocker != "" {
			set = append(set, game.BlockDeclaration{Blocker: uuid.MustParse(one.Blocker), Attacker: uuid.MustParse(one.Attacker)})
		}
		for _, b := range one.Blocks {
			set = append(set, game.BlockDeclaration{Blocker: uuid.MustParse(b.Blocker), Attacker: uuid.MustParse(b.Attacker)})
		}
		out = append(out, set)
	}
	return out
}

// brOffers reports whether some offered move blocks `attacker` with
// `blocker`.
func BrOffers(offered [][]game.BlockDeclaration, blocker, attacker uuid.UUID) bool {
	for _, set := range offered {
		for _, d := range set {
			if d.Blocker == blocker && d.Attacker == attacker {
				return true
			}
		}
	}
	return false
}

// brViewOwes is the wire's block_decision_seats for seat 1.
func BrViewOwes(g *game.Game) bool {
	for _, s := range protocol.ViewOfGame(g).Turn.BlockDecisionSeats {
		if s == 1 {
			return true
		}
	}
	return false
}

// --- CantBeBlockedThisTurnExceptBy: Gingerbrute -------------------

// gingerbruteActivate pays {1} and resolves the evasion ability.
func GingerbruteActivate(t *testing.T, g *game.Game, controller, brute uuid.UUID) {
	t.Helper()
	for _, p := range g.Seats {
		if p.ID == controller {
			p.ManaPool.AddMana(game.ManaToken{Color: "C"})
		}
	}
	if err := g.ActivateCatalogAbility(controller, brute, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	PassPriorityAroundTable(t, g)
}

// --- helpers --------------------------------------------------------

func MustBattlefieldCard(t *testing.T, g *game.Game, id uuid.UUID) *game.Card {
	t.Helper()
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			return &g.Battlefield.Cards[i]
		}
	}
	t.Fatalf("%s is not on the battlefield", id)
	return nil
}

func FindBattlefieldCardForTest(g *game.Game, id uuid.UUID) *game.Card {
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			return &g.Battlefield.Cards[i]
		}
	}
	return nil
}
