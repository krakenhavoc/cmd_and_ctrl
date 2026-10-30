package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ability_cost_either_view_test.go — #4566: an either/or activation cost
// ships one entry per branch, with the picker that branch needs and
// whether the controller can pay it; the discard branch's hand cards and
// verdict reach the controller's frame only (#1369).

func seatEitherCostSource(g *game.Game, owner uuid.UUID) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id, Name: "Either Pinger", TypeLine: "Creature — Human Assassin",
		OracleID: "00000000-0000-0000-0000-000000004566",
		Owner:    owner, Controller: owner, Power: 2, Toughness: 3,
		ActivatedAbilities: []game.ActivatedAbilityShape{{
			Label: "{3}, {T}, Sacrifice an artifact or discard a nonland card: ping",
			Cost: game.AbilityCost{Mana: "{3}", Tap: true, Either: []game.AbilityCostBranch{
				{Kind: game.AbilityCostBranchSacrifice, Count: 1, Label: "an artifact", Of: game.PermanentQuery{Types: []string{"artifact"}}},
				{Kind: game.AbilityCostBranchDiscard, Count: 1, Label: "a nonland card", Nonland: true},
			}},
		}},
	})
	return id
}

func TestActivatedAbilityViewCarriesCostBranches(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	handCardFor(me, "Forest", "Basic Land — Forest")
	spell := handCardFor(me, "Shock", "Instant")
	src := seatEitherCostSource(g, me.ID)

	a := controllerFrameCard(t, g, src).ActivatedAbilities[0]
	if len(a.CostBranches) != 2 {
		t.Fatalf("cost_branches = %+v, want two", a.CostBranches)
	}
	sac, disc := a.CostBranches[0], a.CostBranches[1]
	if sac.Label != "Sacrifice an artifact" || sac.Payable || sac.SacrificeOptions == nil || len(sac.SacrificeOptions.Cards) != 0 {
		t.Errorf("sacrifice branch %+v: want unpayable with no artifact", sac)
	}
	if disc.Label != "Discard a nonland card" || !disc.Payable || disc.DiscardCostN != 1 ||
		len(disc.DiscardCostOptions) != 1 || disc.DiscardCostOptions[0] != spell.String() {
		t.Errorf("discard branch %+v: want payable with the nonland card only", disc)
	}
	if a.SacrificeOptions != nil || a.DiscardCostN != 0 {
		t.Error("the row itself carries a branch's clause; only the branches should")
	}

	art := uuid.New()
	g.Battlefield.PushTop(game.Card{InstanceID: art, Name: "Bauble", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
	a = controllerFrameCard(t, g, src).ActivatedAbilities[0]
	if b := a.CostBranches[0]; !b.Payable || len(b.SacrificeOptions.Cards) != 1 || b.SacrificeOptions.Cards[0] != art.String() {
		t.Errorf("sacrifice branch %+v: want payable with the artifact", b)
	}
}

// Another seat sees the branches and the board's sacrifice options, but
// not the hand cards that could pay the discard, nor whether any can.
func TestCostBranchesHideTheDiscardBranchFromOtherSeats(t *testing.T) {
	g := buildActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	handCardFor(me, "Shock", "Instant")
	src := seatEitherCostSource(g, me.ID)

	a := frameCard(t, g, them.ID.String(), src).ActivatedAbilities[0]
	if len(a.CostBranches) != 2 {
		t.Fatalf("another seat sees %d branches, want both", len(a.CostBranches))
	}
	if d := a.CostBranches[1]; d.DiscardCostOptions != nil || d.Payable || d.DiscardCostN != 1 || d.Label == "" {
		t.Errorf("another seat's discard branch %+v leaks the hand", d)
	}
	if mine := controllerFrameCard(t, g, src).ActivatedAbilities[0].CostBranches[1]; len(mine.DiscardCostOptions) != 1 || !mine.Payable {
		t.Errorf("the controller's discard branch %+v lost its options", mine)
	}
}
