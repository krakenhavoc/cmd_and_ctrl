package testsi

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #4380: real catalog rows and choices, including resolution ordering,
// ownership, and restored keyed prompts. Neyith is held for #4384.
func jumpOracle(t *testing.T, name string) string {
	t.Helper()
	for _, s := range effects.All() {
		if s.Name == name {
			return s.OracleID
		}
	}
	t.Fatalf("missing %s", name)
	return ""
}
func jumpCard(t *testing.T, g *game.Game, who uuid.UUID, name, typ string, power, toughness int) uuid.UUID {
	t.Helper()
	return effectstest.B12Push(g, who, name, typ, jumpOracle(t, name), power, toughness)
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
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			id := effectstest.PushVanillaCreature(g, me.ID, "Recipient", 2, 2)
			if tc.name == "Primal Might" {
				effectstest.CastXSpell(t, g, tc.name, "Sorcery", jumpOracle(t, tc.name), "", 3, effectstest.JumpTargets(id))
			} else {
				effectstest.CastCatalogSpell(t, g, tc.name, "Instant", jumpOracle(t, tc.name), effectstest.JumpTargets(id))
			}
			effectstest.PassPriorityAroundTable(t, g)
			c := effectstest.GaLayered(t, g, id)
			if c.CurrentPower() != tc.power || c.CurrentToughness() != tc.toughness || c.Counters[game.CounterPlusOne] != tc.counters {
				t.Fatalf("pump/counters: %+v", c)
			}
			if tc.keyword != "" && !effectstest.HasEffectiveKeyword(t, g, id, tc.keyword) {
				t.Fatalf("missing %s", tc.keyword)
			}
		})
	}
}

func TestJumpstartCounterThenDamageAndFight(t *testing.T) {
	for _, name := range []string{"Hunter's Edge", "Savage Stomp", "Time to Feed"} {
		t.Run(name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			mine := effectstest.PushVanillaCreature(g, me.ID, "Fighter", 3, 5)
			theirs := effectstest.PushVanillaCreature(g, opp.ID, "Other", 2, 4)
			targets := effectstest.JumpTargets(mine, theirs)
			if name == "Time to Feed" {
				targets = effectstest.JumpTargets(theirs, mine)
			}
			life := me.Life
			effectstest.CastCatalogSpell(t, g, name, "Sorcery", jumpOracle(t, name), targets)
			effectstest.PassPriorityAroundTable(t, g)
			if name == "Time to Feed" {
				if !effectstest.BattlefieldHas(g, theirs) {
					t.Fatal("3 power should not kill a 4 toughness target")
				}
				g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(theirs) })
				effectstest.PassPriorityAroundTable(t, g)
				if me.Life != life+3 {
					t.Fatal("later death did not gain three life")
				}
			} else {
				c := effectstest.GaLayered(t, g, mine)
				if c.Counters[game.CounterPlusOne] != 1 || effectstest.BattlefieldHas(g, theirs) {
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.PushVanillaCreature(g, me.ID, "Mine", 2, 2)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Theirs", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Savage Stomp", "Sorcery", jumpOracle(t, "Savage Stomp"), effectstest.JumpTargets(mine, theirs))
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(theirs) })
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.GaLayered(t, g, mine).Counters[game.CounterPlusOne] != 1 || effectstest.GaLayered(t, g, mine).DamageMarked != 0 {
		t.Fatal("legal first clause did not get its counter independently")
	}
}

func TestJumpstartFungalRebirthMorbid(t *testing.T) {
	for _, dead := range []bool{false, true} {
		t.Run(map[bool]string{false: "no death", true: "death"}[dead], func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			target := effectstest.GyCardWithOracle(me, "Forest", "Basic Land — Forest", "")
			if dead {
				id := effectstest.PushVanillaCreature(g, g.Seats[1].ID, "Doomed", 1, 1)
				g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(id) })
			}
			effectstest.CastCatalogSpell(t, g, "Fungal Rebirth", "Instant", jumpOracle(t, "Fungal Rebirth"), effectstest.JumpTargets(target))
			effectstest.PassPriorityAroundTable(t, g)
			if !me.Hand.Contains(target) {
				t.Fatal("permanent card not returned")
			}
			want := 0
			if dead {
				want = 2
			}
			if effectstest.CountBattlefieldNamed(g, me.ID, "Saproling") != want {
				t.Fatal("morbid token count")
			}
		})
	}
}

