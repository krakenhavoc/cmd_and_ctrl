package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b289_test.go — Shadows over Innistrad Auras, S82 OLD batch 289
// (#5543), played through the real catalog entries.

const (
	oldB289Hope    = "553b9606-0519-4b85-b6ae-acc126e537e1"
	oldB289Witch   = "80b07882-e144-4815-8b6f-04b3ab343d97"
	oldB289Skill   = "8c8c12ff-24ff-4279-9bde-da8899d2c2d5"
	oldB289Motives = "a8f259b5-ef01-4f30-baa1-1b2a5419e61e"
	oldB289Skin    = "b5a146e6-053f-43ce-8b01-24505a398976"
	oldB289Saint   = "cd6e9465-f4eb-482d-9c49-3a16b52c997c"
)

func oldB289WitchRow() cards.Card {
	return cards.Card{
		ID:       uuid.New(),
		Name:     "Accursed Witch // Infectious Curse",
		Layout:   "transform",
		TypeLine: "Creature — Human Shaman // Enchantment — Aura Curse",
		OracleID: uuid.MustParse(oldB289Witch),
		CardFaces: []cards.CardFace{
			{Name: "Accursed Witch", TypeLine: "Creature — Human Shaman", ManaCost: "{3}{B}", Colors: []string{"B"}, Power: "4", Toughness: "2"},
			{Name: "Infectious Curse", TypeLine: "Enchantment — Aura Curse", Colors: []string{"B"}},
		},
	}
}

func oldB289SkinRow() cards.Card {
	return cards.Card{
		ID:       uuid.New(),
		Name:     "Skin Invasion // Skin Shedder",
		Layout:   "transform",
		TypeLine: "Enchantment — Aura // Creature — Insect Horror",
		OracleID: uuid.MustParse(oldB289Skin),
		CardFaces: []cards.CardFace{
			{Name: "Skin Invasion", TypeLine: "Enchantment — Aura", ManaCost: "{R}", Colors: []string{"R"}},
			{Name: "Skin Shedder", TypeLine: "Creature — Insect Horror", Colors: []string{"R"}, Power: "3", Toughness: "4"},
		},
	}
}

func oldB289Attach(g *game.Game, owner, host uuid.UUID, name, oracle string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: effectstest.AuraTypeLine, OracleID: oracle,
		Owner: owner, Controller: owner, AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: host},
	})
}

func TestOldB289EveryCardIsRegisteredFull(t *testing.T) {
	for id, name := range map[string]string{
		oldB289Hope: "Hope Against Hope", oldB289Witch: "Accursed Witch", oldB289Witch + "#1": "Infectious Curse",
		oldB289Skill: "Equestrian Skill", oldB289Motives: "Spiteful Motives",
		oldB289Skin: "Skin Invasion", oldB289Skin + "#1": "Skin Shedder", oldB289Saint: "Invocation of Saint Traft",
	} {
		spec, ok := effects.Lookup(id)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s registration: %+v", name, spec)
		}
	}
}

