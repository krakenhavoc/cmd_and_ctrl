package testsbr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// c21_b057_cards_test.go — Commander 2021 batch 057 (#5504): Scholarship
// Sponsor, Crafty Cutpurse, Naru Meha, Guardian Archon, Sapling of Colfenor,
// Essence Pulse, Ezzaroot Channeler, Blighted Cataract, Sapseep Forest and
// Phyrexia's Core.

const (
	c21B057Sponsor   = "21e3895b-f6b6-48af-9161-71b422818004"
	c21B057Cutpurse  = "2ef53e18-0baf-4eee-b671-4c19b464cbfb"
	c21B057Naru      = "562c9c68-e5a7-45ab-b098-5b649d165ff5"
	c21B057Archon    = "c7748659-0206-448f-8567-83565cb0b51a"
	c21B057Sapling   = "0c1e2bb8-0629-466c-b435-c4c71cdc755d"
	c21B057Pulse     = "0d4bd5e0-8346-4025-8694-d2e1897ee0c2"
	c21B057Channeler = "3e72ec1f-9941-42e6-b832-58b79b44066b"
	c21B057Cataract  = "4d722521-0396-48f9-88b2-08b6249f970d"
	c21B057Sapseep   = "8d4dcab0-86e5-4ff8-a90f-78a062664e16"
	c21B057Core      = "b6cc062c-eb39-46ee-bd6d-17f1db0ac50d"
)

func TestC21B057CardsAreRegisteredFull(t *testing.T) {
	for id, name := range map[string]string{
		c21B057Sponsor: "Scholarship Sponsor", c21B057Cutpurse: "Crafty Cutpurse",
		c21B057Naru: "Naru Meha, Master Wizard", c21B057Archon: "Guardian Archon",
		c21B057Sapling: "Sapling of Colfenor", c21B057Pulse: "Essence Pulse",
		c21B057Channeler: "Ezzaroot Channeler", c21B057Cataract: "Blighted Cataract",
		c21B057Sapseep: "Sapseep Forest", c21B057Core: "Phyrexia's Core",
	} {
		spec, ok := effects.Lookup(id)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, id)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestC21B057ScholarshipSponsorCatchesUpFewerLandsPlayersTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, o1, o2, o3 := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	forest := func(p *game.Player, n int) {
		for i := 0; i < n; i++ {
			effectstest.PushLandFor(g, p.ID, "Forest", "Basic Land — Forest")
		}
	}
	forest(me, 1)
	forest(o1, 4)
	forest(o2, 4)
	forest(o3, 2)
	mine := []uuid.UUID{
		effectstest.B07LibraryLand(me, "Forest", "Basic Land — Forest"),
		effectstest.B07LibraryLand(me, "Forest", "Basic Land — Forest"),
		effectstest.B07LibraryLand(me, "Forest", "Basic Land — Forest"),
		effectstest.B07LibraryLand(me, "Forest", "Basic Land — Forest"),
	}
	theirs := []uuid.UUID{
		effectstest.B07LibraryLand(o3, "Island", "Basic Land — Island"),
		effectstest.B07LibraryLand(o3, "Island", "Basic Land — Island"),
		effectstest.B07LibraryLand(o3, "Island", "Basic Land — Island"),
	}
	effectstest.B07LibraryLand(o1, "Plains", "Basic Land — Plains")

	effectstest.CastCatalogSpell(t, g, "Scholarship Sponsor", "Creature — Human Advisor", c21B057Sponsor, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B07OpenSearchPrompts(g); n != 2 {
		t.Fatalf("%d search prompts, want 2: only the players with fewer lands than the most search", n)
	}
	if effectstest.SearchChoiceFor(g, o1.ID) != nil || effectstest.SearchChoiceFor(g, o2.ID) != nil {
		t.Fatal("a player level with the leader was asked to search")
	}
	// The difference is 3 for me: three is allowed, a fourth is not offered as a legal answer.
	effectstest.AnswerSearchByID(t, g, me.ID, mine[0], mine[1], mine[2])
	effectstest.AnswerSearchByID(t, g, o3.ID, theirs[0], theirs[1])
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.B07LandsOnBattlefield(g, me.ID); got != 4 {
		t.Errorf("my lands = %d, want 4", got)
	}
	if got := effectstest.B07LandsOnBattlefield(g, o3.ID); got != 4 {
		t.Errorf("o3's lands = %d, want 4", got)
	}
	for _, id := range append(mine[:3:3], theirs[:2]...) {
		if c, ok := effectstest.BattlefieldCard(g, id); !ok || !c.Tapped {
			t.Errorf("fetched land %s: on battlefield %v, tapped %v; want tapped on the battlefield", id, ok, c.Tapped)
		}
	}
	if !me.Library.Contains(mine[3]) || !o3.Library.Contains(theirs[2]) {
		t.Error("a land beyond the difference left a library")
	}
}

func TestC21B057ScholarshipSponsorAsksNobodyWhenLandsAreLevel(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.CastCatalogSpell(t, g, "Scholarship Sponsor", "Creature — Human Advisor", c21B057Sponsor, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B07OpenSearchPrompts(g); n != 0 {
		t.Errorf("%d search prompts with every player on zero lands", n)
	}
}

func TestC21B057CraftyCutpurseTakesOpponentsTokensThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tmpl := game.Card{Name: "Spirit", TypeLine: "Token Creature — Spirit", Power: 1, Toughness: 1}
	count := func(p *game.Player) int {
		n := 0
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Spirit" && c.Controller == p.ID && c.IsToken() {
				n++
			}
		}
		return n
	}
	id := effectstest.CastCatalogSpell(t, g, "Crafty Cutpurse", "Creature — Human Pirate", c21B057Cutpurse, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("Crafty Cutpurse did not resolve")
	}
	effectstest.AssertKeywords(t, g, id, "flash")
	g.WithWriteLock(func() {
		if err := g.CreateTokenForEffect(opp.ID, tmpl, 2); err != nil {
			t.Fatal(err)
		}
		if err := g.CreateTokenForEffect(me.ID, tmpl, 1); err != nil {
			t.Fatal(err)
		}
	})
	if got := count(me); got != 3 {
		t.Errorf("my tokens = %d, want 3 (mine and the opponent's two)", got)
	}
	if got := count(opp); got != 0 {
		t.Errorf("the opponent kept %d tokens", got)
	}
}