func TestJumpstartBallLightningSacrificesAtAnyEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := jumpCard(t, g, me.ID, "Ball Lightning", "Creature — Elemental", 6, 1)
	if !effectstest.HasEffectiveKeyword(t, g, id, "haste") || !effectstest.HasEffectiveKeyword(t, g, id, "trample") {
		t.Fatal("printed keywords missing")
	}
	jumpEvent(g, game.Event{Kind: game.EventBeginEndStep, Actor: g.Seats[1].ID})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.BattlefieldHas(g, id) || !me.Graveyard.Contains(id) {
		t.Fatal("did not sacrifice on opponent's end step")
	}
}

func TestJumpstartBallLightningCannotSacrificeAnOpponentsPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
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
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.BattlefieldHas(g, id) {
		t.Fatal("the old controller sacrificed the new controller's Ball Lightning")
	}
}

func TestJumpstartCombatTriggers(t *testing.T) {
	for _, name := range []string{"Rampaging Brontodon", "Ornery Dilophosaur"} {
		t.Run(name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			id := jumpCard(t, g, me.ID, name, "Creature — Dinosaur", 2, 2)
			if name == "Rampaging Brontodon" {
				for i := 0; i < 3; i++ {
					effectstest.B12Permanent(g, me.ID, "Land", "Land")
				}
			} else {
				effectstest.PushVanillaCreature(g, me.ID, "Big", 4, 4)
			}
			jumpEvent(g, game.Event{Kind: game.EventAttack, Actor: me.ID, CardID: id, Source: id, Target: g.Seats[1].ID})
			effectstest.PassPriorityAroundTable(t, g)
			want := 4
			if name == "Rampaging Brontodon" {
				want = 5
			}
			if effectstest.GaLayered(t, g, id).CurrentPower() != want {
				t.Fatal("attack pump")
			}
		})
	}
}

func TestJumpstartOrneryRechecksItsInterveningCondition(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := jumpCard(t, g, me.ID, "Ornery Dilophosaur", "Creature — Dinosaur", 2, 2)
	big := effectstest.PushVanillaCreature(g, me.ID, "Big", 4, 4)
	jumpEvent(g, game.Event{Kind: game.EventAttack, Actor: me.ID, CardID: id, Source: id, Target: g.Seats[1].ID})
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(big) })
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.GaLayered(t, g, id).CurrentPower() != 2 {
		t.Fatal("lost condition still pumped")
	}
}

func TestJumpstartWoodbornThresholdTracksLandChanges(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := jumpCard(t, g, me.ID, "Woodborn Behemoth", "Creature — Elemental", 4, 4)
	for i := 0; i < 7; i++ {
		effectstest.B12Permanent(g, me.ID, "Land", "Land")
	}
	if effectstest.GaLayered(t, g, id).CurrentPower() != 4 {
		t.Fatal("active with seven lands")
	}
	eighth := effectstest.B12Permanent(g, me.ID, "Eighth", "Land")
	if effectstest.GaLayered(t, g, id).CurrentPower() != 8 || !effectstest.HasEffectiveKeyword(t, g, id, "trample") {
		t.Fatal("eight-land boost missing")
	}
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(eighth) })
	if effectstest.GaLayered(t, g, id).CurrentPower() != 4 || effectstest.HasEffectiveKeyword(t, g, id, "trample") {
		t.Fatal("threshold not removed")
	}
}

