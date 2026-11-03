package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soc_b11_loremaster_test.go — Quintorius, Loremaster and Redoubled
// Stormsinger from Secrets of Strixhaven Commander batch 11 (#4834).

// --- Quintorius, Loremaster ---

func socB11PushLoremaster(g *game.Game, me *game.Player) uuid.UUID {
	return pushCatalogPermanent(g, me.ID, "Quintorius, Loremaster", "Legendary Creature — Elephant Cleric", socB11LoremasterOracle, false)
}

// socB11EndStepExile runs my end step with the Loremaster out and the
// instant in the graveyard, picking it as the target if asked.
func socB11EndStepExile(t *testing.T, g *game.Game, me *game.Player, spell uuid.UUID, banned ...uuid.UUID) {
	t.Helper()
	advanceToEndStepOf(t, g, 0)
	if p := latestPickTarget(g, me.ID); p != nil {
		for _, id := range p.PickTargetCards {
			for _, b := range banned {
				if id == b {
					t.Errorf("the creature or land card %s was offered as a target", id)
				}
			}
		}
		pickCard(t, g, me.ID, spell)
	}
	passPriorityAroundTable(t, g)
}

func TestSocB11LoremasterExilesANoncreatureNonlandCardAndMakesASpirit(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	lore := socB11PushLoremaster(g, me)
	assertKeywords(t, g, lore, "vigilance")
	spell := socB11Graveyard(me, "Spent Spell", "Instant", "{R}")
	creature := socB11GraveyardCreature(me, "Dead Bear")
	land := socB11Graveyard(me, "Plains", "Basic Land — Plains", "")

	socB11EndStepExile(t, g, me, spell, creature, land)
	if !g.Exile.Contains(spell) {
		t.Fatal("the instant was not exiled")
	}
	if !me.Graveyard.Contains(creature) || !me.Graveyard.Contains(land) {
		t.Error("the creature or land card left the graveyard")
	}
	if got := b43TokensNamed(g, me.ID, "Spirit"); got != 1 {
		t.Errorf("%d Spirit tokens, want 1", got)
	}
}

func TestSocB11LoremasterWithNoTargetMakesNoSpirit(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	lore := socB11PushLoremaster(g, me)
	socB11GraveyardCreature(me, "Dead Bear")
	socB11Graveyard(me, "Plains", "Basic Land — Plains", "")
	advanceToEndStepOf(t, g, 0)
	if triggerOnStack(g, lore) != nil || latestPickTarget(g, me.ID) != nil {
		t.Error("the trigger was put on the stack with no legal target")
	}
	passPriorityAroundTable(t, g)
	if got := b43TokensNamed(g, me.ID, "Spirit"); got != 0 {
		t.Errorf("%d Spirit tokens with no target, want 0", got)
	}
}

func TestSocB11LoremasterTargetGoneMeansNoSpirit(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	socB11PushLoremaster(g, me)
	spell := socB11Graveyard(me, "Spent Spell", "Instant", "{R}")
	advanceToEndStepOf(t, g, 0)
	if p := latestPickTarget(g, me.ID); p != nil {
		pickCard(t, g, me.ID, spell)
	}
	// The card leaves in response: the ability does nothing at all.
	socB11Move(t, g, me.ID, spell, game.ZoneGraveyard, game.ZoneHand)
	passPriorityAroundTable(t, g)
	if got := b43TokensNamed(g, me.ID, "Spirit"); got != 0 {
		t.Errorf("%d Spirit tokens after the target left, want 0 (ruling 2023-04-14)", got)
	}
}

