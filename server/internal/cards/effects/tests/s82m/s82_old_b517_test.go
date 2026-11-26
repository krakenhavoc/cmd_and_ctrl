package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b517_test.go — Shards of Alara, enters and dies triggers, Auras,
// mana and ramp, artifacts: S82 OLD batch 517 (#5588), each card driven through
// the engine.

const (
	b517Sculler    = "51b55057-8b14-4277-894b-4813b4f7eb39"
	b517GlazeFiend = "6860cde7-ee8e-4a4d-baed-9dc4a1220e36"
	b517Beetle     = "8ec6b0f1-8928-457c-adf0-c74a28edc86f"
	b517Specimens  = "e3b8d70b-17aa-4a31-a4f9-319907d088c2"
	b517Hunger     = "0e608f48-d3a6-422c-839d-f06bb8a32014"
	b517ComaVeil   = "c062bd12-34cf-4177-92d4-8ce91206989d"
	b517Banewasp   = "c1cb870a-8c1f-435c-b678-e456c4d2d627"
	b517Nurturer   = "3518e273-bd92-46bc-b361-da52c670dfe1"
	b517Godspeaker = "891e9a08-35ec-422f-87a8-1f3f5c705699"
	b517Steward    = "ca5d0a56-69c1-4bdb-95f0-af0dec3604a5"
	b517Orb        = "5794b58e-f3b6-4637-b288-283b3a72a718"
	b517Serpent    = "6410e2bb-5ae6-4190-8b4a-6b2d5c9dff8d"
)

func TestS82OldB517EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		b517Sculler: "Tidehollow Sculler", b517GlazeFiend: "Glaze Fiend", b517Beetle: "Blister Beetle",
		b517Specimens: "Gather Specimens", b517Hunger: "Savage Hunger", b517ComaVeil: "Coma Veil",
		b517Banewasp: "Banewasp Affliction", b517Nurturer: "Sunseed Nurturer", b517Godspeaker: "Sacellum Godspeaker",
		b517Steward: "Steward of Valeron", b517Orb: "Mindlock Orb", b517Serpent: "Steelclad Serpent",
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

func b517Layered(t *testing.T, g *game.Game, id uuid.UUID) game.Card {
	t.Helper()
	return effectstest.LayeredCard(t, g, id)
}

func TestS82OldB517TidehollowScullerTakesACardAndReturnsItWhenItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	for opp.Hand.Size() > 0 {
		_, _ = opp.Hand.PopTop()
	}
	land := effectstest.EoHand(opp, "Forest", "Basic Land — Forest")
	spell := effectstest.EoHand(opp, "Bolt", "Instant")
	sculler := effectstest.CastCatalogSpell(t, g, "Tidehollow Sculler", "Artifact Creature — Zombie", b517Sculler, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		effectstest.PickPlayer(t, g, me.ID, opp.ID)
	}
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.OpenVariant(t, g)
	if err := g.ResolvePendingChoice(pick.ID, me.ID, []uuid.UUID{land}); err == nil {
		t.Fatal("a land card was accepted")
	}
	if err := g.ResolvePendingChoice(pick.ID, me.ID, []uuid.UUID{spell}); err != nil {
		t.Fatalf("resolve the pick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if z := effectstest.ZoneOfCard(g, spell); z != game.ZoneExile {
		t.Fatalf("the chosen card is in %v, want exile", z)
	}
	effectstest.B27Kill(g, sculler)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Hand.Contains(spell) {
		t.Errorf("the card did not return to its owner's hand (it is in %v)", effectstest.ZoneOfCard(g, spell))
	}
}

func TestS82OldB517GlazeFiendGrowsWhenAnotherArtifactEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	fiend := effectstest.B12Push(g, me.ID, "Glaze Fiend", "Artifact Creature — Illusion", b517GlazeFiend, 0, 1)
	if !effectstest.HasKeywordOnBattlefield(t, g, fiend, "flying") {
		t.Error("no flying")
	}
	effectstest.CastCatalogSpell(t, g, "Trinket", "Artifact", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if c := b517Layered(t, g, fiend); c.CurrentPower() != 2 || c.CurrentToughness() != 3 {
		t.Fatalf("after an artifact entered: %d/%d, want 2/3", c.CurrentPower(), c.CurrentToughness())
	}
	// A non-artifact does not count, and the boost ends with the turn.
	effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if c := b517Layered(t, g, fiend); c.CurrentPower() != 2 {
		t.Errorf("a non-artifact entering changed the power to %d", c.CurrentPower())
	}
	effectstest.DcCleanup(g)
	if c := b517Layered(t, g, fiend); c.CurrentPower() != 0 || c.CurrentToughness() != 1 {
		t.Errorf("after cleanup: %d/%d, want 0/1", c.CurrentPower(), c.CurrentToughness())
	}
}

