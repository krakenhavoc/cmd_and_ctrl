package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #4380: real catalog rows and choices, including resolution ordering,
// ownership, and restored keyed prompts. Neyith is held for #4384.
func jumpOracle(t *testing.T, name string) string {
	t.Helper()
	for _, s := range All() {
		if s.Name == name {
			return s.OracleID
		}
	}
	t.Fatalf("missing %s", name)
	return ""
}
func jumpCard(t *testing.T, g *game.Game, who uuid.UUID, name, typ string, power, toughness int) uuid.UUID {
	t.Helper()
	return b12Push(g, who, name, typ, jumpOracle(t, name), power, toughness)
}
func jumpTargets(ids ...uuid.UUID) []game.TargetRef {
	out := make([]game.TargetRef, len(ids))
	for i, id := range ids {
		out[i] = game.TargetRef{Kind: game.TargetCard, ID: id, Slot: i}
	}
	return out
}
func jumpEvent(g *game.Game, ev game.Event) { g.WithWriteLock(func() { g.EmitEvent(ev) }) }
func jumpCounter(t *testing.T, g *game.Game, id uuid.UUID, kind string, n int) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(id, kind, n); err != nil {
			t.Fatal(err)
		}
	})
}

func TestJumpstartPumpSpells(t *testing.T) {
	for _, tc := range []struct {
		name, keyword              string
		power, toughness, counters int
	}{{"Arbor Armament", "reach", 3, 3, 1}, {"Sure Strike", "first strike", 5, 2, 0}, {"Primal Might", "", 5, 5, 0}} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			id := pushVanillaCreature(g, me.ID, "Recipient", 2, 2)
			if tc.name == "Primal Might" {
				castXSpell(t, g, tc.name, "Sorcery", jumpOracle(t, tc.name), "", 3, jumpTargets(id))
			} else {
				castCatalogSpell(t, g, tc.name, "Instant", jumpOracle(t, tc.name), jumpTargets(id))
			}
			passPriorityAroundTable(t, g)
			c := gaLayered(t, g, id)
			if c.CurrentPower() != tc.power || c.CurrentToughness() != tc.toughness || c.Counters[game.CounterPlusOne] != tc.counters {
				t.Fatalf("pump/counters: %+v", c)
			}
			if tc.keyword != "" && !hasEffectiveKeyword(t, g, id, tc.keyword) {
				t.Fatalf("missing %s", tc.keyword)
			}
		})
	}
}

func TestJumpstartCounterThenDamageAndFight(t *testing.T) {
	for _, name := range []string{"Hunter's Edge", "Savage Stomp", "Time to Feed"} {
		t.Run(name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			mine := pushVanillaCreature(g, me.ID, "Fighter", 3, 5)
			theirs := pushVanillaCreature(g, opp.ID, "Other", 2, 4)
			targets := jumpTargets(mine, theirs)
			if name == "Time to Feed" {
				targets = jumpTargets(theirs, mine)
			}
			life := me.Life
			castCatalogSpell(t, g, name, "Sorcery", jumpOracle(t, name), targets)
			passPriorityAroundTable(t, g)
			if name == "Time to Feed" {
				if !battlefieldHas(g, theirs) {
					t.Fatal("3 power should not kill a 4 toughness target")
				}
				g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(theirs) })
				passPriorityAroundTable(t, g)
				if me.Life != life+3 {
					t.Fatal("later death did not gain three life")
				}
			} else {
				c := gaLayered(t, g, mine)
				if c.Counters[game.CounterPlusOne] != 1 || battlefieldHas(g, theirs) {
					t.Fatal("counter must precede power damage")
				}
				want := 0
				if name == "Savage Stomp" {
					want = 2
				}
				if c.DamageMarked != want {
					t.Fatal("bite and fight confused")
				}
			}
		})
	}
}

