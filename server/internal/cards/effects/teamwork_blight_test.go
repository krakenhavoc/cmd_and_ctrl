package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// teamwork_blight_test.go — #1703: teamwork (CR 702.194a) and blight
// paid as a cost (CR 701.68a), end to end through CastSpell.

const (
	cinderStrikeOracle      = "54421e69-d79e-4c2e-8ce6-96994d168835"
	requitingHexOracle      = "eda16b31-33db-4ad2-901a-80fbe6273733"
	repulsorBlastOracle     = "150f920c-c942-4feb-824f-79dd9b531687"
	twDoublingSeasonOracle  = "01546b7d-a233-4176-8843-d732074dc5b6"
	twWindingConstrictorOra = "c9404d7d-a026-4082-9fcb-1ab571a136b5"
)

// castPaying is castWithOptionalCosts for the two #1703 components:
// the optional costs announced, the teamwork taps and the blighted
// creature.
func castPaying(t *testing.T, g *game.Game, name, typeLine, oracleID string, targets []game.TargetRef,
	optional []int, teamwork, blight []uuid.UUID) (uuid.UUID, error) {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracleID,
		Owner: active.ID, Controller: active.ID})
	advanceToMain(t, g)
	return id, g.CastSpell(active.ID, id, game.CastSpellParams{
		Targets:       targets,
		OptionalCosts: optional,
		TeamworkIDs:   teamwork,
		BlightIDs:     blight,
	})
}

func twCard(g *game.Game, id uuid.UUID) *game.Card { return findBattlefieldCardByID(g, id) }

func twTarget(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

// --- blight -----------------------------------------------------------

func TestCinderStrikeBlightIsPaidOntoYourOwnCreature(t *testing.T) {
	for _, tc := range []struct {
		name     string
		optional []int
		blighted bool
		want     int
	}{
		{"declined", nil, false, 2},
		{"blighted", []int{0}, true, 4},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
			mine := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
			wall := pushVanillaCreature(g, opp.ID, "Wall", 0, 10)
			var blight []uuid.UUID
			if tc.blighted {
				blight = []uuid.UUID{mine}
			}
			if _, err := castPaying(t, g, "Cinder Strike", "Sorcery", cinderStrikeOracle, twTarget(wall),
				tc.optional, nil, blight); err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			// CR 601.2h: the counter is on the creature while the spell
			// is still on the stack.
			wantCounters := 0
			if tc.blighted {
				wantCounters = 1
			}
			if got := twCard(g, mine).Counters[game.CounterMinusOne]; got != wantCounters {
				t.Errorf("-1/-1 counters on the blighted creature at announce: %d, want %d", got, wantCounters)
			}
			passPriorityAroundTable(t, g)
			if got := twCard(g, wall).DamageMarked; got != tc.want {
				t.Errorf("damage: %d, want %d", got, tc.want)
			}
		})
	}
}

// TestBlightThatKillsTheCreatureStillPaysTheCost is CR 701.68b read the
// right way round: the blight is refused only when the counters cannot
// be put, never because the creature would die of them. The 1/1 dies at
// the state-based check and the spell is still blighted.
func TestBlightThatKillsTheCreatureStillPaysTheCost(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	token := pushVanillaCreature(g, me.ID, "Goblin", 1, 1)
	wall := pushVanillaCreature(g, opp.ID, "Wall", 0, 10)
	if _, err := castPaying(t, g, "Cinder Strike", "Sorcery", cinderStrikeOracle, twTarget(wall),
		[]int{0}, nil, []uuid.UUID{token}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	if twCard(g, token) != nil {
		t.Errorf("a 1/1 blighted for 1 is still on the battlefield after the cast's state-based check")
	}
	passPriorityAroundTable(t, g)
	if got := twCard(g, wall).DamageMarked; got != 4 {
		t.Errorf("damage: %d, want 4 — the cost was paid even though the creature died", got)
	}
}

func TestRequitingHexGainsLifeOnlyWhenBlighted(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	mine := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	victim := pushVanillaCreature(g, opp.ID, "Elf", 1, 1)
	before := me.Life
	if _, err := castPaying(t, g, "Requiting Hex", "Instant", requitingHexOracle, twTarget(victim),
		[]int{0}, nil, []uuid.UUID{mine}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if twCard(g, victim) != nil {
		t.Errorf("the target survived")
	}
	if got := me.Life - before; got != 2 {
		t.Errorf("life gained: %d, want 2", got)
	}
}

func TestBlightRefusals(t *testing.T) {
	for _, tc := range []struct {
		name     string
		optional []int
		whose    int // 0 = mine, 1 = an opponent's, -1 = nothing named
		want     error
	}{
		{"an opponent's creature", []int{0}, 1, game.ErrCardCallerMismatch},
		{"announced, nothing named", []int{0}, -1, game.ErrInvalidParam},
		{"named, not announced", nil, 0, game.ErrInvalidParam},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
			mine := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
			theirs := pushVanillaCreature(g, opp.ID, "Wall", 0, 10)
			var blight []uuid.UUID
			switch tc.whose {
			case 0:
				blight = []uuid.UUID{mine}
			case 1:
				blight = []uuid.UUID{theirs}
			}
			id, err := castPaying(t, g, "Cinder Strike", "Sorcery", cinderStrikeOracle, twTarget(theirs),
				tc.optional, nil, blight)
			if !errors.Is(err, tc.want) {
				t.Fatalf("err = %v, want %v", err, tc.want)
			}
			if !me.Hand.Contains(id) {
				t.Errorf("a refused cast left its hand")
			}
			if len(twCard(g, mine).Counters)+len(twCard(g, theirs).Counters) != 0 {
				t.Errorf("a refused cast placed counters")
			}
		})
	}
}

