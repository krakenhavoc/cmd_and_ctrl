package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b647_test.go — Invasion, "Activated abilities" (S82 OLD batch 647, #5616): each
// card driven through a real activation, cast or trigger.

func oldB647AllFull(t *testing.T, ids ...string) {
	t.Helper()
	for _, id := range ids {
		spec, ok := effects.Lookup(id)
		if !ok {
			t.Errorf("%s is not registered", id)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v, caveats %v", spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func oldB647Ready(g *game.Game, id uuid.UUID) {
	effectstest.UntapForTest(g, id)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
}

// oldB647Perm seeds a ready catalog permanent for the active seat.
func oldB647Perm(t *testing.T, g *game.Game, name, typeLine, oracle string) uuid.UUID {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	id := effectstest.SeedPermanentWithOracle(g, g.Seats[0].ID, name, typeLine, oracle)
	oldB647Ready(g, id)
	return id
}

func oldB647Activate(t *testing.T, g *game.Game, src uuid.UUID, ability int, mana string, params game.ActivateAbilityParams) error {
	t.Helper()
	me := g.Seats[0]
	if mana != "" {
		effectstest.FloatMana(t, g, me, mana)
	}
	params.Strict = true
	return g.ActivateCatalogAbility(me.ID, src, ability, params)
}

func oldB647Effective(g *game.Game, id uuid.UUID) game.Card {
	c, _ := g.LookupCardForEffect(id)
	return c
}

func oldB647Target(id uuid.UUID) game.ActivateAbilityParams {
	return game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: id}}}
}

const (
	oldB647Healer      = "a5ef15b5-83d2-4b23-9403-a9a52be46154"
	oldB647Elephant    = "aaaaf370-6486-4e17-9a93-92e05ebcf29f"
	oldB647Zombie      = "c631887c-e157-4a95-8a59-08a44854d07f"
	oldB647Serpentine  = "cf252d71-77e7-438f-b774-28e5ab2f09f6"
	oldB647Trapper     = "d45c6c3f-6079-40c1-9083-06f2f2431bcc"
	oldB647Acolyte     = "dd1e7e89-72d7-44d7-922c-f08f0b7d5166"
	oldB647Apprentice  = "ebb68e03-b2a5-4c5c-8e3b-968cd241d1f0"
	oldB647Infiltrator = "f444a0f8-32d8-4748-af58-51a1a8ed72b4"
	oldB647Vision      = "f5276e83-ca3f-4755-9239-e44d91a446fd"
	oldB647Seer        = "f905b72d-734a-44e8-b118-c8de889994a9"
)

func TestS82OldB647AllAreFullWithNoCaveats(t *testing.T) {
	oldB647AllFull(t, oldB647Healer, oldB647Elephant, oldB647Zombie, oldB647Serpentine, oldB647Trapper, oldB647Acolyte,
		oldB647Apprentice, oldB647Infiltrator, oldB647Vision, oldB647Seer)
}

func oldB647Colored(g *game.Game, owner uuid.UUID, name, color string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Bear", Colors: []string{color},
		Power: 2, Toughness: 2, Owner: owner, Controller: owner,
	})
}

