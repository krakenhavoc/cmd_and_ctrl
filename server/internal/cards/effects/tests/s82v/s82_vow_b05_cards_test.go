package testss82v

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_vow_b05_cards_test.go — S82 VOW batch 5 (#5495): Blood and Vampires,
// counted amounts. Every test plays the real catalog card through the engine.

const (
	vowB05LacerateFlesh = "201c8de7-9675-475b-8f9e-7b1935234488"
	vowB05Odric         = "690f79f3-e6a5-4542-8401-61a9bf645d55"
	vowB05Attendants    = "6b384c3c-0d6b-4bba-b4c7-5af8b32bdf1c"
	vowB05Vengeance     = "3a82e724-28f6-430f-a474-ab4276197963"
	vowB05Enforcer      = "e262ea6f-c8c7-45c7-a9a6-29dcbe6300a4"
	vowB05Alchemy       = "dfc982a2-9ae8-4206-a2d7-4dcfcde02451"
	vowB05Bloodcaster   = "a720f3d3-c2d7-4e50-8cc6-600120380e9c"
	vowB05Dominating    = "145ef633-a2e0-4bd9-a866-5767821eb44f"
	vowB05Arsonist      = "d9ee7916-2c9d-4489-8dd8-86929b09c6ea"
)

func TestVowB05EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		vowB05LacerateFlesh: "Lacerate Flesh", vowB05Odric: "Odric, Blood-Cursed",
		vowB05Attendants: "Olivia's Attendants", vowB05Vengeance: "Vampires' Vengeance",
		vowB05Enforcer: "Markov Enforcer", vowB05Alchemy: "Arterial Alchemy",
		vowB05Bloodcaster: "Voldaren Bloodcaster", vowB05Bloodcaster + "#1": "Bloodbat Summoner",
		vowB05Dominating: "Dominating Vampire", vowB05Arsonist: "Midnight Arsonist",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// vowB05Cast puts a creature card in the active seat's hand and casts it
// (no mana cost on the fixture, so it is free), settling its entry triggers.
//
// Each pick answers one pick_target prompt, in order, for the entry trigger.
func vowB05Cast(t *testing.T, g *game.Game, name, typeLine, oracle string, power, toughness int, targets []game.TargetRef, picks ...uuid.UUID) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: active.ID, Controller: active.ID,
	})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{Targets: targets}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, pick := range picks {
		effectstest.PickCard(t, g, active.ID, pick)
		effectstest.PassPriorityAroundTable(t, g)
	}
	return id
}

func vowB05Blood(g *game.Game, controller uuid.UUID) int {
	return effectstest.B27CountNamed(g, controller, "Blood Token") + effectstest.B27CountNamed(g, controller, "Blood")
}

func vowB05Vampire(g *game.Game, owner uuid.UUID, name string, p, tough int) uuid.UUID {
	return effectstest.B12Creature(g, owner, name, "Creature — Vampire", p, tough)
}

func TestVowB05LacerateFleshMakesBloodForExcessDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Lacerate Flesh", "Sorcery", vowB05LacerateFlesh, effectstest.B16TargetCard(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Fatal("4 damage kills a 2/2")
	}
	if got := vowB05Blood(g, me.ID); got != 2 {
		t.Errorf("4 damage to a 2/2: %d Blood tokens, want 2", got)
	}
	// Damage already marked lowers what was lethal: 4 into a 5/5 with 3 marked is 2 excess.
	ox := effectstest.B12Creature(g, opp.ID, "Ox", "Creature — Ox", 5, 5)
	effectstest.B27Damage(g, me.ID, ox, 3)
	effectstest.CastCatalogSpell(t, g, "Lacerate Flesh", "Sorcery", vowB05LacerateFlesh, effectstest.B16TargetCard(ox))
	effectstest.PassPriorityAroundTable(t, g)
	if got := vowB05Blood(g, me.ID); got != 4 {
		t.Errorf("4 damage to a 5/5 with 3 marked: %d Blood tokens in all, want 4", got)
	}
	// No excess, no Blood.
	wall := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 6)
	effectstest.CastCatalogSpell(t, g, "Lacerate Flesh", "Sorcery", vowB05LacerateFlesh, effectstest.B16TargetCard(wall))
	effectstest.PassPriorityAroundTable(t, g)
	if got := vowB05Blood(g, me.ID); got != 4 {
		t.Errorf("4 damage to a 0/6: %d Blood tokens, want still 4", got)
	}
	// A player is not a legal target.
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Lacerate Flesh", "Sorcery", vowB05LacerateFlesh,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}); err == nil {
		t.Error("Lacerate Flesh targeted a player")
	}
}