func TestJumpstartActivatedCombatCards(t *testing.T) {
	for _, name := range []string{"Hellkite Punisher", "Suspicious Bookcase"} {
		t.Run(name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			effectstest.AdvanceToMain(t, g)
			src := jumpCard(t, g, me.ID, name, "Creature", 2, 3)
			target := effectstest.PushVanillaCreature(g, me.ID, "Attacker", 2, 2)
			p := game.ActivateAbilityParams{Strict: true}
			mana := "{R}"
			if name == "Suspicious Bookcase" {
				p.Targets = effectstest.JumpTargets(target)
				mana = "{C}{C}{C}"
			}
			_ = g.AddManaForEffect(me.ID, uuid.Nil, mana)
			if err := g.ActivateCatalogAbility(me.ID, src, 0, p); err != nil {
				t.Fatal(err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if name == "Hellkite Punisher" {
				if effectstest.GaLayered(t, g, src).CurrentPower() != 3 || !effectstest.HasEffectiveKeyword(t, g, src, "flying") {
					t.Fatal("firebreathing")
				}
			} else {
				if effectstest.GaLayered(t, g, target).Effective().Restrictions&game.CantBeBlocked == 0 || !effectstest.GaLayered(t, g, src).Tapped {
					t.Fatal("Bookcase activation")
				}
			}
		})
	}
}

func TestJumpstartBattleRattleOptionalPump(t *testing.T) {
	for _, yes := range []bool{false, true} {
		t.Run(map[bool]string{false: "decline", true: "accept"}[yes], func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			jumpCard(t, g, me.ID, "Battle-Rattle Shaman", "Creature — Goblin Shaman", 2, 2)
			target := effectstest.PushVanillaCreature(g, g.Seats[1].ID, "Opponent's creature", 2, 2)
			jumpEvent(g, game.Event{Kind: game.EventStepBegan, Actor: me.ID, Step: game.StepBeginCombat})
			effectstest.AnswerTriggerTargets(t, g, me.ID, target)
			effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, yes)
			effectstest.PassPriorityAroundTable(t, g)
			want := 2
			if yes {
				want = 4
			}
			if effectstest.GaLayered(t, g, target).CurrentPower() != want {
				t.Fatal("optional target pump")
			}
		})
	}
}

func TestJumpstartFurorAndNewHorizons(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	creature := effectstest.PushVanillaCreature(g, me.ID, "Creature", 2, 2)
	land := effectstest.B12Permanent(g, me.ID, "Land", "Land")
	effectstest.CastCatalogSpell(t, g, "Furor of the Bitten", "Enchantment — Aura", jumpOracle(t, "Furor of the Bitten"), effectstest.JumpTargets(creature))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.GaLayered(t, g, creature).CurrentPower() != 4 || len(effectstest.GaLayered(t, g, creature).Effective().AttackRequirements) == 0 {
		t.Fatal("Furor pump/attack requirement")
	}
	effectstest.CastCatalogSpell(t, g, "New Horizons", "Enchantment — Aura", jumpOracle(t, "New Horizons"), effectstest.JumpTargets(land))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerTriggerTargets(t, g, me.ID, creature)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.GaLayered(t, g, creature).Counters[game.CounterPlusOne] != 1 {
		t.Fatal("Horizons entry counter")
	}
	if err := effectstest.TapGrantedMana(t, g, me.ID, land, "U"); err != nil {
		t.Fatal(err)
	}
	if got := me.ManaPool.Total(); got != 2 {
		t.Fatalf("Horizons mana = %d, want 2", got)
	}
}

