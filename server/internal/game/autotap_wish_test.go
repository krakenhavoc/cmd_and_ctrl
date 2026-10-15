package game

import (
	"testing"

	"github.com/google/uuid"
)

// autotap_wish_test.go — #3853: auto-pay picks the mana that does
// something. The engine half: spend riders on sources and on floating
// mana, colour readers, and the bound that keeps all of it an ordering.
// The printed cards are pinned in cards/effects/autotap_rider_wish_test.go
// and the bot's use of the same planner in
// aiseat/autopay_rider_bot_test.go.

const (
	wishTwinTrigger   = "test-3853-twin"
	wishReadsR        = "test-3853-reads-r"
	wishConverge      = "test-3853-converge"
	wishWantsTreasure = "test-3853-treasure"
)

// withWishHooks installs the catalog declarations the tests' spells make:
// "Red Reader" wants {R}, "Converge" is a converge reader, "Moon" wants
// {R} and {W}. They are keyed by oracle ID, as the catalog is.
func withWishHooks(t *testing.T) {
	t.Helper()
	prevColors, prevDistinct, prevKinds := CatalogWantsColors, CatalogWantsDistinctColors, CatalogWantsManaFrom
	CatalogWantsManaFrom = func(key string) ManaSourceKinds {
		if key == wishWantsTreasure {
			return ManaSourceTreasure
		}
		return 0
	}
	CatalogWantsColors = func(key string) []string {
		switch key {
		case wishReadsR:
			return []string{"R"}
		case wishReadsR + "-rw":
			return []string{"R", "W"}
		}
		return nil
	}
	CatalogWantsDistinctColors = func(key string) bool { return key == wishConverge }
	t.Cleanup(func() {
		CatalogWantsColors, CatalogWantsDistinctColors, CatalogWantsManaFrom = prevColors, prevDistinct, prevKinds
	})
}

// withTwinTrigger registers the conditioned rider the Path of Ancestry
// shape needs: it fires only on a spell named "Twin". Registered once per
// process; the registry is append-only.
func withTwinTrigger() {
	if _, ok := ManaSpendTriggerFor(wishTwinTrigger); ok {
		return
	}
	RegisterManaSpendTrigger(wishTwinTrigger, ManaSpendTrigger{
		Label:     "test — twin",
		Condition: func(_ *Game, _ uuid.UUID, paidFor Card) bool { return paidFor.Name == "Twin" },
		Effect:    func(*Game, *StackItem) error { return nil },
	})
}

// hasteForCreatures is Generator Servant's rider shape: it gives a
// creature spell haste and does nothing for any other spell.
func hasteForCreatures() ManaSpendRider {
	return ManaSpendRider{
		Kind: ManaRiderSpellGains, When: []string{ManaRestrictCast, ManaRestrictType("Creature")},
		Keywords: []string{"haste"}, UntilEndOfTurn: true,
	}
}

func twinRider() ManaSpendRider {
	return ManaSpendRider{Kind: ManaRiderTrigger, Trigger: wishTwinTrigger, When: []string{ManaRestrictCast}}
}

func tappedOnBattlefield(t *testing.T, g *Game, id uuid.UUID) bool {
	t.Helper()
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == id {
			return c.Tapped
		}
	}
	t.Fatalf("%s is not on the battlefield", id)
	return false
}

// autoCast puts a spell in hand and casts it with auto-pay under the
// strict gate, returning its stack item.
func autoCast(t *testing.T, g *Game, p *Player, name, typeLine, cost, oracleID string) *StackItem {
	t.Helper()
	c := NewCard(name, p.ID)
	c.TypeLine = typeLine
	c.ManaCost = cost
	c.OracleID = oracleID
	c.Controller = p.ID
	p.Hand.PushTop(c)
	if err := g.CastSpell(p.ID, c.InstanceID, CastSpellParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("auto-pay cast of %s %s: %v", name, cost, err)
	}
	item := g.StackMeta[c.InstanceID]
	if item == nil {
		t.Fatalf("no stack item for %s", name)
	}
	return item
}