// TestBlightCountersAreACostNotAnEffect is CR 614.16. Doubling Season's
// "if an EFFECT would put" does not reach a cost's counters; Winding
// Constrictor's "if one or more counters would be put" names no effect
// and does — the Devoted Druid + Vizier of Remedies reading, and the
// reason the blight opens the counter window at all.
func TestBlightCountersAreACostNotAnEffect(t *testing.T) {
	for _, tc := range []struct {
		name   string
		oracle string
		want   int
	}{
		{"Doubling Season does not double a cost", twDoublingSeasonOracle, 1},
		{"Winding Constrictor adds one to a cost", twWindingConstrictorOra, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
			pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: tc.name,
				TypeLine: "Enchantment", OracleID: tc.oracle, Owner: me.ID, Controller: me.ID})
			mine := pushVanillaCreature(g, me.ID, "Troll", 5, 5)
			wall := pushVanillaCreature(g, opp.ID, "Wall", 0, 10)
			if _, err := castPaying(t, g, "Cinder Strike", "Sorcery", cinderStrikeOracle, twTarget(wall),
				[]int{0}, nil, []uuid.UUID{mine}); err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			if got := twCard(g, mine).Counters[game.CounterMinusOne]; got != tc.want {
				t.Errorf("-1/-1 counters: %d, want %d", got, tc.want)
			}
		})
	}
}

// --- teamwork ---------------------------------------------------------

func TestRepulsorBlastTeamworkPaidWithEnoughPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	a := pushVanillaCreature(g, me.ID, "Elf A", 1, 1)
	b := pushVanillaCreature(g, me.ID, "Elf B", 1, 1)
	victim := pushVanillaCreature(g, opp.ID, "Ogre", 3, 3)
	before := opp.Life
	if _, err := castPaying(t, g, "Repulsor Blast", "Sorcery", repulsorBlastOracle, twTarget(victim),
		[]int{0}, []uuid.UUID{a, b}, nil); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	if !twCard(g, a).Tapped || !twCard(g, b).Tapped {
		t.Errorf("the teamwork creatures were not tapped at announce")
	}
	passPriorityAroundTable(t, g)
	if twCard(g, victim) != nil {
		t.Errorf("the Ogre survived 5 damage")
	}
	if got := before - opp.Life; got != 2 {
		t.Errorf("damage to the creature's controller: %d, want 2", got)
	}
}

func TestRepulsorBlastWithoutTeamworkSparesTheController(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	a := pushVanillaCreature(g, me.ID, "Elf A", 3, 3)
	victim := pushVanillaCreature(g, opp.ID, "Ogre", 3, 3)
	before := opp.Life
	if _, err := castPaying(t, g, "Repulsor Blast", "Sorcery", repulsorBlastOracle, twTarget(victim),
		nil, nil, nil); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if twCard(g, a).Tapped {
		t.Errorf("a creature tapped for a cast that declined teamwork")
	}
	if opp.Life != before {
		t.Errorf("the controller took damage from an unteamed blast")
	}
}

func TestTeamworkRefusals(t *testing.T) {
	for _, tc := range []struct {
		name  string
		setup func(g *game.Game, me uuid.UUID) (optional []int, team []uuid.UUID)
		want  error
	}{
		{"too little power", func(g *game.Game, me uuid.UUID) ([]int, []uuid.UUID) {
			return []int{0}, []uuid.UUID{pushVanillaCreature(g, me, "Elf", 1, 1)}
		}, game.ErrInsufficientTeamwork},
		{"an already tapped creature", func(g *game.Game, me uuid.UUID) ([]int, []uuid.UUID) {
			id := pushVanillaCreature(g, me, "Ogre", 3, 3)
			twCard(g, id).Tapped = true
			return []int{0}, []uuid.UUID{id}
		}, game.ErrAlreadyTapped},
		{"the same creature twice", func(g *game.Game, me uuid.UUID) ([]int, []uuid.UUID) {
			id := pushVanillaCreature(g, me, "Elf", 1, 1)
			return []int{0}, []uuid.UUID{id, id}
		}, game.ErrInvalidParam},
		{"announced, nothing named", func(g *game.Game, me uuid.UUID) ([]int, []uuid.UUID) {
			pushVanillaCreature(g, me, "Ogre", 3, 3)
			return []int{0}, nil
		}, game.ErrInsufficientTeamwork},
		{"named, not announced", func(g *game.Game, me uuid.UUID) ([]int, []uuid.UUID) {
			return nil, []uuid.UUID{pushVanillaCreature(g, me, "Ogre", 3, 3)}
		}, game.ErrInvalidParam},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
			victim := pushVanillaCreature(g, opp.ID, "Ogre", 3, 3)
			optional, team := tc.setup(g, me.ID)
			tappedBefore := map[uuid.UUID]bool{}
			for _, id := range team {
				tappedBefore[id] = twCard(g, id).Tapped
			}
			id, err := castPaying(t, g, "Repulsor Blast", "Sorcery", repulsorBlastOracle, twTarget(victim),
				optional, team, nil)
			if !errors.Is(err, tc.want) {
				t.Fatalf("err = %v, want %v", err, tc.want)
			}
			if !me.Hand.Contains(id) {
				t.Errorf("a refused cast left its hand")
			}
			for _, c := range team {
				if twCard(g, c).Tapped != tappedBefore[c] {
					t.Errorf("a refused cast tapped a creature")
				}
			}
		})
	}
}