func TestC21B057CraftyCutpurseDoesNothingBeforeItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	g.WithWriteLock(func() {
		if err := g.CreateTokenForEffect(opp.ID, game.Card{Name: "Spirit", TypeLine: "Token Creature — Spirit", Power: 1, Toughness: 1}, 1); err != nil {
			t.Fatal(err)
		}
	})
	if got := effectstest.CountOnBattlefield(g, "Spirit", opp.ID); got != 1 {
		t.Errorf("the opponent's token = %d, want 1", got)
	}
}

func TestC21B057NaruMehaCopiesYourSpellAndPumpsOtherWizards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, victimA, victimB := g.Seats[0].ID, g.Seats[1].ID, g.Seats[2].ID
	otherWizard := effectstest.PushCatalogPermanent(g, me, "Test Wizard", "Creature — Human Wizard", "", false)
	oppWizard := effectstest.PushCatalogPermanent(g, victimA, "Opp Wizard", "Creature — Human Wizard", "", false)
	bear := effectstest.PushCatalogPermanent(g, me, "Test Bear", "Creature — Bear", "", false)

	bolt := effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: victimA}})
	naru := effectstest.CastCatalogSpell(t, g, "Naru Meha, Master Wizard", "Legendary Creature — Human Wizard", c21B057Naru, nil)
	for i := 0; i < 8 && effectstest.LatestPickTarget(g, me) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	pick := effectstest.LatestPickTarget(g, me)
	if pick == nil {
		t.Fatal("the ETB should ask for the spell to copy")
	}
	effectstest.PickCard(t, g, me, bolt)
	for i := 0; i < 8 && effectstest.LatestPickTarget(g, me) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	retarget := effectstest.LatestPickTarget(g, me)
	if retarget == nil {
		t.Fatal("no re-target prompt for the copy")
	}
	if err := g.ResolvePickTarget(retarget.ID, me, game.TargetRef{Kind: game.TargetPlayer, ID: victimB}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LifeOf(g, victimA) != 37 || effectstest.LifeOf(g, victimB) != 37 {
		t.Errorf("lives %d / %d, want 37 / 37 (the Bolt and its copy)", effectstest.LifeOf(g, victimA), effectstest.LifeOf(g, victimB))
	}
	effectstest.AssertKeywords(t, g, naru, "flash")
	if p, tough := effectstest.EffectivePTOf(t, g, otherWizard); p != 2 || tough != 2 {
		t.Errorf("my other Wizard is %d/%d, want 2/2", p, tough)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, oppWizard); p != 1 || tough != 1 {
		t.Errorf("an opponent's Wizard is %d/%d, want 1/1", p, tough)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 1 || tough != 1 {
		t.Errorf("my non-Wizard is %d/%d, want 1/1", p, tough)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, naru); p != 0 || tough != 0 {
		t.Errorf("Naru Meha herself is %d/%d, want her own 0/0 test base with no bonus (other Wizards only)", p, tough)
	}
}

func TestC21B057NaruMehaOnlyTargetsYourOwnSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spec, _ := effects.Lookup(c21B057Naru)
	ts := spec.Triggered[0].Targets
	mine := game.Card{Name: "Lightning Bolt", TypeLine: "Instant", Controller: me.ID, Owner: me.ID}
	theirs := game.Card{Name: "Lightning Bolt", TypeLine: "Instant", Controller: opp.ID, Owner: opp.ID}
	g.ReadSnapshot(func() {
		if !ts.CardOK(g, me.ID, mine, game.ZoneStack) {
			t.Error("my own instant is not a legal target")
		}
		if ts.CardOK(g, me.ID, theirs, game.ZoneStack) {
			t.Error("an opponent's instant is a legal target")
		}
		if ts.CardOK(g, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Controller: me.ID}, game.ZoneStack) {
			t.Error("a creature spell is a legal target")
		}
	})
}