// --- source-side riders ------------------------------------------------

// A source whose rider would fire on this spell is the one auto-pay
// taps, and one whose rider would be wasted on it is left for a spell it
// would do something for. Either order on the battlefield: the rider
// decides, not the order the permanents arrived in.
func TestAutoPayTapsTheSourceWhoseRiderFiresAndHoldsBackTheOneThatWouldnt(t *testing.T) {
	withTwinTrigger()
	for _, tc := range []struct {
		name, spell, typeLine string
		rider                 ManaSpendRider
		wantRider             bool
	}{
		{"a creature spell takes the haste land", "Test Bear", "Creature — Bear", hasteForCreatures(), true},
		{"an instant leaves the haste land alone", "Test Bolt", "Instant", hasteForCreatures(), false},
		{"the named spell takes the conditioned land", "Twin", "Creature — Bear", twinRider(), true},
		{"another spell leaves the conditioned land alone", "Other", "Creature — Bear", twinRider(), false},
	} {
		for _, cost := range []string{"{G}", "{1}"} {
			for _, riderFirst := range []bool{true, false} {
				name := tc.name + " " + cost
				if riderFirst {
					name += " (rider land arrived first)"
				}
				t.Run(name, func(t *testing.T) {
					g := newActiveGame(t)
					advanceTo(t, g, StepPrecombatMain)
					me := g.Seats[0]
					var forest, riderLand uuid.UUID
					push := func() {
						riderLand = pushIntrinsicPermanent(g, me, "Rider Land", "Land",
							[]ManaAbilityShape{{TapCost: true, Produced: "{G}", SpendRiders: []ManaSpendRider{tc.rider}}}, nil)
					}
					if riderFirst {
						push()
					}
					forest = pushBattlefieldForTest(g, me.ID, "Forest", "Basic Land — Forest", "")
					if !riderFirst {
						push()
					}

					autoCast(t, g, me, tc.spell, tc.typeLine, cost, "")

					if got := tappedOnBattlefield(t, g, riderLand); got != tc.wantRider {
						t.Errorf("rider land tapped = %v, want %v", got, tc.wantRider)
					}
					if got := tappedOnBattlefield(t, g, forest); got == tc.wantRider {
						t.Errorf("Forest tapped = %v, want %v", got, !tc.wantRider)
					}
				})
			}
		}
	}
}

// Held back is a TIER, not an exclusion: a board whose only source is a
// rider land that would be wasted still pays the cast.
func TestAHeldBackSourceStillPaysWhenItIsTheOnlyOne(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	riderLand := pushIntrinsicPermanent(g, me, "Rider Land", "Land",
		[]ManaAbilityShape{{TapCost: true, Produced: "{G}", SpendRiders: []ManaSpendRider{hasteForCreatures()}}}, nil)

	autoCast(t, g, me, "Test Bolt", "Instant", "{G}", "")

	if !tappedOnBattlefield(t, g, riderLand) {
		t.Error("the only source on the board was not tapped for the instant")
	}
}

// The thrift tiers outrank a rider: a rider is a bonus, and a bonus does
// not cost a life point. A creature spell with a pain-tier haste land and
// a painless Forest taps the Forest.
func TestARiderDoesNotOutrankTheThriftTiers(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	hall := pushIntrinsicPermanent(g, me, "Hall of the Bandit Lord", "Land",
		[]ManaAbilityShape{{TapCost: true, Produced: "{C}", LifeCost: 3, SpendRiders: []ManaSpendRider{hasteForCreatures()}}}, nil)
	forest := pushBattlefieldForTest(g, me.ID, "Forest", "Basic Land — Forest", "")
	life := me.Life

	autoCast(t, g, me, "Test Bear", "Creature — Bear", "{1}", "")

	if tappedOnBattlefield(t, g, hall) || !tappedOnBattlefield(t, g, forest) || me.Life != life {
		t.Errorf("hall tapped %v, forest tapped %v, life %d -> %d; want the painless Forest",
			tappedOnBattlefield(t, g, hall), tappedOnBattlefield(t, g, forest), life, me.Life)
	}
}