func TestOldB289HopeAgainstHopeCountsYourCreaturesAndGivesHumansFirstStrike(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	human := effectstest.B12Creature(g, me.ID, "Human", "Creature — Human", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, g.Seats[1].ID, "Foe", "Creature — Bear", 2, 2)
	oldB289Attach(g, me.ID, human, "Hope Against Hope", oldB289Hope)
	if p, tough := effectstest.PtOf(t, g, human); p != 4 || tough != 4 {
		t.Errorf("the enchanted Human is %d/%d, want 4/4 (+2 for your two creatures)", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, human, "first strike") {
		t.Error("an enchanted Human lacks first strike")
	}
	oldB289Attach(g, me.ID, bear, "Hope Against Hope", oldB289Hope)
	if p, _ := effectstest.PtOf(t, g, bear); p != 4 {
		t.Errorf("the enchanted Bear has power %d, want 4", p)
	}
	if effectstest.HasEffectiveKeyword(t, g, bear, "first strike") {
		t.Error("an enchanted Bear has first strike")
	}
}

func TestOldB289EquestrianSkillPumpsAndGivesHumansTrample(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	human := effectstest.B12Creature(g, me.ID, "Human", "Creature — Human", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	oldB289Attach(g, me.ID, human, "Equestrian Skill", oldB289Skill)
	oldB289Attach(g, me.ID, bear, "Equestrian Skill", oldB289Skill)
	for _, id := range []uuid.UUID{human, bear} {
		if p, tough := effectstest.PtOf(t, g, id); p != 5 || tough != 5 {
			t.Errorf("an enchanted creature is %d/%d, want 5/5", p, tough)
		}
	}
	if !effectstest.HasEffectiveKeyword(t, g, human, "trample") {
		t.Error("an enchanted Human lacks trample")
	}
	if effectstest.HasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("an enchanted Bear has trample")
	}
}

func TestOldB289SpitefulMotivesHasFlashAndGivesPlusThreeAndFirstStrike(t *testing.T) {
	spec, _ := effects.Lookup(oldB289Motives)
	if len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != "flash" {
		t.Errorf("printed keywords %v, want flash", spec.PrintedKeywords)
	}
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.AuraBear(g, me.ID)
	effectstest.AuraCast(t, g, "Spiteful Motives", oldB289Motives, bear)
	if p, tough := effectstest.PtOf(t, g, bear); p != 5 || tough != 2 {
		t.Errorf("the enchanted Bear is %d/%d, want 5/2", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "first strike") {
		t.Error("the enchanted Bear lacks first strike")
	}
}

func TestOldB289InvocationOfSaintTraftMakesAnAttackingAngelThatIsExiled(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := effectstest.AuraBear(g, me.ID)
	effectstest.AuraCast(t, g, "Invocation of Saint Traft", oldB289Saint, bear)
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassAndSettleAttackSplit(t, g, me.ID)
	n, tapped := effectstest.CountTokensNamed(g, me.ID, "Angel")
	if n != 1 || tapped != 1 {
		t.Fatalf("%d Angel tokens, %d tapped; want 1 and 1", n, tapped)
	}
	var angel game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Angel" && c.Controller == me.ID {
			angel = c
		}
	}
	if p, tough := effectstest.PtOf(t, g, angel.InstanceID); p != 4 || tough != 4 {
		t.Errorf("the Angel is %d/%d, want 4/4", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, angel.InstanceID, "flying") || angel.AttackingTarget == uuid.Nil {
		t.Errorf("the Angel flies: %v, attacking: %v", effectstest.HasEffectiveKeyword(t, g, angel.InstanceID, "flying"), angel.AttackingTarget != uuid.Nil)
	}
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Angel"); n != 0 {
		t.Errorf("%d Angel tokens remain after end of combat, want 0", n)
	}
}

func oldB289CastSkin(t *testing.T, g *game.Game, me *game.Player, host uuid.UUID) uuid.UUID {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	skin := effectstest.ImportToHand(oldB289SkinRow(), me)
	effectstest.DcMana(t, g, me.ID, "{R}")
	if err := g.CastSpell(me.ID, skin, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(host)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	return skin
}

func TestOldB289SkinInvasionForcesTheEnchantedCreatureToAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.AuraBear(g, me.ID)
	oldB289CastSkin(t, g, me, bear)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	g.WithWriteLock(func() {
		if g.AttackRequirementsUnmetForEffect() == nil {
			t.Error("the enchanted creature owes an attack but nothing says so")
		}
	})
}

func TestOldB289SkinInvasionReturnsTransformedWhenTheHostDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := effectstest.RtaBear(g, opp.ID, "Their Bear")
	skin := oldB289CastSkin(t, g, me, bear)
	effectstest.KillCreature(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.AdCard(t, g, skin)
	if effectstest.AdZone(g, skin) != game.ZoneBattlefield || c.ActiveFace != 1 || c.Name != "Skin Shedder" || !c.IsCreature() || c.Controller != me.ID {
		t.Fatalf("returned as %q (%s) face %d in %v under %v, want Skin Shedder under its controller", c.Name, c.TypeLine, c.ActiveFace, effectstest.AdZone(g, skin), c.Controller)
	}
	if p, tough := effectstest.PtOf(t, g, skin); p != 3 || tough != 4 {
		t.Errorf("the Skin Shedder is %d/%d, want 3/4", p, tough)
	}
}

