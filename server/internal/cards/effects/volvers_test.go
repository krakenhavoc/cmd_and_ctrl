package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// volvers_test.go — #2360: per-kicker entry counters and the linked
// "enters with <ability>" clauses.

type volverCase struct {
	name, oracle                  string
	first, second                 string // kicker costs, printed order
	firstAbility, secondAbility   string // a keyword, "lifegain" or "regenerate"
	firstCounters, secondCounters int
}

var volverCases = []volverCase{
	{"Necravolver", "639d16ed-daa4-4c8d-8550-5c9a29f4d9fe", "{1}{G}", "{W}", "trample", "lifegain", 2, 1},
	{"Degavolver", "9431c81f-7b71-4c5d-a63c-1b5aba03601b", "{1}{B}", "{R}", "regenerate", "first strike", 2, 1},
	{"Rakavolver", "c00adb98-81dc-490f-9e0f-7f8546ef22ce", "{1}{W}", "{U}", "lifegain", "flying", 2, 1},
	{"Anavolver", "03a4cbe2-cf35-4c2c-b035-666f994ffd09", "{1}{U}", "{B}", "flying", "regenerate", 2, 1},
	{"Cetavolver", "3ae86510-f828-48f8-adab-759df4b5544e", "{1}{R}", "{G}", "first strike", "trample", 2, 1},
}

func volverHas(t *testing.T, g *game.Game, id uuid.UUID, ability string) bool {
	t.Helper()
	for _, a := range effectiveAbilities(t, g, id) {
		if a == ability {
			return true
		}
	}
	return false
}

func castVolver(t *testing.T, g *game.Game, vc volverCase, optional []int) uuid.UUID {
	t.Helper()
	id, err := castWithOptionalCosts(t, g, vc.name, "Creature — Volver", vc.oracle, nil, optional, nil)
	if err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if findBattlefieldCardByID(g, id) == nil {
		t.Fatalf("%s did not enter", vc.name)
	}
	return id
}

func TestVolversEachKickerCombination(t *testing.T) {
	for _, vc := range volverCases {
		for _, tc := range []struct {
			name     string
			optional []int
		}{{"unkicked", nil}, {"first", []int{0}}, {"second", []int{1}}, {"both", []int{0, 1}}} {
			t.Run(vc.name+"/"+tc.name, func(t *testing.T) {
				g := newCatalogGame(t)
				me := g.Seats[g.Turn.ActiveSeat]
				id := castVolver(t, g, vc, tc.optional)
				k1, k2 := false, false
				for _, i := range tc.optional {
					k1, k2 = k1 || i == 0, k2 || i == 1
				}
				want := 0
				if k1 {
					want += vc.firstCounters
				}
				if k2 {
					want += vc.secondCounters
				}
				if got := plusOneCounters(g, id); got != want {
					t.Errorf("+1/+1 counters = %d, want %d", got, want)
				}
				// A keyword is present exactly when the kicker that
				// links to it was paid.
				for _, kw := range []string{"trample", "flying", "first strike"} {
					linked := (k1 && vc.firstAbility == kw) || (k2 && vc.secondAbility == kw)
					if volverHas(t, g, id, kw) != linked {
						t.Errorf("keyword %q present = %v, want %v", kw, !linked, linked)
					}
				}
				for _, a := range []struct {
					on      bool
					ability string
				}{{k1, vc.firstAbility}, {k2, vc.secondAbility}} {
					switch a.ability {
					case "regenerate":
						life := me.Life
						err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{})
						if a.on {
							if err != nil {
								t.Fatalf("regenerate: %v", err)
							}
							if me.Life != life-3 {
								t.Errorf("life %d -> %d, want 3 paid", life, me.Life)
							}
						} else if err == nil || me.Life != life {
							t.Errorf("a Volver not kicked with this cost activated regenerate (err=%v)", err)
						}
					case "lifegain":
						life := me.Life
						g.WithWriteLock(func() {
							_ = g.DealDamageToPlayerForEffect(id, g.Seats[1].ID, 2)
						})
						passPriorityAroundTable(t, g)
						gained := me.Life - life
						if a.on && gained != 2 || !a.on && gained != 0 {
							t.Errorf("gained %d life from 2 damage, linked kicker paid = %v", gained, a.on)
						}
					}
				}
			})
		}
	}
}

// CR 614.1c: the counters are part of the entry, so Doubling Season
// doubles each kicker's clause: 2 and 1 become 4 and 2.
func TestVolverCountersAreDoubledByDoublingSeason(t *testing.T) {
	vc := volverCases[0]
	for _, tc := range []struct {
		optional []int
		want     int
	}{{[]int{0}, 4}, {[]int{1}, 2}, {[]int{0, 1}, 6}} {
		g := newCatalogGame(t)
		_ = seedReplacementPermanent(g, doublingSeasonOracle, "Doubling Season", g.Seats[g.Turn.ActiveSeat].ID)
		id := castVolver(t, g, vc, tc.optional)
		if got := plusOneCounters(g, id); got != tc.want {
			t.Errorf("kicked %v: counters = %d, want %d", tc.optional, got, tc.want)
		}
	}
}

// CR 400.7d: a Volver that was kicked, died and was reanimated is a
// new object that was never cast, so it has no counters and no
// abilities.
func TestAReanimatedVolverWasNotKicked(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := castVolver(t, g, volverCases[0], []int{0, 1}) // Necravolver
	if !volverHas(t, g, id, "trample") {
		t.Fatal("a Necravolver kicked with {1}{G} lacks trample")
	}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(id); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	b30Reanimate(t, g, id, me.ID)
	if got := plusOneCounters(g, id); got != 0 {
		t.Errorf("reanimated counters = %d, want 0", got)
	}
	if volverHas(t, g, id, "trample") {
		t.Error("a reanimated Necravolver still has trample")
	}
	c := findBattlefieldCardByID(g, id)
	if game.CardKickedWith(*c, "{1}{G}") || game.CardKickedWith(*c, "{W}") {
		t.Error("a reanimated Volver remembers being kicked")
	}
}

func TestCastCountsKickedWithIsFalseForAnUnpaidCost(t *testing.T) {
	cc := game.CastCounts{KickersPaid: []string{"{W}"}}
	if !cc.KickedWith("{W}") || cc.KickedWith("{1}{G}") || cc.KickedWith("") {
		t.Errorf("KickedWith misreads %+v", cc)
	}
}