func TestJumpstartFertilidTargetSearchRestores(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	src := jumpCard(t, g, me.ID, "Fertilid", "Creature — Elemental", 0, 0)
	jumpCounter(t, g, src, game.CounterPlusOne, 2)
	ids := effectstest.SeedSearchLibrary(opp, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"}, game.Card{Name: "Other", TypeLine: "Creature — Elf"})
	_ = g.AddManaForEffect(me.ID, uuid.Nil, "{G}{C}")
	if err := g.ActivateCatalogAbility(me.ID, src, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}, Strict: true}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	g = effectstest.RestoreRoundTrip(t, g, true)
	effectstest.AnswerSearchByID(t, g, opp.ID, ids[0])
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.GaLayered(t, g, ids[0]); c.Controller != opp.ID || !c.Tapped {
		t.Fatal("target player did not receive tapped basic")
	}
	if effectstest.GaLayered(t, g, src).Counters[game.CounterPlusOne] != 1 {
		t.Fatal("counter cost")
	}
}

func TestJumpstartCharmbreakerReturnsOnlyEligibleCardAndPumps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	src := jumpCard(t, g, me.ID, "Charmbreaker Devils", "Creature — Devil", 4, 4)
	spell := effectstest.GyCardWithOracle(me, "Old spell", "Instant", "")
	other := effectstest.GyCardWithOracle(me, "Creature", "Creature", "")
	jumpEvent(g, game.Event{Kind: game.EventBeginUpkeep, Actor: me.ID})
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(spell) || !me.Graveyard.Contains(other) {
		t.Fatal("random eligible return")
	}
	effectstest.CastCatalogSpell(t, g, "Arbor Armament", "Instant", jumpOracle(t, "Arbor Armament"), effectstest.JumpTargets(src))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.GaLayered(t, g, src).CurrentPower() != 9 {
		t.Fatal("instant cast should pump +4/+0 as well as the spell's counter")
	}
}

func TestJumpstartHamletbackReadsEnteringCreaturesLastKnownPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := jumpCard(t, g, me.ID, "Hamletback Goliath", "Creature — Giant Warrior", 6, 6)
	other := effectstest.PushVanillaCreature(g, opp.ID, "Entering", 4, 4)
	jumpEvent(g, game.Event{Kind: game.EventETB, Actor: opp.ID, CardID: other, Source: other})
	jumpCounter(t, g, other, game.CounterPlusOne, 2)
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(other) })
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.GaLayered(t, g, src).Counters[game.CounterPlusOne] != 6 {
		t.Fatal("entering creature's power was not read at resolution with LKI")
	}
}

func TestJumpstartHuntersInsightWatchesDamageNotGrantedAbilities(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := effectstest.PushVanillaCreature(g, me.ID, "Source", 3, 3)
	pw := effectstest.B12Permanent(g, opp.ID, "Walker", "Planeswalker — Test")
	jumpCounter(t, g, pw, game.CounterLoyalty, 5)
	creature := effectstest.PushVanillaCreature(g, opp.ID, "Other", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Hunter's Insight", "Instant", jumpOracle(t, "Hunter's Insight"), effectstest.JumpTargets(src))
	effectstest.PassPriorityAroundTable(t, g)
	before := me.Hand.Size()
	for _, ev := range []game.Event{{Kind: game.EventDealDamage, Source: src, Target: opp.ID, Amount: 3, Combat: true}, {Kind: game.EventDealDamage, Source: src, Target: pw, Amount: 2, Combat: true}, {Kind: game.EventDealDamage, Source: src, Target: creature, Amount: 8, Combat: true}, {Kind: game.EventDealDamage, Source: src, Target: opp.ID, Amount: 8, Combat: false}} {
		jumpEvent(g, ev)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if me.Hand.Size() != before+5 {
		t.Fatal("must draw only for combat damage to player/planeswalker")
	}
	g.WithWriteLock(func() {
		_ = g.SacrificePermanentForEffect(src)
		_ = g.ReturnFromGraveyardUnderControlForEffect(src, game.ZoneBattlefield, me.ID)
	})
	jumpEvent(g, game.Event{Kind: game.EventDealDamage, Source: src, Target: opp.ID, Amount: 3, Combat: true})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != before+5 {
		t.Fatal("delayed trigger followed a new object")
	}
}

func TestJumpstartTraitorousGreedStealsAndAddsTwoOfOneColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	target := effectstest.PushVanillaCreature(g, opp.ID, "Stolen", 3, 3)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(target) })
	effectstest.CastCatalogSpell(t, g, "Traitorous Greed", "Sorcery", jumpOracle(t, "Traitorous Greed"), effectstest.JumpTargets(target))
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.GaLayered(t, g, target); c.Controller != me.ID || c.Tapped || !effectstest.HasEffectiveKeyword(t, g, target, "haste") {
		t.Fatal("control, untap or haste missing")
	}
	p := effectstest.ManaPickFor(g, me.ID)
	if p == nil {
		t.Fatal("no color choice")
	}
	if err := g.ResolveManaChoice(p.ID, me.ID, "U"); err != nil {
		t.Fatal(err)
	}
	if me.ManaPool.Total() != 2 || effectstest.ManaPickFor(g, me.ID) != nil {
		t.Fatal("must choose one color for both mana")
	}
}

