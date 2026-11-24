package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	oldB10Master  = "07d40173-53e7-4328-a3b0-6214d93855b6"
	oldB10Jaws    = "1202725c-170b-42d7-9f82-a431a987f30e"
	oldB10RinSeri = "28c46eb7-bdfc-44b9-b7e9-b073260851f0"
	oldB10Malik   = "416608a1-89a4-44ce-8fd0-fab6b023d3d4"
	oldB10Enkira  = "558fd7a9-9b39-4c82-b909-5febf55dc010"
	oldB10Voice   = "5cbbb3f3-63a4-4983-81ea-8c405b10e63f"
)

func TestOldB10CardsAreFullWithNoCaveats(t *testing.T) {
	want := map[string]string{
		oldB10Master: "Master of Waves", oldB10Jaws: "Jaws, Relentless Predator",
		oldB10RinSeri: "Rin and Seri, Inseparable", oldB10Malik: "Malik, Grim Manipulator",
		oldB10Enkira: "Enkira, Hostile Scavenger", oldB10Voice: "Voice of Resurgence",
	}
	for oracle, name := range want {
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

func TestOldB10MasterOfWavesMakesAnElementalPerDevotionAndBuffsThem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ToMain(t, g)
	// Two blue pips already on the table; Master's own {U} makes three.
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Deep Thought", TypeLine: "Enchantment", ManaCost: "{1}{U}{U}",
		Colors: []string{"U"}, Owner: me.ID, Controller: me.ID,
	})
	foe := effectstest.B12Creature(g, opp.ID, "Foe Elemental", "Creature — Elemental", 1, 1)
	_, err := effectstest.Msc17Cast(t, g, me, game.Card{
		Name: "Master of Waves", TypeLine: "Creature — Merfolk Wizard", OracleID: oldB10Master,
		ManaCost: "{3}{U}", Colors: []string{"U"}, Power: 2, Toughness: 1,
	}, game.CastSpellParams{})
	if err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountOnBattlefield(g, "Elemental", me.ID); got != 3 {
		t.Fatalf("Elemental tokens: %d, want 3 (devotion to blue)", got)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Elemental" && c.Controller == me.ID {
			eff := c.Effective()
			if eff.Power != 2 || eff.Toughness != 1 {
				t.Errorf("token is %d/%d, want the 1/0 lifted to 2/1", eff.Power, eff.Toughness)
			}
		}
	}
	if c := effectstest.B12Card(t, g, foe); c.Effective().Power != 1 {
		t.Errorf("an opponent's Elemental got the bonus: power %d", c.Effective().Power)
	}
	if !effectstest.HasKeywordOf(g, g.Battlefield.Cards[0].InstanceID, "protection from red") && !hasProtection(g) {
		t.Error("Master of Waves lacks protection from red")
	}
}

func hasProtection(g *game.Game) bool {
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Master of Waves" {
			for _, k := range c.Effective().Abilities {
				if k == "protection from red" {
					return true
				}
			}
		}
	}
	return false
}

func TestOldB10VoiceOfResurgenceLeavesAnElementalCountingCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMain(t, g)
	voice := effectstest.B12Push(g, me.ID, "Voice of Resurgence", "Creature — Elemental", oldB10Voice, 2, 2)
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(voice) })
	effectstest.PassPriorityAroundTable(t, g)
	var token *game.Card
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].Name == "Elemental" && g.Battlefield.Cards[i].Controller == me.ID {
			token = &g.Battlefield.Cards[i]
		}
	}
	if token == nil {
		t.Fatal("no Elemental token after Voice died")
	}
	eff := token.Effective()
	if eff.Power != 2 || eff.Toughness != 2 {
		t.Errorf("token is %d/%d, want 2/2 (the Bear and itself)", eff.Power, eff.Toughness)
	}
	if len(token.Colors) != 2 {
		t.Errorf("token colours %v, want green and white", token.Colors)
	}
}

func TestOldB10RinAndSeriCastTriggersAndActivation(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ToMain(t, g)
	rin := effectstest.B12Push(g, me.ID, "Rin and Seri, Inseparable", "Legendary Creature — Dog Cat", oldB10RinSeri, 4, 4)
	// A Dog spell makes a Cat; a Cat spell makes a Dog.
	if _, err := effectstest.Msc17Cast(t, g, me, game.Card{Name: "Hound", TypeLine: "Creature — Dog", Power: 1, Toughness: 1}, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountOnBattlefield(g, "Cat", me.ID); got != 1 {
		t.Errorf("Cats after a Dog spell: %d, want 1", got)
	}
	if _, err := effectstest.Msc17Cast(t, g, me, game.Card{Name: "Tabby", TypeLine: "Creature — Cat", Power: 1, Toughness: 1}, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountOnBattlefield(g, "Dog", me.ID); got != 1 {
		t.Errorf("token Dogs after a Cat spell: %d, want 1", got)
	}
	// Dogs: Rin and Seri, Hound, the token Dog = 3. Cats: Rin and Seri, Tabby, the token Cat = 3.
	life, oppLife := me.Life, opp.Life
	effectstest.FloatForTest(g, me, "RGW")
	if err := g.ActivateCatalogAbility(me.ID, rin, 0, game.ActivateAbilityParams{
		Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if oppLife-opp.Life != 3 || me.Life-life != 3 {
		t.Errorf("damage %d and life %d, want 3 and 3 (Dogs and Cats)", oppLife-opp.Life, me.Life-life)
	}
}

func TestOldB10JawsMakesBloodAndPingsForNoncreatureArtifacts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ToMain(t, g)
	effectstest.B12Push(g, me.ID, "Jaws, Relentless Predator", "Legendary Creature — Shark", oldB10Jaws, 5, 5)
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	golem := effectstest.B12Creature(g, opp.ID, "Golem", "Artifact Creature — Golem", 1, 1)
	before := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		before[p.ID] = p.Life
	}
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(golem) })
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		if p.Life != before[p.ID] {
			t.Errorf("a creature artifact left and %v lost life", p.Name)
		}
	}
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(rock) })
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		want := before[p.ID]
		if p.ID != me.ID {
			want--
		}
		if p.Life != want {
			t.Errorf("%s life %d, want %d after a sacrificed noncreature artifact", p.Name, p.Life, want)
		}
	}
	rock2 := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(rock2) })
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before[opp.ID]-2 {
		t.Errorf("opponent life %d, want %d after a destroyed artifact too", opp.Life, before[opp.ID]-2)
	}
}