// --- colour readers ------------------------------------------------

// "If {R} was spent" prefers a source that makes {R} for the generic
// half over the Sol Ring the colourless-first instinct reaches for; a
// spell that reads nothing keeps the instinct.
func TestAutoPayConcentratesInTheColourTheSpellReads(t *testing.T) {
	withWishHooks(t)
	for _, tc := range []struct {
		name, oracle string
		wantRed      int
	}{
		{"a reader of {R}", wishReadsR, 1},
		{"a spell that reads nothing", "", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGame(t)
			advanceTo(t, g, StepPrecombatMain)
			me := g.Seats[0]
			pushIntrinsicPermanent(g, me, "Sol Ring", "Artifact", []ManaAbilityShape{{TapCost: true, Produced: "{C}{C}"}}, nil)
			pushBattlefieldForTest(g, me.ID, "Swamp", "Basic Land — Swamp", "")
			pushBattlefieldForTest(g, me.ID, "Mountain", "Basic Land — Mountain", "")
			pushBattlefieldForTest(g, me.ID, "Forest", "Basic Land — Forest", "")

			item := autoCast(t, g, me, "Red Reader", "Sorcery", "{2}{G}", tc.oracle)

			if got := item.Paid.SpentOfColor("R"); got != tc.wantRed {
				t.Errorf("{R} spent = %d, want %d (%+v)", got, tc.wantRed, item.Paid.Mana)
			}
		})
	}
}

// A reader of several colours (Moonhold: {R} and {W}) gets BOTH when the
// board has them, not the first twice.
func TestAutoPayReachesForEachColourTheSpellNames(t *testing.T) {
	withWishHooks(t)
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	pushIntrinsicPermanent(g, me, "Sol Ring", "Artifact", []ManaAbilityShape{{TapCost: true, Produced: "{C}{C}"}}, nil)
	pushBattlefieldForTest(g, me.ID, "Mountain", "Basic Land — Mountain", "")
	pushBattlefieldForTest(g, me.ID, "Mountain", "Basic Land — Mountain", "")
	pushBattlefieldForTest(g, me.ID, "Plains", "Basic Land — Plains", "")
	pushBattlefieldForTest(g, me.ID, "Island", "Basic Land — Island", "")

	item := autoCast(t, g, me, "Moon", "Instant", "{2}{U}", wishReadsR+"-rw")

	if item.Paid.SpentOfColor("R") == 0 || item.Paid.SpentOfColor("W") == 0 {
		t.Errorf("R %d, W %d spent; want both (%+v)", item.Paid.SpentOfColor("R"), item.Paid.SpentOfColor("W"), item.Paid.Mana)
	}
}

// Converge: the generic half is paid with colours the plan does not make
// yet. Three Plains and an Island beside a Sol Ring: the default spends
// the Sol Ring; the converge reader spends a Plains and the Island.
func TestAutoPaySpreadsAConvergeSpellAcrossColours(t *testing.T) {
	withWishHooks(t)
	for _, tc := range []struct {
		name, oracle string
		wantColours  int
	}{
		{"a converge reader", wishConverge, 2},
		{"a spell that reads nothing", "", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGame(t)
			advanceTo(t, g, StepPrecombatMain)
			me := g.Seats[0]
			pushIntrinsicPermanent(g, me, "Sol Ring", "Artifact", []ManaAbilityShape{{TapCost: true, Produced: "{C}{C}"}}, nil)
			for range 3 {
				pushBattlefieldForTest(g, me.ID, "Plains", "Basic Land — Plains", "")
			}
			pushBattlefieldForTest(g, me.ID, "Island", "Basic Land — Island", "")

			item := autoCast(t, g, me, "Converge", "Sorcery", "{2}", tc.oracle)

			if got := item.Paid.ColorsSpentCount(); got != tc.wantColours {
				t.Errorf("colours spent = %d, want %d (%+v)", got, tc.wantColours, item.Paid.Mana)
			}
		})
	}
}

