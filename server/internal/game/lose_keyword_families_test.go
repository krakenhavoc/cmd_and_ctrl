package game

import (
	"testing"

	"github.com/google/uuid"
)

// lose_keyword_families_test.go — #5948: ModRemoveKeywordPrefixes and
// ModSuppressWard, the two records behind Shay Cormac's "permanents your
// opponents control lose hexproof, indestructible, protection, shroud, and
// ward until end of turn". The catalog half is in cards/effects.

func TestRemoveKeywordPrefixesDropsEveryTokenOfTheFamily(t *testing.T) {
	g := newActiveGame(t)
	id := pushScopedTestCreature(g, g.Seats[0].ID, 2, 2)
	registerScopedEffectForTest(t, g, id, []Mod{AddKeywordsMod("protection from red", "protection from black", "hexproof from white", "flying", "hexproof")}, IndefiniteDuration())
	registerScopedEffectForTest(t, g, id, []Mod{RemoveKeywordPrefixesMod("protection from ", "hexproof from ")}, IndefiniteDuration())
	got := abilitiesOf(t, g, id)
	for _, a := range got {
		if a != "flying" && a != "hexproof" {
			t.Errorf("abilities = %v: %q should have been removed", got, a)
		}
	}
	if !hasAbility(got, "flying") || !hasAbility(got, "hexproof") {
		t.Errorf("abilities = %v: plain hexproof and flying must stay (only the families go)", got)
	}
}

func TestRemoveKeywordPrefixesIsTimestampOrdered(t *testing.T) {
	g := newActiveGame(t)
	id := pushScopedTestCreature(g, g.Seats[0].ID, 2, 2)
	registerScopedEffectForTest(t, g, id, []Mod{RemoveKeywordPrefixesMod("protection from ")}, IndefiniteDuration())
	registerScopedEffectForTest(t, g, id, []Mod{AddKeywordsMod("protection from red")}, IndefiniteDuration())
	if !hasAbility(abilitiesOf(t, g, id), "protection from red") {
		t.Error("a protection granted after the removal must stay (CR 613.7: it is a plain removal, not a can't-have)")
	}
}

func TestSuppressWardCoversOnlyTheRecordsSet(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	covered := pushScopedTestCreature(g, me.ID, 2, 2)
	other := pushScopedTestCreature(g, me.ID, 2, 2)
	registerScopedEffectForTest(t, g, covered, []Mod{SuppressWardMod()}, IndefiniteDuration())
	suppressed := func(id uuid.UUID) bool {
		out := false
		g.WithWriteLock(func() {
			for i := range g.Battlefield.Cards {
				if g.Battlefield.Cards[i].InstanceID == id {
					out = g.WardSuppressedForEffect(&g.Battlefield.Cards[i])
				}
			}
		})
		return out
	}
	if !suppressed(covered) {
		t.Error("ward is not suppressed for the permanent the record names")
	}
	if suppressed(other) {
		t.Error("ward is suppressed for a permanent the record does not name")
	}
}

func TestLoseKeywordFamiliesModsAreChecked(t *testing.T) {
	if loseKeywordFamiliesModProblem(RemoveKeywordPrefixesMod()) == "" {
		t.Error("a removeKeywordPrefixes mod with no prefix is accepted")
	}
	if loseKeywordFamiliesModProblem(Mod{Kind: ModSuppressWard, Keywords: []string{"ward"}}) == "" {
		t.Error("a suppressWard mod that names a keyword is accepted")
	}
	if loseKeywordFamiliesModProblem(SuppressWardMod()) != "" || loseKeywordFamiliesModProblem(RemoveKeywordPrefixesMod("protection from ")) != "" {
		t.Error("a sound mod is refused")
	}
}