func c21B057ArchonOnTheBattlefield(t *testing.T, g *game.Game, chosenIdx int) uuid.UUID {
	t.Helper()
	me := g.Seats[0]
	id := effectstest.CastCatalogSpell(t, g, "Guardian Archon", "Creature — Archon", c21B057Archon, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, chosenIdx)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("Guardian Archon is not on the battlefield")
	}
	return id
}

func c21B057ProtectedFrom(g *game.Game, id, from uuid.UUID) bool {
	var out bool
	g.ReadSnapshot(func() {
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			if c.InstanceID != id {
				continue
			}
			for _, q := range game.ProtectionQualities(c) {
				if q.Kind == game.ProtectionQualityPlayer && q.Player == from {
					out = true
				}
			}
		}
	})
	return out
}

func c21B057PlayerProtectedFrom(g *game.Game, p *game.Player, from uuid.UUID) bool {
	var out bool
	g.ReadSnapshot(func() {
		_, out = g.PlayerProtectedFromLocked(p, &game.Characteristic{Controller: from})
	})
	return out
}

func TestC21B057GuardianArchonSecretlyChoosesThenProtectsYouAndAPermanentOnce(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, o1, o2 := g.Seats[0], g.Seats[1], g.Seats[2]
	ally := effectstest.PushCatalogPermanent(g, me.ID, "Test Bear", "Creature — Bear", "", false)
	archon := c21B057ArchonOnTheBattlefield(t, g, 1)
	effectstest.AssertKeywords(t, g, archon, "flying")
	if w, by := g.SecretChosenPlayerOf(archon); w != o2.ID || by != me.ID {
		t.Fatalf("secret choice = %v by %v, want %v by %v", w, by, o2.ID, me.ID)
	}
	if g.CanRevealSecretChoiceForEffect(o1.ID, archon) {
		t.Error("an opponent could reveal the choice")
	}
	if c21B057ProtectedFrom(g, ally, o2.ID) || c21B057PlayerProtectedFrom(g, me, o2.ID) {
		t.Fatal("protection before the reveal")
	}
	effectstest.Pr7Activate(t, g, me.ID, archon, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: ally}}})
	if !c21B057ProtectedFrom(g, ally, o2.ID) {
		t.Error("the target permanent has no protection from the chosen player")
	}
	if c21B057ProtectedFrom(g, ally, o1.ID) {
		t.Error("the target permanent is protected from an opponent who was not chosen")
	}
	if !c21B057PlayerProtectedFrom(g, me, o2.ID) {
		t.Error("the controller has no protection from the chosen player")
	}
	if c21B057PlayerProtectedFrom(g, me, o1.ID) {
		t.Error("the controller is protected from an opponent who was not chosen")
	}
	// "Activate only once".
	if err := g.ActivateCatalogAbility(me.ID, archon, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: ally}}}); err == nil {
		t.Error("the ability activated a second time")
	}
}