// A dual land recruited for a wished colour makes that colour, not its
// first option. The Forest pays the {G}; the {G|R} land pays the generic
// pip and has to make {R}.
func TestARecruitedDualLandMakesTheWishedColour(t *testing.T) {
	withWishHooks(t)
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	pushBattlefieldForTest(g, me.ID, "Forest", "Basic Land — Forest", "")
	pushIntrinsicPermanent(g, me, "Taiga", "Land", []ManaAbilityShape{{TapCost: true, Produced: "{G|R}"}}, nil)

	item := autoCast(t, g, me, "Red Reader", "Sorcery", "{1}{G}", wishReadsR)

	if got := item.Paid.SpentOfColor("R"); got != 1 {
		t.Errorf("{R} spent = %d, want 1 — the dual land should have made the colour the spell reads (%+v)", got, item.Paid.Mana)
	}
}

// --- the floating pool -------------------------------------------------

func wishToken(color string, r ManaSpendRider) ManaToken {
	return ManaToken{Color: color, Source: uuid.New(), Riders: []ManaSpendRider{r}}
}

// The pool spends the mana whose rider fires on this spell, and keeps it
// from a spell it would be wasted on — in either insertion order.
func TestThePoolSpendsTheRiderMarkedManaOnTheSpellItBenefits(t *testing.T) {
	for _, tc := range []struct {
		name, spell, typeLine string
		wantRider             bool
	}{
		{"a creature spell spends the haste mana", "Test Bear", "Creature — Bear", true},
		{"an instant keeps the haste mana", "Test Bolt", "Instant", false},
	} {
		for _, cost := range []string{"{G}", "{1}"} {
			for _, riderFirst := range []bool{true, false} {
				t.Run(tc.name+" "+cost, func(t *testing.T) {
					g := newActiveGame(t)
					advanceTo(t, g, StepPrecombatMain)
					me := g.Seats[0]
					rider := wishToken("G", hasteForCreatures())
					plain := ManaToken{Color: "G", Source: uuid.New()}
					if riderFirst {
						me.ManaPool.AddMana(rider, plain)
					} else {
						me.ManaPool.AddMana(plain, rider)
					}
					c := NewCard(tc.spell, me.ID)
					c.TypeLine, c.ManaCost, c.Controller = tc.typeLine, cost, me.ID
					me.Hand.PushTop(c)
					if err := g.CastSpell(me.ID, c.InstanceID, CastSpellParams{Strict: true}); err != nil {
						t.Fatalf("cast: %v", err)
					}
					paid := g.StackMeta[c.InstanceID].Paid.Mana
					if len(paid) != 1 {
						t.Fatalf("spent %d tokens, want 1", len(paid))
					}
					if got := paid[0].Source == rider.Source; got != tc.wantRider {
						t.Errorf("spent the rider mana = %v, want %v", got, tc.wantRider)
					}
				})
			}
		}
	}
}

// The conditioned rider (Path of Ancestry's shape) is asked of the spell
// at the spend, through the same Condition the rider fires by.
func TestThePoolAsksAConditionedRiderOfTheSpell(t *testing.T) {
	withTwinTrigger()
	for _, tc := range []struct {
		spell     string
		wantRider bool
	}{{"Twin", true}, {"Other", false}} {
		for _, cost := range []string{"{G}", "{1}"} {
			t.Run(tc.spell+" "+cost, func(t *testing.T) {
				g := newActiveGame(t)
				advanceTo(t, g, StepPrecombatMain)
				me := g.Seats[0]
				rider := wishToken("G", twinRider())
				plain := ManaToken{Color: "G", Source: uuid.New()}
				me.ManaPool.AddMana(plain, rider)
				c := NewCard(tc.spell, me.ID)
				c.TypeLine, c.ManaCost, c.Controller = "Creature — Bear", cost, me.ID
				me.Hand.PushTop(c)
				if err := g.CastSpell(me.ID, c.InstanceID, CastSpellParams{Strict: true}); err != nil {
					t.Fatalf("cast: %v", err)
				}
				paid := g.StackMeta[c.InstanceID].Paid.Mana
				if len(paid) != 1 || (paid[0].Source == rider.Source) != tc.wantRider {
					t.Errorf("spent %+v, want the rider mana = %v", paid, tc.wantRider)
				}
			})
		}
	}
}