func TestJumpstartFightSpellKeepsCounterWhenOtherTargetLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := pushVanillaCreature(g, me.ID, "Mine", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Theirs", 2, 2)
	castCatalogSpell(t, g, "Savage Stomp", "Sorcery", jumpOracle(t, "Savage Stomp"), jumpTargets(mine, theirs))
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(theirs) })
	passPriorityAroundTable(t, g)
	if gaLayered(t, g, mine).Counters[game.CounterPlusOne] != 1 || gaLayered(t, g, mine).DamageMarked != 0 {
		t.Fatal("legal first clause did not get its counter independently")
	}
}

func TestJumpstartFungalRebirthMorbid(t *testing.T) {
	for _, dead := range []bool{false, true} {
		t.Run(map[bool]string{false: "no death", true: "death"}[dead], func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			target := gyCardWithOracle(me, "Forest", "Basic Land — Forest", "")
			if dead {
				id := pushVanillaCreature(g, g.Seats[1].ID, "Doomed", 1, 1)
				g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(id) })
			}
			castCatalogSpell(t, g, "Fungal Rebirth", "Instant", jumpOracle(t, "Fungal Rebirth"), jumpTargets(target))
			passPriorityAroundTable(t, g)
			if !me.Hand.Contains(target) {
				t.Fatal("permanent card not returned")
			}
			want := 0
			if dead {
				want = 2
			}
			if countBattlefieldNamed(g, me.ID, "Saproling") != want {
				t.Fatal("morbid token count")
			}
		})
	}
}

func TestJumpstartBallLightningSacrificesAtAnyEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := jumpCard(t, g, me.ID, "Ball Lightning", "Creature — Elemental", 6, 1)
	if !hasEffectiveKeyword(t, g, id, "haste") || !hasEffectiveKeyword(t, g, id, "trample") {
		t.Fatal("printed keywords missing")
	}
	jumpEvent(g, game.Event{Kind: game.EventBeginEndStep, Actor: g.Seats[1].ID})
	passPriorityAroundTable(t, g)
	if battlefieldHas(g, id) || !me.Graveyard.Contains(id) {
		t.Fatal("did not sacrifice on opponent's end step")
	}
}

func TestJumpstartBallLightningCannotSacrificeAnOpponentsPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := jumpCard(t, g, me.ID, "Ball Lightning", "Creature — Elemental", 6, 1)
	jumpEvent(g, game.Event{Kind: game.EventBeginEndStep, Actor: me.ID})
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].Controller = opp.ID
			}
		}
	})
	passPriorityAroundTable(t, g)
	if !battlefieldHas(g, id) {
		t.Fatal("the old controller sacrificed the new controller's Ball Lightning")
	}
}

func TestJumpstartCombatTriggers(t *testing.T) {
	for _, name := range []string{"Rampaging Brontodon", "Ornery Dilophosaur"} {
		t.Run(name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			id := jumpCard(t, g, me.ID, name, "Creature — Dinosaur", 2, 2)
			if name == "Rampaging Brontodon" {
				for i := 0; i < 3; i++ {
					b12Permanent(g, me.ID, "Land", "Land")
				}
			} else {
				pushVanillaCreature(g, me.ID, "Big", 4, 4)
			}
			jumpEvent(g, game.Event{Kind: game.EventAttack, Actor: me.ID, CardID: id, Source: id, Target: g.Seats[1].ID})
			passPriorityAroundTable(t, g)
			want := 4
			if name == "Rampaging Brontodon" {
				want = 5
			}
			if gaLayered(t, g, id).CurrentPower() != want {
				t.Fatal("attack pump")
			}
		})
	}
}

func TestJumpstartOrneryRechecksItsInterveningCondition(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := jumpCard(t, g, me.ID, "Ornery Dilophosaur", "Creature — Dinosaur", 2, 2)
	big := pushVanillaCreature(g, me.ID, "Big", 4, 4)
	jumpEvent(g, game.Event{Kind: game.EventAttack, Actor: me.ID, CardID: id, Source: id, Target: g.Seats[1].ID})
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(big) })
	passPriorityAroundTable(t, g)
	if gaLayered(t, g, id).CurrentPower() != 2 {
		t.Fatal("lost condition still pumped")
	}
}