func TestJumpstartCrushingCanopyModes(t *testing.T) {
	for mode := 0; mode < 2; mode++ {
		t.Run(string(rune('0'+mode)), func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			opp := g.Seats[1]
			typ := "Enchantment"
			oracle := ""
			if mode == 0 {
				typ = "Creature"
				oracle = jumpOracle(t, "Hellkite Punisher")
			}
			target := effectstest.B12Push(g, opp.ID, "Target", typ, oracle, 2, 3)
			effectstest.CastModal(t, g, "Crushing Canopy", "Instant", jumpOracle(t, "Crushing Canopy"), []int{mode}, effectstest.JumpTargets(target))
			effectstest.PassPriorityAroundTable(t, g)
			if effectstest.BattlefieldHas(g, target) {
				t.Fatal("selected mode did not destroy its target")
			}
		})
	}
}

func TestJumpstartChainLightningPaysAndCopiesForHitPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	_ = g.AddManaForEffect(opp.ID, uuid.Nil, "{R}{R}")
	a, b := me.Life, opp.Life
	effectstest.CastCatalogSpell(t, g, "Chain Lightning", "Sorcery", jumpOracle(t, "Chain Lightning"), []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != b-3 || !effectstest.HasPayUnlessFor(g, opp.ID) {
		t.Fatal("damage must precede hit player's payment")
	}
	g = effectstest.RestoreRoundTrip(t, g, true)
	me, opp = g.Seats[0], g.Seats[1]
	effectstest.AnswerPayUnless(t, g, opp.ID, true)
	effectstest.AnswerMayChoice(t, g, opp.ID, true)
	pick := effectstest.LatestPickTarget(g, opp.ID)
	if pick == nil {
		t.Fatal("copy has no retarget choice")
	}
	if err := g.ResolvePickTarget(pick.ID, opp.ID, game.TargetRef{Kind: game.TargetPlayer, ID: me.ID}); err != nil {
		t.Fatal(err)
	}
	copies := effectstest.CopiesOnStack(g)
	if len(copies) != 1 || copies[0].Controller != opp.ID {
		t.Fatal("copy belongs to the paying player")
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != a-3 || opp.ManaPool.Total() != 0 {
		t.Fatal("copy damage/payment")
	}
}

func TestJumpstartSavageStompDiscountRequiresYourDinosaur(t *testing.T) {
	for _, yours := range []bool{false, true} {
		t.Run(map[bool]string{false: "opposing-dinosaur", true: "your-dinosaur"}[yours], func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			effectstest.AdvanceToMain(t, g)
			mineType, theirType := "Creature — Human", "Creature — Dinosaur"
			if yours {
				mineType, theirType = theirType, mineType
			}
			mine := effectstest.B12Creature(g, me.ID, "Mine", mineType, 2, 3)
			theirs := effectstest.B12Creature(g, opp.ID, "Theirs", theirType, 2, 3)
			id := effectstest.PushCatalogHandCard(me, "Savage Stomp", "Sorcery", jumpOracle(t, "Savage Stomp"))
			for i := range me.Hand.Cards {
				if me.Hand.Cards[i].InstanceID == id {
					me.Hand.Cards[i].ManaCost = "{2}{G}"
				}
			}
			_ = g.AddManaForEffect(me.ID, uuid.Nil, "{G}")
			err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: effectstest.JumpTargets(mine, theirs), Strict: true})
			if (err == nil) != yours {
				t.Fatalf("one-green cast with your Dinosaur=%v: %v", yours, err)
			}
		})
	}
}