func TestS82OldB517GlazeFiendIgnoresItsOwnEntryAndAnOpponentsArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	fiend := effectstest.CastCatalogSpell(t, g, "Glaze Fiend", "Artifact Creature — Illusion", b517GlazeFiend, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if c := b517Layered(t, g, fiend); c.CurrentPower() != 0 {
		t.Errorf("its own entry boosted it to %d", c.CurrentPower())
	}
	effectstest.B12Push(g, opp.ID, "Their Rock", "Artifact", "", 0, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if c := b517Layered(t, g, fiend); c.CurrentPower() != 0 {
		t.Errorf("an opponent's artifact boosted it to %d", c.CurrentPower())
	}
	_ = me
}

func TestS82OldB517BlisterBeetleShrinksATargetCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Blister Beetle", "Creature — Insect", b517Beetle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickTargetRefs(t, g, g.Seats[g.Turn.ActiveSeat].ID, nil, effectstest.CardRefs(bear)...)
	effectstest.PassPriorityAroundTable(t, g)
	if c := b517Layered(t, g, bear); c.CurrentPower() != 1 || c.CurrentToughness() != 1 {
		t.Fatalf("target is %d/%d, want 1/1", c.CurrentPower(), c.CurrentToughness())
	}
	effectstest.DcCleanup(g)
	if c := b517Layered(t, g, bear); c.CurrentPower() != 2 || c.CurrentToughness() != 2 {
		t.Errorf("after cleanup %d/%d, want 2/2", c.CurrentPower(), c.CurrentToughness())
	}
}

func TestS82OldB517GatherSpecimensRedirectsAnOpponentsCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.CastCatalogSpell(t, g, "Gather Specimens", "Instant", b517Specimens, nil)
	effectstest.PassPriorityAroundTable(t, g)
	theirs := effectstest.EoHand(opp, "Their Bear", "Creature — Bear")
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: opp.ID}, game.ZoneRef{Kind: game.ZoneBattlefield}, theirs); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	if c, ok := g.LookupCardForEffect(theirs); !ok || c.Controller != me.ID {
		t.Errorf("their creature entered under %v, want the caster", c.Controller)
	}
	mine := effectstest.EoHand(me, "My Bear", "Creature — Bear")
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneBattlefield}, mine); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	if c, ok := g.LookupCardForEffect(mine); !ok || c.Controller != me.ID {
		t.Errorf("my creature entered under %v", c.Controller)
	}
	effectstest.DcCleanup(g)
	later := effectstest.EoHand(opp, "Later Bear", "Creature — Bear")
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: opp.ID}, game.ZoneRef{Kind: game.ZoneBattlefield}, later); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	if c, ok := g.LookupCardForEffect(later); !ok || c.Controller != opp.ID {
		t.Errorf("the replacement outlived the turn: entered under %v", c.Controller)
	}
}

func TestS82OldB517SavageHungerPumpsAndGrantsTrample(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	aura := effectstest.AuraCast(t, g, "Savage Hunger", b517Hunger, bear)
	if !g.Battlefield.Contains(aura) {
		t.Fatal("the Aura did not resolve")
	}
	if c := b517Layered(t, g, bear); c.CurrentPower() != 3 || c.CurrentToughness() != 2 {
		t.Errorf("host %d/%d, want 3/2", c.CurrentPower(), c.CurrentToughness())
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, bear, "trample") {
		t.Error("host has no trample")
	}
	spec, _ := effects.Lookup(b517Hunger)
	if len(spec.Activated) != 1 {
		t.Errorf("%d activated abilities, want cycling", len(spec.Activated))
	}
	effectstest.B27Kill(g, aura)
	g.RunStateChecksForTest()
	if c := b517Layered(t, g, bear); c.CurrentPower() != 2 || effectstest.HasKeywordOnBattlefield(t, g, bear, "trample") {
		t.Error("the bonus outlived the Aura")
	}
}