func TestJumpstartWoodbornThresholdTracksLandChanges(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := jumpCard(t, g, me.ID, "Woodborn Behemoth", "Creature — Elemental", 4, 4)
	for i := 0; i < 7; i++ {
		b12Permanent(g, me.ID, "Land", "Land")
	}
	if gaLayered(t, g, id).CurrentPower() != 4 {
		t.Fatal("active with seven lands")
	}
	eighth := b12Permanent(g, me.ID, "Eighth", "Land")
	if gaLayered(t, g, id).CurrentPower() != 8 || !hasEffectiveKeyword(t, g, id, "trample") {
		t.Fatal("eight-land boost missing")
	}
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(eighth) })
	if gaLayered(t, g, id).CurrentPower() != 4 || hasEffectiveKeyword(t, g, id, "trample") {
		t.Fatal("threshold not removed")
	}
}

func TestJumpstartActivatedCombatCards(t *testing.T) {
	for _, name := range []string{"Hellkite Punisher", "Suspicious Bookcase"} {
		t.Run(name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			advanceToMain(t, g)
			src := jumpCard(t, g, me.ID, name, "Creature", 2, 3)
			target := pushVanillaCreature(g, me.ID, "Attacker", 2, 2)
			p := game.ActivateAbilityParams{Strict: true}
			mana := "{R}"
			if name == "Suspicious Bookcase" {
				p.Targets = jumpTargets(target)
				mana = "{C}{C}{C}"
			}
			_ = g.AddManaForEffect(me.ID, uuid.Nil, mana)
			if err := g.ActivateCatalogAbility(me.ID, src, 0, p); err != nil {
				t.Fatal(err)
			}
			passPriorityAroundTable(t, g)
			if name == "Hellkite Punisher" {
				if gaLayered(t, g, src).CurrentPower() != 3 || !hasEffectiveKeyword(t, g, src, "flying") {
					t.Fatal("firebreathing")
				}
			} else {
				if gaLayered(t, g, target).Effective().Restrictions&game.CantBeBlocked == 0 || !gaLayered(t, g, src).Tapped {
					t.Fatal("Bookcase activation")
				}
			}
		})
	}
}

func TestJumpstartBattleRattleOptionalPump(t *testing.T) {
	for _, yes := range []bool{false, true} {
		t.Run(map[bool]string{false: "decline", true: "accept"}[yes], func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			jumpCard(t, g, me.ID, "Battle-Rattle Shaman", "Creature — Goblin Shaman", 2, 2)
			target := pushVanillaCreature(g, g.Seats[1].ID, "Opponent's creature", 2, 2)
			jumpEvent(g, game.Event{Kind: game.EventStepBegan, Actor: me.ID, Step: game.StepBeginCombat})
			answerTriggerTargets(t, g, me.ID, target)
			answerLatestTriggerPrompt(t, g, me.ID, yes)
			passPriorityAroundTable(t, g)
			want := 2
			if yes {
				want = 4
			}
			if gaLayered(t, g, target).CurrentPower() != want {
				t.Fatal("optional target pump")
			}
		})
	}
}

func TestJumpstartFurorAndNewHorizons(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	creature := pushVanillaCreature(g, me.ID, "Creature", 2, 2)
	land := b12Permanent(g, me.ID, "Land", "Land")
	castCatalogSpell(t, g, "Furor of the Bitten", "Enchantment — Aura", jumpOracle(t, "Furor of the Bitten"), jumpTargets(creature))
	passPriorityAroundTable(t, g)
	if gaLayered(t, g, creature).CurrentPower() != 4 || len(gaLayered(t, g, creature).Effective().AttackRequirements) == 0 {
		t.Fatal("Furor pump/attack requirement")
	}
	castCatalogSpell(t, g, "New Horizons", "Enchantment — Aura", jumpOracle(t, "New Horizons"), jumpTargets(land))
	passPriorityAroundTable(t, g)
	answerTriggerTargets(t, g, me.ID, creature)
	passPriorityAroundTable(t, g)
	if gaLayered(t, g, creature).Counters[game.CounterPlusOne] != 1 {
		t.Fatal("Horizons entry counter")
	}
	if err := tapGrantedMana(t, g, me.ID, land, "U"); err != nil {
		t.Fatal(err)
	}
	if got := me.ManaPool.Total(); got != 2 {
		t.Fatalf("Horizons mana = %d, want 2", got)
	}
}