func TestS82OldB647SelfGrantAndRegenerateRows(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	kavu := oldB647Perm(t, g, "Serpentine Kavu", "Creature — Kavu", oldB647Serpentine)
	if effectstest.P1g2HasKeyword(g, kavu, "haste") {
		t.Fatal("the Kavu starts with haste")
	}
	if err := oldB647Activate(t, g, kavu, 0, "", game.ActivateAbilityParams{}); err == nil {
		t.Error("activated without mana")
	}
	if err := oldB647Activate(t, g, kavu, 0, "{R}", game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.P1g2HasKeyword(g, kavu, "haste") {
		t.Error("no haste")
	}
	g2 := effectstest.NewCatalogGame(t)
	z := oldB647Perm(t, g2, "Metathran Zombie", "Creature — Metathran Zombie", oldB647Zombie)
	if err := oldB647Activate(t, g2, z, 0, "", game.ActivateAbilityParams{}); err == nil {
		t.Error("regenerated without paying {B}")
	}
	if err := oldB647Activate(t, g2, z, 0, "{B}", game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("regenerate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g2)
	if !(g2.RegenerationShieldsOn(z) > 0) {
		t.Error("no regeneration shield")
	}
}

func TestS82OldB647TreefolkHealerPreventsTwoDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	healer := oldB647Perm(t, g, "Treefolk Healer", "Creature — Treefolk Cleric", oldB647Healer)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 4, 4)
	if err := oldB647Activate(t, g, healer, 0, "{C}{C}{W}", oldB647Target(bear)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.Pr6Damage(t, g, healer, bear, 3)
	if got := effectstest.Pr6Marked(g, bear); got != 1 {
		t.Errorf("marked damage %d, want 1 (3 damage with 2 prevented)", got)
	}
	if !effectstest.TapCostTapped(g, healer) {
		t.Error("the {T} did not tap the Healer")
	}
}

func TestS82OldB647BenalishTrapperAndThornscapeApprenticeTap(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	trapper := oldB647Perm(t, g, "Benalish Trapper", "Creature — Human Soldier", oldB647Trapper)
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	if err := oldB647Activate(t, g, trapper, 0, "", oldB647Target(ogre)); err == nil {
		t.Fatal("activated without paying {W}")
	}
	if err := oldB647Activate(t, g, trapper, 0, "{W}", oldB647Target(ogre)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !oldB647Effective(g, ogre).Tapped {
		t.Error("the Ogre is untapped")
	}
	g2 := effectstest.NewCatalogGame(t)
	opp2 := g2.Seats[1]
	app := oldB647Perm(t, g2, "Thornscape Apprentice", "Creature — Human Wizard", oldB647Apprentice)
	ogre2 := effectstest.B12Creature(g2, opp2.ID, "Ogre", "Creature — Ogre", 3, 3)
	if err := oldB647Activate(t, g2, app, 0, "{R}", oldB647Target(ogre2)); err != nil {
		t.Fatalf("first strike: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g2)
	if !effectstest.P1g2HasKeyword(g2, ogre2, "first strike") {
		t.Error("no first strike")
	}
	oldB647Ready(g2, app)
	if err := oldB647Activate(t, g2, app, 1, "{W}", oldB647Target(ogre2)); err != nil {
		t.Fatalf("tap: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g2)
	if !oldB647Effective(g2, ogre2).Tapped {
		t.Error("the Ogre is untapped")
	}
}

func TestS82OldB647ObsidianAcolyteGivesProtectionFromBlack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ac := oldB647Perm(t, g, "Obsidian Acolyte", "Creature — Human Cleric", oldB647Acolyte)
	bear := oldB647Colored(g, me.ID, "Bear", "G")
	if err := oldB647Activate(t, g, ac, 0, "{W}", oldB647Target(bear)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.P1g2HasKeyword(g, bear, game.ProtectionFromColor("B")) {
		t.Error("no protection from black")
	}
	if !effectstest.P1g2HasKeyword(g, ac, game.ProtectionFromColor("B")) {
		t.Error("the Acolyte lacks its printed protection from black")
	}
}

func TestS82OldB647RampantElephantForcesABlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	ele := effectstest.B12Push(g, me.ID, "Rampant Elephant", "Creature — Elephant", oldB647Elephant, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	effectstest.AdvanceToMain(t, g)
	if err := oldB647Activate(t, g, ele, 0, "", oldB647Target(theirs)); err == nil {
		t.Fatal("activated without paying {G}")
	}
	if err := oldB647Activate(t, g, ele, 0, "{G}", oldB647Target(theirs)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.DeclareAttack(t, g, opp.ID, ele)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlockers(nil); err == nil {
		t.Error("the opponent declared no blockers although the target must block if able")
	}
}

func TestS82OldB647PhyrexianInfiltratorExchangesControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	opp := g.Seats[1]
	inf := oldB647Perm(t, g, "Phyrexian Infiltrator", "Creature — Phyrexian Minion", oldB647Infiltrator)
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	if err := oldB647Activate(t, g, inf, 0, "{C}{C}{U}", oldB647Target(ogre)); err == nil {
		t.Fatal("activated without {U}{U}")
	}
	if err := oldB647Activate(t, g, inf, 0, "{U}", oldB647Target(ogre)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if oldB647Effective(g, ogre).Controller != me.ID {
		t.Error("the Ogre is not under my control")
	}
	if oldB647Effective(g, inf).Controller != opp.ID {
		t.Error("the Infiltrator is not under the opponent's control")
	}
}

func TestS82OldB647SeersVisionRevealsHandsAndDiscardsAtSorcerySpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	vision := effectstest.SeedPermanentWithOracle(g, g.Seats[0].ID, "Seer's Vision", "Enchantment", oldB647Vision)
	spec, _ := effects.Lookup(oldB647Vision)
	if spec.HandsRevealed != game.HandRevealOpponents {
		t.Errorf("hands revealed %v, want opponents'", spec.HandsRevealed)
	}
	keep := effectstest.HandSpell(opp, "Keep", "Sorcery", "")
	toss := effectstest.HandSpell(opp, "Toss", "Sorcery", "")
	hand := opp.Hand.Size()
	if err := oldB647Activate(t, g, vision, 0, "", game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.OpenVariant(t, g)
	if len(pick.DiscardOptions) != hand || pick.FromPlayer != opp.ID {
		t.Fatalf("pick offers %d cards from %v, want the target's whole hand of %d", len(pick.DiscardOptions), pick.FromPlayer, hand)
	}
	if err := g.ResolvePendingChoice(pick.ID, g.Seats[0].ID, []uuid.UUID{toss}); err != nil {
		t.Fatalf("ResolvePendingChoice: %v", err)
	}
	if g.Battlefield.Contains(vision) {
		t.Error("the Vision was not sacrificed")
	}
	if opp.Hand.Size() != hand-1 || !opp.Graveyard.Contains(toss) || opp.Graveyard.Contains(keep) {
		t.Errorf("hand %d graveyard has toss=%v keep=%v, want only Toss discarded", opp.Hand.Size(), opp.Graveyard.Contains(toss), opp.Graveyard.Contains(keep))
	}
}

func TestS82OldB647BlindSeerRecolorsAPermanentAndASpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	opp := g.Seats[1]
	seer := oldB647Perm(t, g, "Blind Seer", "Legendary Creature — Human Wizard", oldB647Seer)
	ogre := oldB647Colored(g, opp.ID, "Ogre", "R")
	if err := oldB647Activate(t, g, seer, 0, "", oldB647Target(ogre)); err == nil {
		t.Fatal("activated without paying {1}{U}")
	}
	if err := oldB647Activate(t, g, seer, 0, "{C}{U}", oldB647Target(ogre)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerColor(t, g, me.ID, "G")
	c := oldB647Effective(g, ogre)
	if !c.HasColor("G") || c.HasColor("R") {
		t.Errorf("colours %v, want only green", c.Colors)
	}
	// A spell on the stack: a blue Bolt cannot hit a creature with protection from blue.
	g2 := effectstest.NewCatalogGame(t)
	effectstest.AdvanceTo(t, g2, game.StepPrecombatMain)
	me2, opp2, _ := effectstest.BlbB1Seats(g2)
	seer2 := effectstest.SeedPermanentWithOracle(g2, me2.ID, "Blind Seer", "Legendary Creature — Human Wizard", oldB647Seer)
	oldB647Ready(g2, seer2)
	bear := effectstest.PushBattlefieldCardWithTimestamp(g2, game.Card{InstanceID: uuid.New(), Name: "Warded Bear",
		TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Keywords: []string{"protection from blue"}, Owner: opp2.ID, Controller: opp2.ID})
	bolt := effectstest.CastFor(t, g2, me2, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		game.CastSpellParams{Targets: effectstest.RtCardTarget(bear)})
	effectstest.FloatMana(t, g2, me2, "{C}{U}")
	if err := g2.ActivateCatalogAbility(me2.ID, seer2, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.RtCardTarget(bolt)}); err != nil {
		t.Fatalf("activate on a spell: %v", err)
	}
	for i := 0; i < 8 && effectstest.PendingOfKind(g2, game.PendingChoiceColor) == nil; i++ {
		if err := g2.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	effectstest.AnswerColor(t, g2, me2.ID, "U")
	effectstest.PassPriorityAroundTable(t, g2)
	if !g2.Battlefield.Contains(bear) {
		t.Error("the blue Bolt hit a creature with protection from blue")
	}
}