func TestVowB05VampiresVengeanceSparesVampiresAndMakesBlood(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "Elf", "Creature — Elf", 2, 2)
	big := effectstest.B12Creature(g, opp.ID, "Giant", "Creature — Giant", 4, 4)
	vamp := vowB05Vampire(g, opp.ID, "Vampire", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Vampires' Vengeance", "Instant", vowB05Vengeance, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || g.Battlefield.Contains(mine) {
		t.Error("2 damage kills each non-Vampire 2/2, yours included")
	}
	if !g.Battlefield.Contains(vamp) || !g.Battlefield.Contains(big) {
		t.Error("the Vampire and the 4/4 survive")
	}
	if got := vowB05Blood(g, me.ID); got != 1 {
		t.Errorf("%d Blood tokens, want 1", got)
	}
	// The Blood token comes even when nothing is hit.
	g2 := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g2)
	effectstest.CastCatalogSpell(t, g2, "Vampires' Vengeance", "Instant", vowB05Vengeance, nil)
	effectstest.PassPriorityAroundTable(t, g2)
	if got := vowB05Blood(g2, g2.Seats[0].ID); got != 1 {
		t.Errorf("empty board: %d Blood tokens, want 1", got)
	}
}

func TestVowB05OdricCountsEachListedAbilityOnce(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	a := effectstest.B12Creature(g, me.ID, "Flyer", "Creature — Bird", 1, 1)
	b := effectstest.B12Creature(g, me.ID, "Other Flyer", "Creature — Bird", 1, 1)
	c := effectstest.B12Creature(g, me.ID, "Strider", "Creature — Elf", 1, 1)
	d := effectstest.B12Creature(g, g.Seats[1].ID, "Their Reach", "Creature — Spider", 1, 1)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			switch g.Battlefield.Cards[i].InstanceID {
			case a:
				g.Battlefield.Cards[i].Keywords = []string{"flying", "vigilance", "ward"}
			case b:
				g.Battlefield.Cards[i].Keywords = []string{"flying"}
			case c:
				g.Battlefield.Cards[i].Keywords = []string{"trample", "first strike"}
			case d:
				g.Battlefield.Cards[i].Keywords = []string{"reach"}
			}
		}
	})
	vowB05Cast(t, g, "Odric, Blood-Cursed", "Legendary Creature — Vampire Soldier", vowB05Odric, 3, 3, nil)
	// flying, vigilance, trample, first strike: ward is not listed, flying counts once, theirs don't count.
	if got := vowB05Blood(g, me.ID); got != 4 {
		t.Errorf("%d Blood tokens, want 4", got)
	}
}

func TestVowB05OdricWithNoAbilitiesMakesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	vowB05Cast(t, g, "Odric, Blood-Cursed", "Legendary Creature — Vampire Soldier", vowB05Odric, 3, 3, nil)
	if got := vowB05Blood(g, me.ID); got != 0 {
		t.Errorf("%d Blood tokens, want 0", got)
	}
}

func TestVowB05OliviasAttendantsPingMakesABloodToken(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	att := effectstest.PushCatalogPermanent(g, me.ID, "Olivia's Attendants", "Creature — Vampire", vowB05Attendants, false)
	life := opp.Life
	effectstest.ApaMana(me, "C", "C", "R")
	effectstest.ApaActivate(t, g, me, att, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("opponent life %d -> %d, want -1", life, opp.Life)
	}
	if got := vowB05Blood(g, me.ID); got != 1 {
		t.Errorf("%d Blood tokens, want 1 (1 damage dealt)", got)
	}
}

func vowB05SetManaCost(g *game.Game, id uuid.UUID, cost string, tapped bool) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].ManaCost = cost
				g.Battlefield.Cards[i].Tapped = tapped
			}
		}
	})
}

func TestVowB05DominatingVampireStealsWithinTheVampireCount(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	vowB05Vampire(g, me.ID, "Friend", 1, 1)
	cheap := effectstest.B12Creature(g, opp.ID, "Cheap", "Creature — Bear", 2, 2)
	vowB05SetManaCost(g, cheap, "{1}{G}", true)
	pricey := effectstest.B12Creature(g, opp.ID, "Pricey", "Creature — Bear", 2, 2)
	vowB05SetManaCost(g, pricey, "{2}{G}{G}", false)
	// Two Vampires once it enters: mana value 2 is in reach, 4 is not.
	vowB05Cast(t, g, "Dominating Vampire", "Creature — Vampire", vowB05Dominating, 3, 3, nil, cheap)
	if c, ok := effectstest.BattlefieldCard(g, cheap); !ok || c.Controller != me.ID {
		t.Fatal("the mana value 2 creature should be under my control")
	}
	if c, _ := effectstest.BattlefieldCard(g, cheap); c.Tapped {
		t.Error("the stolen creature is untapped")
	}
	if !game.HasKeyword(effectstest.ApaLive(g, cheap), "haste") {
		t.Error("the stolen creature has haste")
	}
	if c, _ := effectstest.BattlefieldCard(g, pricey); c.Controller != opp.ID {
		t.Error("the mana value 4 creature is out of reach")
	}
}

func TestVowB05DominatingVampireWithNothingInReachDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	pricey := effectstest.B12Creature(g, opp.ID, "Pricey", "Creature — Bear", 2, 2)
	vowB05SetManaCost(g, pricey, "{2}{G}{G}", false)
	vowB05Cast(t, g, "Dominating Vampire", "Creature — Vampire", vowB05Dominating, 3, 3, nil)
	if c, _ := effectstest.BattlefieldCard(g, pricey); c.Controller != opp.ID {
		t.Error("one Vampire reaches mana value 1 only; nothing was stolen")
	}
	_ = me
}

func TestVowB05MarkovEnforcerFightsOnEntryAndOnOtherVampires(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	ox := effectstest.B12Creature(g, opp.ID, "Ox", "Creature — Ox", 3, 6)
	enforcer := vowB05Cast(t, g, "Markov Enforcer", "Creature — Vampire Soldier", vowB05Enforcer, 6, 6, nil, bear)
	if g.Battlefield.Contains(bear) && g.Battlefield.Contains(ox) {
		t.Fatal("the Enforcer fights a creature as it enters")
	}
	if g.Battlefield.Contains(bear) {
		t.Fatal("the Bear fought a 6/6 and should be dead")
	}
	if got := vowB05Blood(g, me.ID); got != 1 {
		t.Errorf("a creature dealt damage by the Enforcer died: %d Blood tokens, want 1", got)
	}
	// A non-Vampire entering does not fight.
	before := vowB05Blood(g, me.ID)
	vowB05Cast(t, g, "Elf", "Creature — Elf", "", 1, 1, nil)
	if vowB05Blood(g, me.ID) != before {
		t.Error("a non-Vampire entering triggered the fight")
	}
	// Another Vampire entering fights again: the Ox (3/6) is dealt 6 and dies.
	vowB05Cast(t, g, "Vampire", "Creature — Vampire", "", 1, 1, nil, ox)
	if g.Battlefield.Contains(ox) {
		t.Error("another Vampire entering should have made the Enforcer fight the Ox")
	}
	if got := vowB05Blood(g, me.ID); got != before+1 {
		t.Errorf("the Ox died after the Enforcer fought it: %d Blood tokens, want %d", got, before+1)
	}
	if !g.Battlefield.Contains(enforcer) {
		t.Error("the Enforcer survives")
	}
}

func vowB05PushToken(g *game.Game, owner uuid.UUID, tok game.Card) uuid.UUID {
	tok.InstanceID = uuid.New()
	tok.Owner, tok.Controller = owner, owner
	return effectstest.PushBattlefieldCardWithTimestamp(g, tok)
}

func TestVowB05MidnightArsonistDestroysUpToXArtifactsWithoutManaAbilities(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	treasure := vowB05PushToken(g, opp.ID, effects.TreasureToken())
	blood := vowB05PushToken(g, opp.ID, effects.BloodToken())
	trinket := effectstest.B12Permanent(g, opp.ID, "Trinket", "Artifact")
	vowB05Vampire(g, me.ID, "Friend", 1, 1)
	// Two Vampires: up to two targets. The Treasure has a mana ability and is no target.
	vowB05Cast(t, g, "Midnight Arsonist", "Creature — Vampire", vowB05Arsonist, 3, 2, nil, blood)
	if g.Battlefield.Contains(blood) {
		t.Error("the Blood token should be destroyed")
	}
	if !g.Battlefield.Contains(treasure) {
		t.Error("a Treasure has a mana ability and is not a legal target")
	}
	_ = trinket
}

func TestVowB05MidnightArsonistCannotTargetAManaArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	treasure := vowB05PushToken(g, opp.ID, effects.TreasureToken())
	vowB05Cast(t, g, "Midnight Arsonist", "Creature — Vampire", vowB05Arsonist, 3, 2, nil)
	if !g.Battlefield.Contains(treasure) {
		t.Error("the only artifact has a mana ability; nothing may be destroyed")
	}
	if p := effectstest.LatestPickTarget(g, g.Seats[0].ID); p != nil {
		t.Error("with no legal artifact the trigger asks for nothing")
	}
}