func TestJumpstartFertilidTargetSearchRestores(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	src := jumpCard(t, g, me.ID, "Fertilid", "Creature — Elemental", 0, 0)
	jumpCounter(t, g, src, game.CounterPlusOne, 2)
	ids := seedSearchLibrary(opp, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"}, game.Card{Name: "Other", TypeLine: "Creature — Elf"})
	_ = g.AddManaForEffect(me.ID, uuid.Nil, "{G}{C}")
	if err := g.ActivateCatalogAbility(me.ID, src, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}, Strict: true}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	g = restoreRoundTrip(t, g, true)
	answerSearchByID(t, g, opp.ID, ids[0])
	passPriorityAroundTable(t, g)
	if c := gaLayered(t, g, ids[0]); c.Controller != opp.ID || !c.Tapped {
		t.Fatal("target player did not receive tapped basic")
	}
	if gaLayered(t, g, src).Counters[game.CounterPlusOne] != 1 {
		t.Fatal("counter cost")
	}
}

func TestJumpstartCharmbreakerReturnsOnlyEligibleCardAndPumps(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	src := jumpCard(t, g, me.ID, "Charmbreaker Devils", "Creature — Devil", 4, 4)
	spell := gyCardWithOracle(me, "Old spell", "Instant", "")
	other := gyCardWithOracle(me, "Creature", "Creature", "")
	jumpEvent(g, game.Event{Kind: game.EventBeginUpkeep, Actor: me.ID})
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(spell) || !me.Graveyard.Contains(other) {
		t.Fatal("random eligible return")
	}
	castCatalogSpell(t, g, "Arbor Armament", "Instant", jumpOracle(t, "Arbor Armament"), jumpTargets(src))
	passPriorityAroundTable(t, g)
	if gaLayered(t, g, src).CurrentPower() != 9 {
		t.Fatal("instant cast should pump +4/+0 as well as the spell's counter")
	}
}

func TestJumpstartHamletbackReadsEnteringCreaturesLastKnownPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := jumpCard(t, g, me.ID, "Hamletback Goliath", "Creature — Giant Warrior", 6, 6)
	other := pushVanillaCreature(g, opp.ID, "Entering", 4, 4)
	jumpEvent(g, game.Event{Kind: game.EventETB, Actor: opp.ID, CardID: other, Source: other})
	jumpCounter(t, g, other, game.CounterPlusOne, 2)
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(other) })
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if gaLayered(t, g, src).Counters[game.CounterPlusOne] != 6 {
		t.Fatal("entering creature's power was not read at resolution with LKI")
	}
}

func TestJumpstartHuntersInsightWatchesDamageNotGrantedAbilities(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushVanillaCreature(g, me.ID, "Source", 3, 3)
	pw := b12Permanent(g, opp.ID, "Walker", "Planeswalker — Test")
	jumpCounter(t, g, pw, game.CounterLoyalty, 5)
	creature := pushVanillaCreature(g, opp.ID, "Other", 2, 2)
	castCatalogSpell(t, g, "Hunter's Insight", "Instant", jumpOracle(t, "Hunter's Insight"), jumpTargets(src))
	passPriorityAroundTable(t, g)
	before := me.Hand.Size()
	for _, ev := range []game.Event{{Kind: game.EventDealDamage, Source: src, Target: opp.ID, Amount: 3, Combat: true}, {Kind: game.EventDealDamage, Source: src, Target: pw, Amount: 2, Combat: true}, {Kind: game.EventDealDamage, Source: src, Target: creature, Amount: 8, Combat: true}, {Kind: game.EventDealDamage, Source: src, Target: opp.ID, Amount: 8, Combat: false}} {
		jumpEvent(g, ev)
		passPriorityAroundTable(t, g)
	}
	if me.Hand.Size() != before+5 {
		t.Fatal("must draw only for combat damage to player/planeswalker")
	}
	g.WithWriteLock(func() {
		_ = g.SacrificePermanentForEffect(src)
		_ = g.ReturnFromGraveyardUnderControlForEffect(src, game.ZoneBattlefield, me.ID)
	})
	jumpEvent(g, game.Event{Kind: game.EventDealDamage, Source: src, Target: opp.ID, Amount: 3, Combat: true})
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != before+5 {
		t.Fatal("delayed trigger followed a new object")
	}
}