// socB11LoremasterCastSetup exiles `kind` with the end-step trigger, then
// walks to my next main phase with mana and a Spirit ready, and runs the
// activated ability on the exiled card.
func socB11LoremasterCastSetup(t *testing.T, g *game.Game, me *game.Player, kind string) (spell, lore uuid.UUID) {
	t.Helper()
	lore = socB11PushLoremaster(g, me)
	spell = socB11Graveyard(me, "Spent Spell", kind, "{R}")
	socB11EndStepExile(t, g, me, spell)
	advanceToStepOf(t, g, 1, game.StepPrecombatMain)
	advanceToStepOf(t, g, 0, game.StepPrecombatMain)
	spirits := socB11Named(g, "Spirit")
	if len(spirits) != 1 {
		t.Fatalf("%d Spirits to sacrifice, want 1", len(spirits))
	}
	b06AddMana(me, "C", "R", "W")
	if err := g.ActivateCatalogAbility(me.ID, lore, 0, game.ActivateAbilityParams{
		Strict:       true,
		Targets:      []game.TargetRef{{Kind: game.TargetCard, ID: spell}},
		SacrificeIDs: spirits,
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(spirits[0]) {
		t.Fatal("the Spirit was not sacrificed as a cost")
	}
	passPriorityAroundTable(t, g)
	return spell, lore
}

func TestSocB11LoremasterCastsTheExiledCardFreeAndBottomsItInsteadOfTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	spell, _ := socB11LoremasterCastSetup(t, g, me, "Instant")
	me.ManaPool = nil
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{FromZone: "exile", AlternativeCost: game.AltCostKeyGrantedCost, Strict: true}); err != nil {
		t.Fatalf("free cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Graveyard.Contains(spell) || g.Exile.Contains(spell) {
		t.Fatal("the resolved spell went to the graveyard or stayed in exile")
	}
	bottom, err := me.Library.Bottom()
	if err != nil || bottom.InstanceID != spell {
		t.Errorf("the spell is not on the bottom of its owner's library (bottom %v, err %v)", bottom.Name, err)
	}
}

func TestSocB11LoremasterCounteredSpellAlsoGoesToTheBottom(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	spell, _ := socB11LoremasterCastSetup(t, g, me, "Instant")
	me.ManaPool = nil
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{FromZone: "exile", AlternativeCost: game.AltCostKeyGrantedCost, Strict: true}); err != nil {
		t.Fatalf("free cast: %v", err)
	}
	if err := g.CounterSpell(spell, nil); err != nil {
		t.Fatalf("counter: %v", err)
	}
	if me.Graveyard.Contains(spell) {
		t.Fatal("the countered spell went to the graveyard")
	}
	bottom, err := me.Library.Bottom()
	if err != nil || bottom.InstanceID != spell {
		t.Errorf("the countered spell is not on the bottom of its owner's library (bottom %v, err %v)", bottom.Name, err)
	}
}

func TestSocB11LoremasterCastFollowsTheCardsOwnTiming(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	spell, _ := socB11LoremasterCastSetup(t, g, me, "Sorcery")
	me.ManaPool = nil
	// A sorcery can be cast only in a main phase with an empty stack
	// (ruling 2023-04-14).
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{FromZone: "exile", AlternativeCost: game.AltCostKeyGrantedCost, Strict: true}); err == nil {
		t.Error("a sorcery was cast during combat")
	}
}

func TestSocB11LoremasterOnlyTargetsCardsItExiled(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	lore := socB11PushLoremaster(g, me)
	advanceToMain(t, g)
	socB11Token(t, g, me.ID, "3/2 red and white Spirit")
	other := uuid.New()
	g.Exile.PushTop(game.Card{InstanceID: other, Name: "Foreign Spell", TypeLine: "Instant", ManaCost: "{R}", Owner: me.ID, Controller: me.ID})
	b06AddMana(me, "C", "R", "W")
	if err := g.ActivateCatalogAbility(me.ID, lore, 0, game.ActivateAbilityParams{
		Strict:       true,
		Targets:      []game.TargetRef{{Kind: game.TargetCard, ID: other}},
		SacrificeIDs: socB11Named(g, "Spirit"),
	}); err == nil {
		t.Error("targeted a card Quintorius did not exile")
	}
}