func TestOldB10JawsCombatDamageMakesThatManyBlood(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	jaws := effectstest.B12Push(g, me.ID, "Jaws, Relentless Predator", "Legendary Creature — Shark", oldB10Jaws, 5, 5)
	effectstest.AttackWith(t, g, opp.ID, jaws)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountOnBattlefield(g, "Blood", me.ID); got != 5 {
		t.Errorf("Blood tokens: %d, want 5", got)
	}
}

func TestOldB10EnkiraMakesTwoWalkersWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMain(t, g)
	if _, err := effectstest.Msc17Cast(t, g, me, game.Card{
		Name: "Enkira, Hostile Scavenger", TypeLine: "Legendary Creature — Human Warrior", OracleID: oldB10Enkira, Power: 3, Toughness: 3,
	}, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Walker" && c.Controller == me.ID {
			n++
			if c.Power != 2 || c.Toughness != 2 || !c.HasSubtype("Zombie") {
				t.Errorf("Walker %d/%d %q, want a 2/2 Zombie", c.Power, c.Toughness, c.TypeLine)
			}
		}
	}
	if n != 2 {
		t.Errorf("Walkers: %d, want 2", n)
	}
}

func TestOldB10EnkiraGainsIndestructibleOnlyWithTwoZombiesAttacking(t *testing.T) {
	for _, zombies := range []int{1, 2} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		enkira := effectstest.B12Push(g, me.ID, "Enkira, Hostile Scavenger", "Legendary Creature — Human Warrior", oldB10Enkira, 3, 3)
		attackers := []uuid.UUID{enkira}
		for i := 0; i < zombies; i++ {
			attackers = append(attackers, effectstest.PushToken(g, me.ID, effects.TokenCard("2/2 black Walker")))
		}
		effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
		for _, id := range attackers {
			if err := g.DeclareAttacker(id, opp.ID); err != nil {
				t.Fatalf("DeclareAttacker: %v", err)
			}
		}
		effectstest.LockInAttacks(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		got := effectstest.HasKeywordOf(g, enkira, "indestructible")
		if got != (zombies == 2) {
			t.Errorf("attacking with %d Zombies: indestructible = %v", zombies, got)
		}
	}
}

func TestOldB10MalikSacrificesTheSecretlyChosenCreaturesAndMakesTreasure(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ToMain(t, g)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	wolf := effectstest.B12Creature(g, opp.ID, "Wolf", "Creature — Wolf", 2, 2)
	ox := effectstest.B12Creature(g, opp.ID, "Ox", "Creature — Ox", 2, 2)
	if _, err := effectstest.Msc17Cast(t, g, me, game.Card{
		Name: "Malik, Grim Manipulator", TypeLine: "Legendary Creature — Human Rogue", OracleID: oldB10Malik,
		Power: 4, Toughness: 3,
	}, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatalf("no target prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	// The controller picks the Bear, the opponent picks the Wolf: both go.
	effectstest.VoteFor(t, g, me.ID, "Bear")
	effectstest.VoteFor(t, g, opp.ID, "Wolf")
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || g.Battlefield.Contains(wolf) {
		t.Error("a chosen creature was not sacrificed")
	}
	if !g.Battlefield.Contains(ox) {
		t.Error("an unchosen creature was sacrificed")
	}
	if got := effectstest.CountOnBattlefield(g, "Treasure", me.ID); got != 2 {
		t.Errorf("Treasures: %d, want one per creature the opponent sacrificed (2)", got)
	}
}

func TestOldB10EnkiraMustBeBlockedOnlyWhileEquipped(t *testing.T) {
	spec, _ := effects.Lookup(oldB10Enkira)
	if len(spec.Static) != 1 {
		t.Fatalf("%d statics, want the one block requirement", len(spec.Static))
	}
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	enkira := effectstest.B12Push(g, me.ID, "Enkira, Hostile Scavenger", "Legendary Creature — Human Warrior", oldB10Enkira, 3, 3)
	applies := func() bool {
		var ok bool
		g.WithWriteLock(func() {
			c := effectstest.BattlefieldCardPtr(t, g, enkira)
			ok = spec.Static[0].AppliesTo(c, g, c)
		})
		return ok
	}
	if applies() {
		t.Error("the block requirement applies to an unequipped Enkira")
	}
	sword := effectstest.B12Permanent(g, me.ID, "Sword", "Artifact — Equipment")
	var err error
	g.WithWriteLock(func() { err = g.AttachForEffect(sword, game.TargetRef{Kind: game.TargetCard, ID: enkira}) })
	if err != nil {
		t.Fatal(err)
	}
	if !applies() {
		t.Error("the block requirement does not apply to an equipped Enkira")
	}
}