func TestJumpstartTraitorousGreedStealsAndAddsTwoOfOneColor(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	target := pushVanillaCreature(g, opp.ID, "Stolen", 3, 3)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(target) })
	castCatalogSpell(t, g, "Traitorous Greed", "Sorcery", jumpOracle(t, "Traitorous Greed"), jumpTargets(target))
	passPriorityAroundTable(t, g)
	if c := gaLayered(t, g, target); c.Controller != me.ID || c.Tapped || !hasEffectiveKeyword(t, g, target, "haste") {
		t.Fatal("control, untap or haste missing")
	}
	p := manaPickFor(g, me.ID)
	if p == nil {
		t.Fatal("no color choice")
	}
	if err := g.ResolveManaChoice(p.ID, me.ID, "U"); err != nil {
		t.Fatal(err)
	}
	if me.ManaPool.Total() != 2 || manaPickFor(g, me.ID) != nil {
		t.Fatal("must choose one color for both mana")
	}
}

func TestJumpstartCrushingCanopyModes(t *testing.T) {
	for mode := 0; mode < 2; mode++ {
		t.Run(string(rune('0'+mode)), func(t *testing.T) {
			g := newCatalogGame(t)
			opp := g.Seats[1]
			typ := "Enchantment"
			oracle := ""
			if mode == 0 {
				typ = "Creature"
				oracle = jumpOracle(t, "Hellkite Punisher")
			}
			target := b12Push(g, opp.ID, "Target", typ, oracle, 2, 3)
			castModal(t, g, "Crushing Canopy", "Instant", jumpOracle(t, "Crushing Canopy"), []int{mode}, jumpTargets(target))
			passPriorityAroundTable(t, g)
			if battlefieldHas(g, target) {
				t.Fatal("selected mode did not destroy its target")
			}
		})
	}
}

func TestJumpstartChainLightningPaysAndCopiesForHitPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	_ = g.AddManaForEffect(opp.ID, uuid.Nil, "{R}{R}")
	a, b := me.Life, opp.Life
	castCatalogSpell(t, g, "Chain Lightning", "Sorcery", jumpOracle(t, "Chain Lightning"), []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	if opp.Life != b-3 || !hasPayUnlessFor(g, opp.ID) {
		t.Fatal("damage must precede hit player's payment")
	}
	g = restoreRoundTrip(t, g, true)
	me, opp = g.Seats[0], g.Seats[1]
	answerPayUnless(t, g, opp.ID, true)
	answerMayChoice(t, g, opp.ID, true)
	pick := latestPickTarget(g, opp.ID)
	if pick == nil {
		t.Fatal("copy has no retarget choice")
	}
	if err := g.ResolvePickTarget(pick.ID, opp.ID, game.TargetRef{Kind: game.TargetPlayer, ID: me.ID}); err != nil {
		t.Fatal(err)
	}
	copies := copiesOnStack(g)
	if len(copies) != 1 || copies[0].Controller != opp.ID {
		t.Fatal("copy belongs to the paying player")
	}
	passPriorityAroundTable(t, g)
	answerPayUnless(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if me.Life != a-3 || opp.ManaPool.Total() != 0 {
		t.Fatal("copy damage/payment")
	}
}

func TestJumpstartSavageStompDiscountRequiresYourDinosaur(t *testing.T) {
	for _, yours := range []bool{false, true} {
		t.Run(map[bool]string{false: "opposing-dinosaur", true: "your-dinosaur"}[yours], func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			advanceToMain(t, g)
			mineType, theirType := "Creature — Human", "Creature — Dinosaur"
			if yours {
				mineType, theirType = theirType, mineType
			}
			mine := b12Creature(g, me.ID, "Mine", mineType, 2, 3)
			theirs := b12Creature(g, opp.ID, "Theirs", theirType, 2, 3)
			id := pushCatalogHandCard(me, "Savage Stomp", "Sorcery", jumpOracle(t, "Savage Stomp"))
			for i := range me.Hand.Cards {
				if me.Hand.Cards[i].InstanceID == id {
					me.Hand.Cards[i].ManaCost = "{2}{G}"
				}
			}
			_ = g.AddManaForEffect(me.ID, uuid.Nil, "{G}")
			err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: jumpTargets(mine, theirs), Strict: true})
			if (err == nil) != yours {
				t.Fatalf("one-green cast with your Dinosaur=%v: %v", yours, err)
			}
		})
	}
}