// The generic half of a reader's cost is paid from the colour it reads
// first: with {R} and {G} floating, a reader of {G} spends the {G} where
// the default order would have spent the {R}.
func TestThePoolSpendsTheColourTheSpellReadsOnTheGenericHalf(t *testing.T) {
	withWishHooks(t)
	// {R} is already first in the default order (colourless, then W U B R
	// G); a reader of {G} is the case that moves.
	prev := CatalogWantsColors
	CatalogWantsColors = func(key string) []string {
		if key == "test-3853-reads-g" {
			return []string{"G"}
		}
		return nil
	}
	t.Cleanup(func() { CatalogWantsColors = prev })

	for _, tc := range []struct {
		name, oracle, wantSpent string
	}{
		{"a reader of {G}", "test-3853-reads-g", "G"},
		{"a spell that reads nothing", "", "R"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGame(t)
			advanceTo(t, g, StepPrecombatMain)
			me := g.Seats[0]
			floatMana(me, "RG")
			c := NewCard("Reader", me.ID)
			c.TypeLine, c.ManaCost, c.Controller, c.OracleID = "Sorcery", "{1}", me.ID, tc.oracle
			me.Hand.PushTop(c)
			if err := g.CastSpell(me.ID, c.InstanceID, CastSpellParams{Strict: true}); err != nil {
				t.Fatalf("cast: %v", err)
			}
			paid := g.StackMeta[c.InstanceID].Paid.Mana
			if len(paid) != 1 || paid[0].Color != tc.wantSpent {
				t.Errorf("spent %+v, want {%s}", paid, tc.wantSpent)
			}
		})
	}
}

// #1212's wish reaches the pool too: a spell that reads Treasure mana
// spends the floating Treasure mana before a plain mana of the same colour.
func TestThePoolSpendsTheManaFromTheKindTheSpellReads(t *testing.T) {
	withWishHooks(t)
	for _, tc := range []struct {
		name, oracle string
		wantTreasure bool
	}{{"a Treasure reader", wishWantsTreasure, true}, {"a spell that reads nothing", "", false}} {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGame(t)
			advanceTo(t, g, StepPrecombatMain)
			me := g.Seats[0]
			treasure := ManaToken{Color: "B", Source: uuid.New(), SourceKinds: ManaSourceTreasure | ManaSourceArtifact}
			plain := ManaToken{Color: "B", Source: uuid.New(), SourceKinds: ManaSourceLand}
			me.ManaPool.AddMana(plain, treasure)
			c := NewCard("Reader", me.ID)
			c.TypeLine, c.ManaCost, c.Controller, c.OracleID = "Creature — Elf", "{B}", me.ID, tc.oracle
			me.Hand.PushTop(c)
			if err := g.CastSpell(me.ID, c.InstanceID, CastSpellParams{Strict: true}); err != nil {
				t.Fatalf("cast: %v", err)
			}
			paid := g.StackMeta[c.InstanceID].Paid.Mana
			if len(paid) != 1 || (paid[0].Source == treasure.Source) != tc.wantTreasure {
				t.Errorf("spent %+v, want the Treasure mana = %v", paid, tc.wantTreasure)
			}
		})
	}
}

// --- the bound ---------------------------------------------------------