func TestS82OldB517ComaVeilKeepsAnArtifactOrCreatureTapped(t *testing.T) {
	spec, _ := effects.Lookup(b517ComaVeil)
	if len(spec.UntapStepRestrictions) != 1 {
		t.Fatalf("%d untap-step restrictions, want 1", len(spec.UntapStepRestrictions))
	}
	for _, tc := range []struct{ name, typeLine string }{{"Bear", "Creature — Bear"}, {"Rock", "Artifact"}} {
		g := effectstest.NewCatalogGame(t)
		opp := g.Seats[1]
		host := effectstest.B12Push(g, opp.ID, tc.name, tc.typeLine, "", 2, 2)
		effectstest.AuraCast(t, g, "Coma Veil", b517ComaVeil, host)
		effectstest.ArnTap(t, g, host)
		effectstest.AdvanceToUpkeepOf(t, g, 1)
		if !effectstest.ArnTapped(t, g, host) {
			t.Errorf("%s untapped in its controller's untap step", tc.name)
		}
	}
	// A land is not a legal host.
	g := effectstest.NewCatalogGame(t)
	land := effectstest.B12Permanent(g, g.Seats[1].ID, "Forest", "Basic Land — Forest")
	id := uuid.New()
	me := g.Seats[g.Turn.ActiveSeat]
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Coma Veil", TypeLine: "Enchantment — Aura", OracleID: b517ComaVeil, Owner: me.ID, Controller: me.ID})
	effectstest.ToMainForCost(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.CardRefs(land)}); err == nil {
		t.Error("Coma Veil was cast on a land")
	}
}

func TestS82OldB517BanewaspAfflictionDrainsTheDeadCreaturesControllerByToughness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	wall := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	effectstest.AuraCast(t, g, "Banewasp Affliction", b517Banewasp, wall)
	oppLife, myLife := opp.Life, me.Life
	effectstest.KillCreature(t, g, opp.ID, wall)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != oppLife-4 || me.Life != myLife {
		t.Errorf("life: opponent %d → %d, me %d → %d; want the opponent −4", oppLife, opp.Life, myLife, me.Life)
	}
}

func TestS82OldB517SunseedNurturerMayGainLifeWithAPowerFiveCreature(t *testing.T) {
	for _, tc := range []struct {
		name  string
		power int
		yes   bool
		want  int
	}{{"power 5, yes", 5, true, 2}, {"power 5, no", 5, false, 0}, {"power 4", 4, true, 0}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			effectstest.B12Push(g, me.ID, "Sunseed Nurturer", "Creature — Human Druid Wizard", b517Nurturer, 1, 1)
			effectstest.B12Creature(g, me.ID, "Big", "Creature — Giant", tc.power, tc.power)
			life := me.Life
			effectstest.AdvanceToEndStepOf(t, g, g.Turn.ActiveSeat)
			effectstest.PassPriorityAroundTable(t, g)
			if effectstest.B18TriggerPromptCount(g, me.ID) > 0 {
				effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, tc.yes)
				effectstest.PassPriorityAroundTable(t, g)
			}
			if me.Life != life+tc.want {
				t.Errorf("life %d → %d, want +%d", life, me.Life, tc.want)
			}
		})
	}
}

