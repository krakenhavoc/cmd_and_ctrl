package game

import "testing"

// land_plays_unlimited_test.go — #6213, Fastbond: "You may play any
// number of lands on each of your turns. / Whenever you play a land, if
// it wasn't the first land you played this turn, this enchantment deals
// 1 damage to you." CR 305.2: "A player can normally play one land
// during their turn; however, continuous effects may increase this
// number."

// Under an unlimited grant a player plays three lands, and each played
// entry says which land play of the turn it was: the second and third
// are the two "wasn't the first" a Fastbond trigger counts.
func TestUnlimitedLandPlaysAndTheLandOrdinal(t *testing.T) {
	withCatalogAdditionalLandPlays(t, func(oracleID string) int {
		if oracleID == "fastbond" {
			return UnlimitedLandPlays
		}
		return 0
	})
	g := newWrapGame(t, 2)
	advanceToMainByPriority(t, g)
	active := g.Seats[g.Turn.ActiveSeat]
	pushTappedPermanent(g, active.ID, "Fastbond", "fastbond", "Enchantment", false)
	// A second grant on top must not overflow the ceiling.
	g.WithWriteLock(func() { g.GrantAdditionalLandPlayForEffect(active.ID, 3) })
	if got := g.EffectiveLandDropsLocked(active); got != UnlimitedLandPlays {
		t.Fatalf("allowance = %d, want UnlimitedLandPlays", got)
	}

	start := len(g.Events)
	for i := 0; i < 3; i++ {
		if err := dropLandFromHand(t, g, active); err != nil {
			t.Fatalf("land play %d: %v", i+1, err)
		}
	}
	var ordinals []int
	for _, ev := range g.Events[start:] {
		if ev.Kind == EventZoneMove && ev.Played {
			ordinals = append(ordinals, ev.LandOrdinal)
		}
	}
	if len(ordinals) != 3 || ordinals[0] != 1 || ordinals[1] != 2 || ordinals[2] != 3 {
		t.Fatalf("land ordinals = %v, want [1 2 3]", ordinals)
	}
	notFirst := 0
	for _, o := range ordinals {
		if o > 1 {
			notFirst++
		}
	}
	if notFirst != 2 {
		t.Errorf("lands that weren't the first = %d, want 2 (Fastbond deals 2)", notFirst)
	}
	if got := g.LandDropsRemainingFor(active.ID); got <= 0 {
		t.Errorf("land drops remaining = %d, want still open", got)
	}
}

// A land put onto the battlefield is not a play and carries no ordinal.
func TestPutLandCarriesNoOrdinal(t *testing.T) {
	g := newWrapGame(t, 2)
	advanceToMainByPriority(t, g)
	active := g.Seats[g.Turn.ActiveSeat]
	if err := dropLandFromHand(t, g, active); err != nil {
		t.Fatal(err)
	}
	for _, ev := range g.Events {
		if ev.Kind == EventZoneMove && !ev.Played && ev.LandOrdinal != 0 {
			t.Fatalf("a non-play entry carries ordinal %d", ev.LandOrdinal)
		}
	}
}