func TestC21B057GuardianArchonProtectionEndsAtCleanup(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, o1 := g.Seats[0], g.Seats[1]
	ally := effectstest.PushCatalogPermanent(g, me.ID, "Test Bear", "Creature — Bear", "", false)
	archon := c21B057ArchonOnTheBattlefield(t, g, 0)
	effectstest.Pr7Activate(t, g, me.ID, archon, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: ally}}})
	if !c21B057ProtectedFrom(g, ally, o1.ID) {
		t.Fatal("no protection granted")
	}
	effectstest.EndTurn(t, g)
	if c21B057ProtectedFrom(g, ally, o1.ID) || c21B057PlayerProtectedFrom(g, me, o1.ID) {
		t.Error("protection lasted past end of turn")
	}
}

func TestC21B057GuardianArchonRefusesAnOpponentsPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, o1 := g.Seats[0], g.Seats[1]
	theirs := effectstest.PushCatalogPermanent(g, o1.ID, "Their Bear", "Creature — Bear", "", false)
	archon := c21B057ArchonOnTheBattlefield(t, g, 0)
	if err := g.ActivateCatalogAbility(me.ID, archon, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirs}}}); err == nil {
		t.Error("activated on a permanent the controller does not control")
	}
}

func TestC21B057SaplingOfColfenorTakesTheTopCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, o1 := g.Seats[0], g.Seats[1]
	sap := effectstest.PushCatalogPermanent(g, me.ID, "Sapling of Colfenor", "Legendary Creature — Treefolk Shaman", c21B057Sapling, false)
	effectstest.AssertKeywords(t, g, sap, "indestructible")
	top := effectstest.PushLibraryCard(me, game.Card{Name: "Big Guy", TypeLine: "Creature — Bear", Power: 3, Toughness: 4})
	life, hand := me.Life, len(me.Hand.Cards)
	effectstest.DeclareAttack(t, g, o1.ID, sap)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+4-3 {
		t.Errorf("life = %d, want %d (gain 4 toughness, lose 3 power)", me.Life, life+1)
	}
	if !effectstest.InHandOf(me, top) || len(me.Hand.Cards) != hand+1 {
		t.Error("the creature card did not go to hand")
	}
}