func TestOldB289AccursedWitchReturnsAsACurseOnAnOpponentAndDrainsThem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	witch := effectstest.ImportToBattlefield(t, g, oldB289WitchRow(), me)
	effectstest.RtaDestroy(t, g, witch)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no pending target pick")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: me.ID}); err == nil {
		t.Fatal("targeting yourself was accepted")
	}
	effectstest.PickTriggerTargetPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.AdCard(t, g, witch)
	if effectstest.AdZone(g, witch) != game.ZoneBattlefield || c.ActiveFace != 1 || c.Name != "Infectious Curse" || c.Controller != me.ID {
		t.Fatalf("returned as %q face %d in %v, want Infectious Curse under its controller", c.Name, c.ActiveFace, effectstest.AdZone(g, witch))
	}
	if c.AttachedTo.Kind != game.TargetPlayer || c.AttachedTo.ID != opp.ID {
		t.Fatalf("attached to %+v, want the opponent", c.AttachedTo)
	}
	life, mine := opp.Life, me.Life
	effectstest.IwNextUpkeep(t, g, map[int]int{})
	if g.Turn.ActiveSeat != 1 {
		t.Fatalf("it is seat %d's upkeep, want the enchanted opponent's", g.Turn.ActiveSeat)
	}
	if opp.Life != life-1 || me.Life != mine+1 {
		t.Errorf("opponent life %d -> %d, your life %d -> %d; want -1 and +1", life, opp.Life, mine, me.Life)
	}
}

func TestOldB289InfectiousCurseDiscountsSpellsThatTargetTheEnchantedPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	third := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Infectious Curse", TypeLine: "Enchantment — Aura Curse", OracleID: oldB289Witch,
		Owner: me.ID, Controller: me.ID, ActiveFace: 1, AttachedTo: game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID},
	})
	effectstest.AdvanceToMain(t, g)
	cast := func(target uuid.UUID) error {
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Incinerate", TypeLine: "Instant", ManaCost: "{1}{R}", OracleID: effectstest.P1IncinerateOracle, Owner: me.ID, Controller: me.ID})
		effectstest.FlashIfFund(g, me, "R")
		return g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: target}}})
	}
	if err := cast(third.ID); err == nil {
		t.Fatal("a {1}{R} spell at another player was cast for {R}")
	}
	if err := cast(opp.ID); err != nil {
		t.Fatalf("a {1}{R} spell at the enchanted player for {R}: %v", err)
	}
}

func TestOldB289AccursedWitchMakesOpponentsSpellsAtItCheaper(t *testing.T) {
	g, me, opp := effectstest.P7Table(t)
	witch := effectstest.B12Push(g, me.ID, "Accursed Witch", "Creature — Human Shaman", oldB289Witch, 4, 2)
	effectstest.AdvanceToMain(t, g)
	for i := 0; i < len(g.Seats); i++ {
		if g.Turn.ActiveSeat == 1 {
			break
		}
		effectstest.EndTurn(t, g)
	}
	effectstest.AdvanceToMain(t, g)
	other := effectstest.B12Creature(g, me.ID, "Other", "Creature — Bear", 2, 2)
	cast := func(target uuid.UUID) error {
		id := uuid.New()
		opp.Hand.PushTop(game.Card{InstanceID: id, Name: "Incinerate", TypeLine: "Instant", ManaCost: "{1}{R}", OracleID: effectstest.P1IncinerateOracle, Owner: opp.ID, Controller: opp.ID})
		effectstest.FlashIfFund(g, opp, "R")
		return g.CastSpell(opp.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(target)})
	}
	if err := cast(other); err == nil {
		t.Fatal("a {1}{R} spell at another creature was cast for {R}")
	}
	if err := cast(witch); err != nil {
		t.Fatalf("a {1}{R} spell at the Witch for {R}: %v", err)
	}
}