// The whole feature is an ORDERING: with every wish a spell can make,
// every cost is payable exactly when it is payable with none. This is
// what lets the legal-move enumerator ask the same planner for its bool
// and never offer a cast the gate refuses (ADR 0068 A5).
func TestTheWishNeverChangesWhetherACostIsPayable(t *testing.T) {
	withWishHooks(t)
	withTwinTrigger()
	boards := map[string]func(g *Game, me *Player){
		"mixed": func(g *Game, me *Player) {
			pushIntrinsicPermanent(g, me, "Sol Ring", "Artifact", []ManaAbilityShape{{TapCost: true, Produced: "{C}{C}"}}, nil)
			pushBattlefieldForTest(g, me.ID, "Mountain", "Basic Land — Mountain", "")
			pushBattlefieldForTest(g, me.ID, "Forest", "Basic Land — Forest", "")
			pushIntrinsicPermanent(g, me, "Rider Land", "Land",
				[]ManaAbilityShape{{TapCost: true, Produced: "{G}", SpendRiders: []ManaSpendRider{hasteForCreatures()}}}, nil)
			pushIntrinsicPermanent(g, me, "Twin Land", "Land",
				[]ManaAbilityShape{{TapCost: true, Produced: "{W|U|B|R|G}", SpendRiders: []ManaSpendRider{twinRider()}}}, nil)
		},
		"two lands": func(g *Game, me *Player) {
			pushBattlefieldForTest(g, me.ID, "Swamp", "Basic Land — Swamp", "")
			pushBattlefieldForTest(g, me.ID, "Swamp", "Basic Land — Swamp", "")
		},
		"only riders": func(g *Game, me *Player) {
			for range 3 {
				pushIntrinsicPermanent(g, me, "Rider Land", "Land",
					[]ManaAbilityShape{{TapCost: true, Produced: "{G}", SpendRiders: []ManaSpendRider{hasteForCreatures()}}}, nil)
			}
		},
	}
	costs := []string{"{G}", "{1}{G}", "{2}{G}", "{3}{G}", "{R}{R}", "{2}{R}{G}", "{4}", "{5}", "{6}", "{W}", "{1}{W}{U}"}
	spells := []struct{ name, typeLine, oracle string }{
		{"Twin", "Creature — Bear", ""},
		{"Plain", "Instant", ""},
		{"Red Reader", "Sorcery", wishReadsR},
		{"Converge", "Sorcery", wishConverge},
		{"Moon", "Instant", wishReadsR + "-rw"},
	}
	for bname, build := range boards {
		for _, cost := range costs {
			for _, sp := range spells {
				g := newActiveGame(t)
				advanceTo(t, g, StepPrecombatMain)
				me := g.Seats[0]
				build(g, me)
				card := NewCard(sp.name, me.ID)
				card.TypeLine, card.ManaCost, card.OracleID, card.Controller = sp.typeLine, cost, sp.oracle, me.ID
				me.Hand.PushTop(card)
				parsed := costFor(t, cost)

				var bare, wished bool
				var barePlan, wishedPlan tapPlan
				g.WithWriteLock(func() {
					barePlan, _, bare = g.autoTapTopUpLocked(me.ID, parsed, 0, ManaSpendContext{}, nil, 0)
					wishedPlan, _, wished = g.autoTapTopUpLocked(me.ID, parsed, 0, ManaSpendForCast(card), nil, WantedManaSourcesFor(card))
				})
				if bare != wished {
					t.Errorf("%s: %s as %s: payable %v without the wish, %v with it", bname, cost, sp.name, bare, wished)
				}
				// The same set of permanents is in play either way, so the
				// plan can only be a different choice of them.
				_ = barePlan
				_ = wishedPlan
			}
		}
	}
}

// A zero context (the pool probes and previews that name no object) ranks
// nothing, and compares equal for every token: the pool solver and the
// planner are byte-identical to what they were.
func TestAContextThatNamesNoObjectRanksNothing(t *testing.T) {
	var ctx ManaSpendContext
	if got := ctx.tokenRank(wishToken("G", hasteForCreatures())); got != 0 {
		t.Errorf("tokenRank on a zero context = %d, want 0", got)
	}
	if got := ctx.ridersRank([]ManaSpendRider{hasteForCreatures()}, nil); got != 0 {
		t.Errorf("ridersRank on a zero context = %d, want 0", got)
	}
}