func TestC21B057SaplingOfColfenorLeavesANoncreatureOnTop(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, o1 := g.Seats[0], g.Seats[1]
	sap := effectstest.PushCatalogPermanent(g, me.ID, "Sapling of Colfenor", "Legendary Creature — Treefolk Shaman", c21B057Sapling, false)
	top := effectstest.B07LibraryLand(me, "Island", "Basic Land — Island")
	life, hand := me.Life, len(me.Hand.Cards)
	effectstest.DeclareAttack(t, g, o1.ID, sap)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life || len(me.Hand.Cards) != hand {
		t.Errorf("life %d→%d, hand %d→%d: a land changes nothing", life, me.Life, hand, len(me.Hand.Cards))
	}
	if !me.Library.Contains(top) {
		t.Error("the revealed land left the library")
	}
}

func TestC21B057EssencePulseShrinksEveryCreatureByLifeGainedThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, o1 := g.Seats[0], g.Seats[1]
	mine := effectstest.PushCatalogPermanent(g, me.ID, "Small Mine", "Creature — Bear", "", false)
	theirsBig := effectstest.PushCatalogPermanent(g, o1.ID, "Big Theirs", "Creature — Bear", "", false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			switch g.Battlefield.Cards[i].InstanceID {
			case mine:
				g.Battlefield.Cards[i].Power, g.Battlefield.Cards[i].Toughness = 2, 2
			case theirsBig:
				g.Battlefield.Cards[i].Power, g.Battlefield.Cards[i].Toughness = 4, 4
			}
		}
	})
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Essence Pulse", "Sorcery", c21B057Pulse, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life = %d, want %d", me.Life, life+2)
	}
	if g.Battlefield.Contains(mine) {
		t.Error("my 2/2 survived -2/-2")
	}
	if p, tough := effectstest.EffectivePTOf(t, g, theirsBig); p != 2 || tough != 2 {
		t.Errorf("the 4/4 is %d/%d, want 2/2", p, tough)
	}
}

func TestC21B057EssencePulseCountsLifeGainedEarlierThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, o1 := g.Seats[0], g.Seats[1]
	big := effectstest.PushCatalogPermanent(g, o1.ID, "Big Theirs", "Creature — Bear", "", false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == big {
				g.Battlefield.Cards[i].Power, g.Battlefield.Cards[i].Toughness = 5, 5
			}
		}
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 3); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.CastCatalogSpell(t, g, "Essence Pulse", "Sorcery", c21B057Pulse, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) {
		t.Error("a 5/5 survived -5/-5 (3 gained earlier plus 2)")
	}
}

func c21B057Creature(t *testing.T, g *game.Game, p *game.Player, cost string) game.Card {
	t.Helper()
	c := game.Card{
		InstanceID: uuid.New(), Name: "Bear Cub", TypeLine: "Creature — Bear", ManaCost: cost,
		Colors: []string{"G"}, Power: 2, Toughness: 2, Owner: p.ID, Controller: p.ID,
	}
	p.Hand.PushTop(c)
	return c
}