// --- Redoubled Stormsinger ---

func socB11StormsingerAttack(t *testing.T, g *game.Game, me, opp *game.Player) uuid.UUID {
	t.Helper()
	advanceToMain(t, g)
	stormsinger := b12Push(g, me.ID, "Redoubled Stormsinger", "Creature — Orc Wizard", socB11StormsingerOracle, 3, 3)
	assertKeywords(t, g, stormsinger, "first strike")
	return stormsinger
}

func TestSocB11StormsingerCopiesEachCreatureTokenThatEnteredThisTurnTappedAndAttacking(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	stormsinger := socB11StormsingerAttack(t, g, me, opp)
	socB11Token(t, g, me.ID, "3/2 red and white Spirit")
	socB11Token(t, g, me.ID, "1/1 white Soldier")
	// None of these count: a Treasure is not a creature, an opponent's token
	// is not yours, and a nontoken creature is not a token.
	g.WithWriteLock(func() {
		ctx := NewContext(g, &game.StackItem{ID: uuid.New(), Controller: me.ID, SourceCardID: uuid.Nil})
		if err := (CreateToken{Template: TreasureToken(), N: 1, Controller: me.ID}).Apply(ctx); err != nil {
			t.Fatalf("treasure: %v", err)
		}
	})
	socB11Token(t, g, opp.ID, "1/1 white Soldier")
	b12Push(g, me.ID, "Plain Bear", "Creature — Bear", "", 2, 2)

	declareAttack(t, g, opp.ID, stormsinger)
	passAndSettleAttackSplit(t, g, me.ID)

	var copies []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller != me.ID || !c.IsToken() || !c.Tapped || c.AttackingTarget == uuid.Nil {
			continue
		}
		copies = append(copies, c.InstanceID)
		if c.Name != "Spirit" && c.Name != "Soldier" {
			t.Errorf("a tapped and attacking token named %q", c.Name)
		}
	}
	if len(copies) != 2 {
		t.Fatalf("%d tapped and attacking token copies, want 2 (one per creature token)", len(copies))
	}
	names := map[string]int{}
	for _, id := range copies {
		c, _ := battlefieldCardByID(g, id)
		names[c.Name]++
	}
	if names["Spirit"] != 1 || names["Soldier"] != 1 {
		t.Errorf("copies %v, want one Spirit and one Soldier", names)
	}

	// At the beginning of the next end step the copies are sacrificed; the
	// originals stay.
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	for _, id := range copies {
		if findBattlefieldCardForTest(g, id) != nil {
			t.Errorf("copy %s survived the end step", id)
		}
	}
	if got := b43TokensNamed(g, me.ID, "Spirit") + b43TokensNamed(g, me.ID, "Soldier"); got != 2 {
		t.Errorf("%d original tokens left, want 2", got)
	}
}

func TestSocB11StormsingerWithNoTokenDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	stormsinger := socB11StormsingerAttack(t, g, me, opp)
	declareAttack(t, g, opp.ID, stormsinger)
	passAndSettleAttackSplit(t, g, me.ID)
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() {
			t.Errorf("a token %q appeared with no token to copy", c.Name)
		}
	}
}

func TestSocB11StormsingerCopiesAreNeverDeclaredAsAttackers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	stormsinger := socB11StormsingerAttack(t, g, me, opp)
	socB11Token(t, g, me.ID, "3/2 red and white Spirit")
	declareAttack(t, g, opp.ID, stormsinger)
	passAndSettleAttackSplit(t, g, me.ID)
	// The Stormsinger's trigger was the only attack trigger: the copy did not
	// trigger it again, so exactly one copy exists.
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.IsToken() && c.Name == "Spirit" {
			n++
		}
	}
	if n != 2 {
		t.Errorf("%d Spirit tokens, want the original and one copy", n)
	}
}