func TestJumpstartGarruksHarbingerLooksChoosesAndRestores(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := jumpCard(t, g, me.ID, "Garruk's Harbinger", "Creature — Beast", 4, 3)
	pw := b12Permanent(g, opp.ID, "Walker", "Planeswalker — Test")
	jumpCounter(t, g, pw, game.CounterLoyalty, 5)
	ids := seedSearchLibrary(me, game.Card{Name: "Creature", TypeLine: "Creature — Elf", Power: 2, Toughness: 2}, game.Card{Name: "Garruk", TypeLine: "Legendary Planeswalker — Garruk"}, game.Card{Name: "Spell", TypeLine: "Instant"})
	if !hasEffectiveKeyword(t, g, src, game.HexproofFromColor("B")) {
		t.Fatal("hexproof from black missing")
	}
	jumpEvent(g, game.Event{Kind: game.EventDealDamage, Source: src, Target: pw, Amount: 3, Combat: true})
	passPriorityAroundTable(t, g)
	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	pick := latestChooseCardsFor(g, me.ID)
	if pick == nil || !hasID(pick.ChooseCards, ids[0]) || !hasID(pick.ChooseCards, ids[1]) || hasID(pick.ChooseCards, ids[2]) {
		t.Fatal("wrong creature/Garruk candidates")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{ids[1]}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(ids[1]) || me.Library.Size() != 2 {
		t.Fatal("chosen card/rest handling")
	}
	c := gaLayered(t, g, ids[1])
	for _, p := range g.Seats {
		if !c.KnownBy[p.ID] {
			t.Fatal("chosen card not revealed")
		}
	}
}

func TestJumpstartGarrukUnleashedLoyaltyAbilities(t *testing.T) {
	for _, targeted := range []bool{false, true} {
		t.Run(map[bool]string{false: "plus-one-none", true: "plus-one-target"}[targeted], func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			advanceToMain(t, g)
			src := jumpCard(t, g, me.ID, "Garruk, Unleashed", "Legendary Planeswalker — Garruk", 0, 0)
			jumpCounter(t, g, src, game.CounterLoyalty, 4)
			creature := pushVanillaCreature(g, me.ID, "Creature", 2, 2)
			p := game.ActivateAbilityParams{Strict: true}
			if targeted {
				p.Targets = jumpTargets(creature)
			}
			if err := g.ActivateCatalogAbility(me.ID, src, 0, p); err != nil {
				t.Fatal(err)
			}
			passPriorityAroundTable(t, g)
			if gaLayered(t, g, src).Counters[game.CounterLoyalty] != 5 {
				t.Fatal("+1 cost")
			}
			want := 2
			if targeted {
				want = 5
			}
			if gaLayered(t, g, creature).CurrentPower() != want || hasEffectiveKeyword(t, g, creature, "trample") != targeted {
				t.Fatal("optional pump/trample")
			}
		})
	}
	for _, opponents := range []int{1, 3} {
		t.Run(map[int]string{1: "beast-equal", 3: "beast-outnumbered"}[opponents], func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			advanceToMain(t, g)
			src := jumpCard(t, g, me.ID, "Garruk, Unleashed", "Legendary Planeswalker — Garruk", 0, 0)
			jumpCounter(t, g, src, game.CounterLoyalty, 4)
			for i := 0; i < opponents; i++ {
				pushVanillaCreature(g, opp.ID, "Other", 1, 1)
			}
			if err := g.ActivateCatalogAbility(me.ID, src, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
				t.Fatal(err)
			}
			passPriorityAroundTable(t, g)
			want := 2
			if opponents == 3 {
				want = 3
			}
			if countBattlefieldNamed(g, me.ID, "Beast") != 1 || gaLayered(t, g, src).Counters[game.CounterLoyalty] != want {
				t.Fatal("must count creatures after creating the Beast")
			}
		})
	}
	t.Run("emblem", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[0]
		advanceToMain(t, g)
		src := jumpCard(t, g, me.ID, "Garruk, Unleashed", "Legendary Planeswalker — Garruk", 0, 0)
		jumpCounter(t, g, src, game.CounterLoyalty, 7)
		ids := seedSearchLibrary(me, game.Card{Name: "Found", TypeLine: "Creature — Elf", Power: 2, Toughness: 2})
		if err := g.ActivateCatalogAbility(me.ID, src, 2, game.ActivateAbilityParams{Strict: true}); err != nil {
			t.Fatal(err)
		}
		passPriorityAroundTable(t, g)
		jumpEvent(g, game.Event{Kind: game.EventBeginEndStep, Actor: me.ID})
		passPriorityAroundTable(t, g)
		answerSearchByID(t, g, me.ID, ids[0])
		passPriorityAroundTable(t, g)
		if !battlefieldHas(g, ids[0]) {
			t.Fatal("emblem did not search a creature onto the battlefield")
		}
	})
}