func TestVowB05ArterialAlchemyMakesBloodPerOpponentAndEquipsThem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Arterial Alchemy", "Enchantment", vowB05Alchemy, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := vowB05Blood(g, me.ID); got != 3 {
		t.Fatalf("%d Blood tokens, want 3 (one per opponent)", got)
	}
	var blood uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.HasSubtype("Blood") {
			blood = c.InstanceID
		}
	}
	live := effectstest.ApaLive(g, blood)
	if !live.HasSubtype("Equipment") || !live.IsArtifact() {
		t.Fatalf("the Blood token should be an Equipment artifact: %+v", live.Effective().Subtypes)
	}
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	// Equip {2}: the ability is the token's own, after its Blood ability.
	effectstest.ApaMana(me, "C", "C")
	idx := -1
	for i := 0; i < 4; i++ {
		if err := g.ActivateCatalogAbility(me.ID, blood, i, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(bear)}); err == nil {
			idx = i
			break
		}
	}
	if idx < 0 {
		t.Fatal("no ability of the Blood token equips it")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, bear); p != 4 {
		t.Errorf("equipped Bear power %d, want 4", p)
	}
}

func vowB05BloodcasterCard(owner uuid.UUID) game.Card {
	c := game.Card{
		InstanceID: uuid.New(), OracleID: vowB05Bloodcaster, Layout: game.LayoutTransform,
		Owner: owner, Controller: owner,
		Faces: []game.Face{
			{Name: "Voldaren Bloodcaster", TypeLine: "Creature — Vampire Wizard", ManaCost: "{1}{B}", Colors: []string{"B"}, Power: 2, Toughness: 1},
			{Name: "Bloodbat Summoner", TypeLine: "Creature — Vampire Wizard", Colors: []string{"B"}, Power: 3, Toughness: 3},
		},
	}
	c.SetFace(0)
	return c
}

func vowB05PushBloodcaster(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, vowB05BloodcasterCard(owner))
}

func TestVowB05BloodcasterTransformsOnTheFifthBloodToken(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	caster := vowB05PushBloodcaster(g, me.ID)
	for i := 0; i < 3; i++ {
		vowB05PushToken(g, me.ID, effects.BloodToken())
	}
	effectstest.CastCatalogSpell(t, g, "Vampires' Vengeance", "Instant", vowB05Vengeance, nil) // Blood #4
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, caster); c.ActiveFace != 0 {
		t.Fatal("four Blood tokens do not transform it")
	}
	if got := vowB05Blood(g, me.ID); got != 4 {
		t.Fatalf("%d Blood tokens, want 4", got)
	}
	effectstest.CastCatalogSpell(t, g, "Vampires' Vengeance", "Instant", vowB05Vengeance, nil) // Blood #5
	effectstest.PassPriorityAroundTable(t, g)
	if c, ok := effectstest.BattlefieldCard(g, caster); !ok || c.ActiveFace != 1 {
		t.Fatal("the fifth Blood token should transform it into Bloodbat Summoner")
	}
}

func TestVowB05BloodcasterMakesBloodWhenItOrANontokenCreatureDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	vowB05PushBloodcaster(g, me.ID)
	nontoken := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	servo := effects.TokenCard("1/1 colorless Servo artifact")
	token := vowB05PushToken(g, me.ID, servo)
	effectstest.B27Damage(g, me.ID, token, 5)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if got := vowB05Blood(g, me.ID); got != 0 {
		t.Errorf("a token dying made %d Blood tokens, want 0", got)
	}
	effectstest.B27Damage(g, me.ID, nontoken, 5)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if got := vowB05Blood(g, me.ID); got != 1 {
		t.Errorf("a nontoken creature dying made %d Blood tokens, want 1", got)
	}
}

func TestVowB05BloodbatSummonerTurnsABloodTokenIntoABat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	card := vowB05BloodcasterCard(me.ID)
	card.SetFace(1)
	summoner := effectstest.PushBattlefieldCardWithTimestamp(g, card)
	_ = summoner
	blood := vowB05PushToken(g, me.ID, effects.BloodToken())
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PickCard(t, g, me.ID, blood)
	effectstest.PassPriorityAroundTable(t, g)
	live := effectstest.ApaLive(g, blood)
	eff := live.Effective()
	if !live.IsCreature() || !live.HasSubtype("Bat") || !live.HasSubtype("Blood") || !live.IsArtifact() {
		t.Fatalf("the Blood token should be a Bat creature in addition to its other types: %+v", eff)
	}
	if p, tt := effectstest.EffectivePower(t, g, blood), effectstest.EffectiveToughness(t, g, blood); p != 2 || tt != 2 {
		t.Errorf("Bat is %d/%d, want 2/2", p, tt)
	}
	if !game.HasKeyword(live, "flying") || !game.HasKeyword(live, "haste") {
		t.Error("the Bat has flying and haste")
	}
}