func TestC21B057EzzarootChannelerDiscountsCreatureSpellsByLifeGained(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ch := effectstest.PushCatalogPermanent(g, me.ID, "Ezzaroot Channeler", "Creature — Treefolk Druid", c21B057Channeler, false)
	effectstest.AdvanceToMain(t, g)
	effectstest.PushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	effectstest.PushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	cub := c21B057Creature(t, g, me, "{3}{G}")
	// No life gained yet: {3}{G} cannot be paid with two Forests.
	if err := g.CastSpell(me.ID, cub.InstanceID, game.CastSpellParams{Strict: true, AutoTap: true}); err == nil {
		t.Fatal("cast a {3}{G} creature for two mana before gaining any life")
	}
	effectstest.Pr7Activate(t, g, me.ID, ch, 0, game.ActivateAbilityParams{})
	if me.Life != 42 {
		t.Fatalf("life = %d, want 42 after {T}: gain 2", me.Life)
	}
	// Two life gained: {3}{G} costs {1}{G}.
	if err := g.CastSpell(me.ID, cub.InstanceID, game.CastSpellParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("cast for {1}{G} after gaining 2: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(cub.InstanceID) {
		t.Error("the discounted creature did not resolve")
	}
}

func TestC21B057EzzarootChannelerDoesNotDiscountNoncreaturesOrOpponents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, o1 := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Ezzaroot Channeler", "Creature — Treefolk Druid", c21B057Channeler, false)
	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 3); err != nil {
			t.Fatal(err)
		}
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, o1.ID, 3); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.AdvanceToMain(t, g)
	effectstest.PushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	sorc := game.Card{InstanceID: uuid.New(), Name: "Plain Sorcery", TypeLine: "Sorcery", ManaCost: "{3}{G}", Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(sorc)
	if err := g.CastSpell(me.ID, sorc.InstanceID, game.CastSpellParams{Strict: true, AutoTap: true}); err == nil {
		t.Error("a noncreature spell was discounted")
	}
	// An opponent's creature spell is not "you cast".
	var gotOpp int
	g.ReadSnapshot(func() {
		spec, _ := effects.Lookup(c21B057Channeler)
		q := game.CostQuery{Game: g, Card: game.Card{TypeLine: "Creature — Bear"}, Controller: o1.ID, Source: game.Card{Controller: me.ID}}
		for _, m := range spec.CostModifiers {
			if m.AppliesTo(q) {
				gotOpp += m.Amount(q)
			}
		}
	})
	if gotOpp != 0 {
		t.Errorf("an opponent's creature spell was discounted by %d", gotOpp)
	}
}

func TestC21B057BlightedCataractDrawsTwoForSevenMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := effectstest.PushCatalogPermanent(g, me.ID, "Blighted Cataract", "Land", c21B057Cataract, false)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "U", "G", "G", "G", "G")
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with only five mana")
	}
	effectstest.B06AddMana(me, "G")
	hand := len(me.Hand.Cards)
	effectstest.B16Activate(t, g, me.ID, land, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if len(me.Hand.Cards) != hand+2 {
		t.Errorf("hand %d→%d, want +2", hand, len(me.Hand.Cards))
	}
	if g.Battlefield.Contains(land) {
		t.Error("the land was not sacrificed")
	}
}

func TestC21B057SapseepForestEntersTappedAndGainsLifeWithTwoGreenPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	forest := effectstest.PlayLandFromHand(t, g, "Sapseep Forest", c21B057Sapseep)
	if tapped, found := effectstest.BattlefieldCardTapped(g, forest); !found || !tapped {
		t.Fatalf("Sapseep Forest tapped=%v found=%v, want it to enter tapped", tapped, found)
	}
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == forest {
				g.Battlefield.Cards[i].Tapped = false
				g.Battlefield.Cards[i].TypeLine = "Land — Forest"
			}
		}
	})
	effectstest.AcPermanent(g, me.ID, "Green Bear", "Creature — Bear", 2, 2, "G")
	effectstest.B06AddMana(me, "G")
	effectstest.AcRefused(t, g, me.ID, forest, 0, game.ActivateAbilityParams{})
	effectstest.AcPermanent(g, me.ID, "Green Elf", "Creature — Elf", 1, 1, "G")
	life := me.Life
	effectstest.B16Activate(t, g, me.ID, forest, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Errorf("life %d→%d, want +1", life, me.Life)
	}
}

func TestC21B057PhyrexiasCoreSacrificesAnArtifactForLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	core := effectstest.PushCatalogPermanent(g, me.ID, "Phyrexia's Core", "Land", c21B057Core, false)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "G")
	if err := g.ActivateCatalogAbility(me.ID, core, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with no artifact to sacrifice")
	}
	rock := effectstest.PushCatalogPermanent(g, me.ID, "Test Rock", "Artifact", "", false)
	life := me.Life
	effectstest.B16Activate(t, g, me.ID, core, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{rock}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact was not sacrificed")
	}
	if me.Life != life+1 {
		t.Errorf("life %d→%d, want +1", life, me.Life)
	}
}