func TestS82OldB517SunseedNurturerTapsForColorless(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	n := effectstest.PushCatalogPermanent(g, me.ID, "Sunseed Nurturer", "Creature — Human Druid Wizard", b517Nurturer, false)
	effectstest.Unsick(g, n)
	if err := g.ActivateManaAbility(me.ID, n, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	got := 0
	for _, tok := range me.ManaPool {
		if tok.Color == "C" {
			got += tok.N()
		}
	}
	if got != 1 {
		t.Errorf("pool has %d {C}, want 1", got)
	}
}

func TestS82OldB517SacellumGodspeakerAddsGreenForEachBigCreatureRevealed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.EmptyHand(g, me)
	id := effectstest.PushCatalogPermanent(g, me.ID, "Sacellum Godspeaker", "Creature — Elf Druid", b517Godspeaker, false)
	effectstest.Unsick(g, id)
	big1 := effectstest.AvB1Hand(me, game.Card{Name: "Wurm", TypeLine: "Creature — Wurm", Power: 5, Toughness: 5})
	big2 := effectstest.AvB1Hand(me, game.Card{Name: "Hydra", TypeLine: "Creature — Hydra", Power: 7, Toughness: 7})
	small := effectstest.AvB1Hand(me, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 4, Toughness: 4})
	sorcery := effectstest.AvB1Hand(me, game.Card{Name: "Spell", TypeLine: "Sorcery"})
	for _, bad := range []uuid.UUID{small, sorcery} {
		if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{RevealIDs: []uuid.UUID{bad}}); err == nil {
			t.Fatal("revealed a card that is not a creature with power 5 or greater")
		}
	}
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{RevealIDs: []uuid.UUID{big1, big2}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	got := 0
	for _, tok := range me.ManaPool {
		if tok.Color == "G" {
			got += tok.N()
		}
	}
	if got != 2 || me.Hand.Size() != 4 {
		t.Errorf("pool has %d {G}, hand %d; want 2 and the four cards still in hand", got, me.Hand.Size())
	}
}

func TestS82OldB517SacellumGodspeakerRevealingNothingAddsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.PushCatalogPermanent(g, me.ID, "Sacellum Godspeaker", "Creature — Elf Druid", b517Godspeaker, false)
	effectstest.Unsick(g, id)
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	for _, tok := range me.ManaPool {
		if tok.N() > 0 && tok.Color == "G" {
			t.Errorf("a {G} was added for no cards")
		}
	}
}

func TestS82OldB517StewardOfValeronHasVigilanceAndTapsForGreen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.PushCatalogPermanent(g, me.ID, "Steward of Valeron", "Creature — Human Druid Knight", b517Steward, false)
	effectstest.Unsick(g, id)
	if !effectstest.HasKeywordOnBattlefield(t, g, id, "vigilance") {
		t.Error("no vigilance")
	}
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	got := 0
	for _, tok := range me.ManaPool {
		if tok.Color == "G" {
			got += tok.N()
		}
	}
	if got != 1 {
		t.Errorf("pool has %d {G}, want 1", got)
	}
}

func TestS82OldB517MindlockOrbStopsEveryPlayersSearch(t *testing.T) {
	for _, seat := range []int{0, 1} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.AdvanceToMain(t, g)
		effectstest.PushCatalogPermanent(g, me.ID, "Mindlock Orb", "Artifact", b517Orb, false)
		p := g.Seats[seat]
		land := game.NewCard("Mountain", p.ID)
		land.TypeLine = "Basic Land — Mountain"
		p.Library.PushTop(land)
		g.WithWriteLock(func() {
			_ = g.SearchLibraryForEffectWithOptions(p.ID, func(c game.Card) bool { return c.IsLand() }, game.ZoneHand, 1, false, true, false)
		})
		if effectstest.SearchPromptFor(g, p.ID) != nil {
			t.Errorf("seat %d was offered a search with Mindlock Orb in play", seat)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if p.Hand.Contains(land.InstanceID) {
			t.Errorf("seat %d found a land", seat)
		}
	}
}

func TestS82OldB517SteelcladSerpentNeedsAnotherArtifactToAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	serpent := effectstest.PushCatalogPermanent(g, me.ID, "Steelclad Serpent", "Artifact Creature — Serpent", b517Serpent, false)
	effectstest.Unsick(g, serpent)
	effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(serpent, opp.ID); err == nil {
		t.Fatal("the Serpent attacked with no other artifact under your control")
	}
	g2 := effectstest.NewCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	serpent2 := effectstest.PushCatalogPermanent(g2, me2.ID, "Steelclad Serpent", "Artifact Creature — Serpent", b517Serpent, false)
	effectstest.Unsick(g2, serpent2)
	effectstest.B12Permanent(g2, me2.ID, "My Rock", "Artifact")
	effectstest.AdvanceTo(t, g2, game.StepDeclareAttackers)
	if err := g2.DeclareAttacker(serpent2, opp2.ID); err != nil {
		t.Errorf("the Serpent could not attack with another artifact: %v", err)
	}
}