func TestJumpstartGarruksHarbingerLooksChoosesAndRestores(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := jumpCard(t, g, me.ID, "Garruk's Harbinger", "Creature — Beast", 4, 3)
	pw := effectstest.B12Permanent(g, opp.ID, "Walker", "Planeswalker — Test")
	jumpCounter(t, g, pw, game.CounterLoyalty, 5)
	ids := effectstest.SeedSearchLibrary(me, game.Card{Name: "Creature", TypeLine: "Creature — Elf", Power: 2, Toughness: 2}, game.Card{Name: "Garruk", TypeLine: "Legendary Planeswalker — Garruk"}, game.Card{Name: "Spell", TypeLine: "Instant"})
	if !effectstest.HasEffectiveKeyword(t, g, src, game.HexproofFromColor("B")) {
		t.Fatal("hexproof from black missing")
	}
	jumpEvent(g, game.Event{Kind: game.EventDealDamage, Source: src, Target: pw, Amount: 3, Combat: true})
	effectstest.PassPriorityAroundTable(t, g)
	g = effectstest.RestoreRoundTrip(t, g, true)
	me = g.Seats[0]
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil || !effectstest.HasID(pick.ChooseCards, ids[0]) || !effectstest.HasID(pick.ChooseCards, ids[1]) || effectstest.HasID(pick.ChooseCards, ids[2]) {
		t.Fatal("wrong creature/Garruk candidates")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{ids[1]}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(ids[1]) || me.Library.Size() != 2 {
		t.Fatal("chosen card/rest handling")
	}
	c := effectstest.GaLayered(t, g, ids[1])
	for _, p := range g.Seats {
		if !c.KnownBy[p.ID] {
			t.Fatal("chosen card not revealed")
		}
	}
}

func TestJumpstartGarrukUnleashedLoyaltyAbilities(t *testing.T) {
	for _, targeted := range []bool{false, true} {
		t.Run(map[bool]string{false: "plus-one-none", true: "plus-one-target"}[targeted], func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			effectstest.AdvanceToMain(t, g)
			src := jumpCard(t, g, me.ID, "Garruk, Unleashed", "Legendary Planeswalker — Garruk", 0, 0)
			jumpCounter(t, g, src, game.CounterLoyalty, 4)
			creature := effectstest.PushVanillaCreature(g, me.ID, "Creature", 2, 2)
			p := game.ActivateAbilityParams{Strict: true}
			if targeted {
				p.Targets = effectstest.JumpTargets(creature)
			}
			if err := g.ActivateCatalogAbility(me.ID, src, 0, p); err != nil {
				t.Fatal(err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if effectstest.GaLayered(t, g, src).Counters[game.CounterLoyalty] != 5 {
				t.Fatal("+1 cost")
			}
			want := 2
			if targeted {
				want = 5
			}
			if effectstest.GaLayered(t, g, creature).CurrentPower() != want || effectstest.HasEffectiveKeyword(t, g, creature, "trample") != targeted {
				t.Fatal("optional pump/trample")
			}
		})
	}
	for _, opponents := range []int{1, 3} {
		t.Run(map[int]string{1: "beast-equal", 3: "beast-outnumbered"}[opponents], func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			effectstest.AdvanceToMain(t, g)
			src := jumpCard(t, g, me.ID, "Garruk, Unleashed", "Legendary Planeswalker — Garruk", 0, 0)
			jumpCounter(t, g, src, game.CounterLoyalty, 4)
			for i := 0; i < opponents; i++ {
				effectstest.PushVanillaCreature(g, opp.ID, "Other", 1, 1)
			}
			if err := g.ActivateCatalogAbility(me.ID, src, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
				t.Fatal(err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			want := 2
			if opponents == 3 {
				want = 3
			}
			if effectstest.CountBattlefieldNamed(g, me.ID, "Beast") != 1 || effectstest.GaLayered(t, g, src).Counters[game.CounterLoyalty] != want {
				t.Fatal("must count creatures after creating the Beast")
			}
		})
	}
	t.Run("emblem", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.AdvanceToMain(t, g)
		src := jumpCard(t, g, me.ID, "Garruk, Unleashed", "Legendary Planeswalker — Garruk", 0, 0)
		jumpCounter(t, g, src, game.CounterLoyalty, 7)
		ids := effectstest.SeedSearchLibrary(me, game.Card{Name: "Found", TypeLine: "Creature — Elf", Power: 2, Toughness: 2})
		if err := g.ActivateCatalogAbility(me.ID, src, 2, game.ActivateAbilityParams{Strict: true}); err != nil {
			t.Fatal(err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		jumpEvent(g, game.Event{Kind: game.EventBeginEndStep, Actor: me.ID})
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerSearchByID(t, g, me.ID, ids[0])
		effectstest.PassPriorityAroundTable(t, g)
		if !effectstest.BattlefieldHas(g, ids[0]) {
			t.Fatal("emblem did not search a creature onto the battlefield")
		}
	})
}

func TestJumpstartSarkhansUnsealingPowerBands(t *testing.T) {
	for _, power := range []int{3, 4, 6, 7} {
		t.Run(string(rune('0'+power)), func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			effectstest.AdvanceToMain(t, g)
			jumpCard(t, g, me.ID, "Sarkhan's Unsealing", "Enchantment", 0, 0)
			victim := effectstest.PushVanillaCreature(g, opp.ID, "Victim", 1, 4)
			life := opp.Life
			id := effectstest.PushCatalogHandCard(me, "Creature spell", "Creature — Beast", jumpOracle(t, "Hellkite Punisher"))
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
				p := effectstest.LatestPickTarget(g, me.ID)
				if p == nil {
					t.Fatal("mid-band has no target")
				}
				if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
					t.Fatal(err)
				}
			}
			effectstest.PassPriorityAroundTable(t, g)
			damage := 0
			if power >= 4 {
				damage = 4
			}
			if opp.Life != life-damage {
				t.Fatal("wrong power-band damage")
			}
			if effectstest.BattlefieldHas(g, victim) != (power < 7) {
				t.Fatal("only the seven-power band should sweep the opponent's creature")
			}
		})
	}
}

func TestJumpstartSarkhansUnsealingDamageUsesOriginalSourceObject(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	src := jumpCard(t, g, me.ID, "Sarkhan's Unsealing", "Enchantment", 0, 0)
	g.WithWriteLock(func() {
		if err := (effects.GrantKeywordUntilEOT{Target: src, Keywords: []string{"lifelink"}, Label: "test lifelink"}).Apply(effects.NewContext(g, &game.StackItem{Controller: me.ID})); err != nil {
			t.Fatal(err)
		}
	})
	id := effectstest.PushCatalogHandCard(me, "Four-power creature", "Creature — Beast", jumpOracle(t, "Hellkite Punisher"))
	for i := range me.Hand.Cards {
		if me.Hand.Cards[i].InstanceID == id {
			me.Hand.Cards[i].Power = 4
			me.Hand.Cards[i].Toughness = 4
		}
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	pick := effectstest.LatestPickTarget(g, me.ID)
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
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+4 {
		t.Fatal("the old source's lifelink was replaced by the returned enchantment's traits")
	}
}