// TestTeamworkAllowsSummoningSickCreatures: teamwork taps creatures
// but is not the {T} symbol, so CR 302.6 does not apply — crew's
// CR 702.122 reading, one keyword over.
func TestTeamworkAllowsSummoningSickCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	sick := pushVanillaCreature(g, me.ID, "Just cast", 2, 2)
	twCard(g, sick).SummonedThisTurn = true
	victim := pushVanillaCreature(g, opp.ID, "Ogre", 3, 3)
	if _, err := castPaying(t, g, "Repulsor Blast", "Sorcery", repulsorBlastOracle, twTarget(victim),
		[]int{0}, []uuid.UUID{sick}, nil); err != nil {
		t.Fatalf("a summoning-sick creature was refused for teamwork: %v", err)
	}
}

// TestTeamworkReadsEffectivePower: a +1/+1 counter counts, so a 1/1
// with one pays teamwork 2 alone.
func TestTeamworkReadsEffectivePower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	elf := pushVanillaCreature(g, me.ID, "Elf", 1, 1)
	twCard(g, elf).Counters = map[string]int{game.CounterPlusOne: 1}
	victim := pushVanillaCreature(g, opp.ID, "Ogre", 3, 3)
	if _, err := castPaying(t, g, "Repulsor Blast", "Sorcery", repulsorBlastOracle, twTarget(victim),
		[]int{0}, []uuid.UUID{elf}, nil); err != nil {
		t.Fatalf("a 2-power (1/1 + counter) creature was refused for teamwork 2: %v", err)
	}
}

// TestTeamworkAndBlightSurviveUndoAndRestore: the paid record lives on
// the stack item, so a restored or undone table still resolves the
// spell as cast using teamwork, and undoing the cast untaps the team.
func TestTeamworkAndBlightSurviveUndoAndRestore(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	a := pushVanillaCreature(g, me.ID, "Ogre", 3, 3)
	victim := pushVanillaCreature(g, opp.ID, "Ogre", 3, 3)
	advanceToMain(t, g)
	before := g.Clone()
	lifeBefore := opp.Life
	if _, err := castPaying(t, g, "Repulsor Blast", "Sorcery", repulsorBlastOracle, twTarget(victim),
		[]int{0}, []uuid.UUID{a}, nil); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}

	// Undo: the whole cast goes, and the teamwork creature is untapped.
	undone := g.Clone()
	undone.RestoreFrom(before)
	if twCard(undone, a).Tapped {
		t.Errorf("undoing the cast left the teamwork creature tapped")
	}

	// Restore with the spell on the stack: it still resolves as cast
	// using teamwork.
	restored := restoreRoundTrip(t, g, false)
	passPriorityAroundTable(t, restored)
	var p *game.Player
	for _, s := range restored.Seats {
		if s.ID == opp.ID {
			p = s
		}
	}
	if got := lifeBefore - p.Life; got != 2 {
		t.Errorf("after a restore, damage to the controller: %d, want 2 (the teamwork record was lost)", got)
	}
}

func TestRegisterRefusesAMalformedTeamworkOrBlight(t *testing.T) {
	mustPanic(t, "mandatory AdditionalCost slot", func() {
		Register(Spec{OracleID: "tw-test-mandatory", Name: "Mandatory",
			AdditionalCost: &game.AdditionalCost{Teamwork: 2}})
	})
	mustPanic(t, "declares a teamwork cost keyed", func() {
		Register(Spec{OracleID: "tw-test-key", Name: "Key",
			OptionalCosts: []game.AdditionalCost{{Optional: true, Key: game.KickerKey, Teamwork: 2}}})
	})
	mustPanic(t, "two teamwork or two blight", func() {
		b := OptionalBlight(1)
		b2 := OptionalBlight(2)
		b2.Key = game.BlightKey
		Register(Spec{OracleID: "tw-test-two", Name: "Two",
			OptionalCosts: []game.AdditionalCost{b, b2}})
	})
	mustPanic(t, "mixes teamwork / blight", func() {
		b := OptionalBlight(1)
		b.ManaCost = "{1}"
		Register(Spec{OracleID: "tw-test-mixed", Name: "Mixed",
			OptionalCosts: []game.AdditionalCost{b}})
	})
}