func TestJumpstartSarkhansUnsealingPowerBands(t *testing.T) {
	for _, power := range []int{3, 4, 6, 7} {
		t.Run(string(rune('0'+power)), func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			advanceToMain(t, g)
			jumpCard(t, g, me.ID, "Sarkhan's Unsealing", "Enchantment", 0, 0)
			victim := pushVanillaCreature(g, opp.ID, "Victim", 1, 4)
			life := opp.Life
			id := pushCatalogHandCard(me, "Creature spell", "Creature — Beast", jumpOracle(t, "Hellkite Punisher"))
			for i := range me.Hand.Cards {
				if me.Hand.Cards[i].InstanceID == id {
					me.Hand.Cards[i].Power = power
					me.Hand.Cards[i].Toughness = 7
				}
			}
			if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
				t.Fatal(err)
			}
			if power == 4 || power == 6 {
				p := latestPickTarget(g, me.ID)
				if p == nil {
					t.Fatal("mid-band has no target")
				}
				if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
					t.Fatal(err)
				}
			}
			passPriorityAroundTable(t, g)
			damage := 0
			if power >= 4 {
				damage = 4
			}
			if opp.Life != life-damage {
				t.Fatal("wrong power-band damage")
			}
			if battlefieldHas(g, victim) != (power < 7) {
				t.Fatal("only the seven-power band should sweep the opponent's creature")
			}
		})
	}
}

func TestJumpstartSarkhansUnsealingDamageUsesOriginalSourceObject(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	src := jumpCard(t, g, me.ID, "Sarkhan's Unsealing", "Enchantment", 0, 0)
	g.WithWriteLock(func() {
		if err := (GrantKeywordUntilEOT{Target: src, Keywords: []string{"lifelink"}, Label: "test lifelink"}).Apply(NewContext(g, &game.StackItem{Controller: me.ID})); err != nil {
			t.Fatal(err)
		}
	})
	id := pushCatalogHandCard(me, "Four-power creature", "Creature — Beast", jumpOracle(t, "Hellkite Punisher"))
	for i := range me.Hand.Cards {
		if me.Hand.Cards[i].InstanceID == id {
			me.Hand.Cards[i].Power = 4
			me.Hand.Cards[i].Toughness = 4
		}
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("missing damage target")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatal(err)
	}
	g.WithWriteLock(func() {
		_ = g.SacrificePermanentForEffect(src)
		_ = g.ReturnFromGraveyardUnderControlForEffect(src, game.ZoneBattlefield, opp.ID)
	})
	life := me.Life
	passPriorityAroundTable(t, g)
	if me.Life != life+4 {
		t.Fatal("the old source's lifelink was replaced by the returned enchantment's traits")
	}
}
