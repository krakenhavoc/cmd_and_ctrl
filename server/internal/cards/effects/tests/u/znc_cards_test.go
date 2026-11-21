package testsu

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Zendikar Rising Commander (#5509): Anowon, the Ruin Thief; Enigma Thief;
// Geode Rager; Obuun, Mul Daya Ancestor; Trove Warden; Whispersteel Dagger.

const (
	zncAnowonOracle = "e7e69583-8326-417e-8a9a-6b917c8ba4b6"
	zncEnigmaOracle = "d39a1e12-bb6b-4e7a-96a4-5dab98084716"
	zncGeodeOracle  = "3969d1d6-0aed-4fa8-aa09-d9be50576fb4"
	zncObuunOracle  = "28fcf855-6372-4d6b-b3fd-76f85247f309"
	zncTroveOracle  = "9056eba4-612b-4e82-8689-fe098241b007"
	zncDaggerOracle = "ca59fbf4-c774-49d3-8630-108c053e01fb"
)

func TestZncCardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		zncAnowonOracle: "Anowon, the Ruin Thief",
		zncEnigmaOracle: "Enigma Thief",
		zncGeodeOracle:  "Geode Rager",
		zncObuunOracle:  "Obuun, Mul Daya Ancestor",
		zncTroveOracle:  "Trove Warden",
		zncDaggerOracle: "Whispersteel Dagger",
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

// zncSeedLibrary puts cards with the given type lines on top of p's
// library, the first one on top, and returns their IDs in that order.
func zncSeedLibrary(p *game.Player, typeLines ...string) []uuid.UUID {
	ids := make([]uuid.UUID, len(typeLines))
	for i := range typeLines {
		ids[i] = uuid.New()
	}
	for i := len(typeLines) - 1; i >= 0; i-- {
		p.Library.PushTop(game.Card{
			InstanceID: ids[i], Name: "Library Card", TypeLine: typeLines[i],
			Power: 1, Toughness: 1, Owner: p.ID, Controller: p.ID,
		})
	}
	return ids
}

// zncGraveyardCard puts a card into p's graveyard, known to every seat
// (CR 404.2).
func zncGraveyardCard(g *game.Game, p *game.Player, name, typeLine, cost string, colors []string) uuid.UUID {
	id := uuid.New()
	everyone := map[uuid.UUID]bool{}
	for _, s := range g.Seats {
		everyone[s.ID] = true
	}
	p.Graveyard.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, Colors: colors,
		Power: 2, Toughness: 2, Owner: p.ID, Controller: p.ID, KnownBy: everyone,
	})
	return id
}

// --- Anowon, the Ruin Thief -------------------------------------------

// Other Rogues you control get +1/+1; Anowon and an opponent's Rogue do
// not. The Rogues' combat damage to one player is one trigger that mills
// that many cards, a non-Rogue's damage at the same time does not count,
// and one creature card among them draws one card.
func TestZncAnowonPumpsOtherRoguesAndMillsForTheRoguesDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	anowon := effectstest.B43Catalog(g, me.ID, "Anowon, the Ruin Thief", "Legendary Creature — Vampire Rogue", zncAnowonOracle, 2, 4)
	rogueA := effectstest.B12Creature(g, me.ID, "Rogue A", "Creature — Human Rogue", 1, 1)
	rogueB := effectstest.B12Creature(g, me.ID, "Rogue B", "Creature — Human Rogue", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirRogue := effectstest.B12Creature(g, opp.ID, "Their Rogue", "Creature — Human Rogue", 1, 1)
	for _, tc := range []struct {
		id   uuid.UUID
		want int
		why  string
	}{
		{anowon, 2, "Anowon is not another Rogue"},
		{rogueA, 2, "another Rogue you control gets +1/+1"},
		{bear, 2, "a Bear is not a Rogue"},
		{theirRogue, 1, "an opponent's Rogue is not yours"},
	} {
		if got := effectstest.EffectivePower(t, g, tc.id); got != tc.want {
			t.Errorf("%s: power %d, want %d", tc.why, got, tc.want)
		}
	}

	lib := zncSeedLibrary(opp, "Sorcery", "Sorcery", "Sorcery", "Creature — Elf", "Creature — Elf", "Sorcery")
	hand := me.Hand.Size()
	effectstest.AttackWith(t, g, opp.ID, rogueA, rogueB, bear)
	effectstest.PassPriorityAroundTable(t, g)
	for i := 0; i < 4; i++ {
		if z := effectstest.E2Zone(g, lib[i]); z != game.ZoneGraveyard {
			t.Errorf("library card %d in %q, want the graveyard: the Rogues dealt 4", i, z)
		}
	}
	for i := 4; i < len(lib); i++ {
		if z := effectstest.E2Zone(g, lib[i]); z != game.ZoneLibrary {
			t.Errorf("library card %d in %q: the Bear's damage must not mill", i, z)
		}
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d: a creature card was milled", got, hand+1)
	}
}

// No creature card among the milled cards draws nothing.
func TestZncAnowonDrawsNothingWithoutACreatureMilled(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.B43Catalog(g, me.ID, "Anowon, the Ruin Thief", "Legendary Creature — Vampire Rogue", zncAnowonOracle, 2, 4)
	rogue := effectstest.B12Creature(g, me.ID, "Rogue", "Creature — Human Rogue", 1, 1)
	lib := zncSeedLibrary(opp, "Sorcery", "Sorcery", "Creature — Elf")
	hand := me.Hand.Size()
	effectstest.AttackWith(t, g, opp.ID, rogue)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.E2Zone(g, lib[0]) != game.ZoneGraveyard || effectstest.E2Zone(g, lib[1]) != game.ZoneGraveyard {
		t.Fatal("the Rogue's 2 damage did not mill two cards")
	}
	if effectstest.E2Zone(g, lib[2]) != game.ZoneLibrary {
		t.Error("a third card was milled")
	}
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand %d, want %d: no creature card was milled", got, hand)
	}
}

// --- Enigma Thief -----------------------------------------------------

// Prowl {3}{U} is offered only once a Sphinx or Rogue you controlled dealt
// combat damage to a player this turn; another creature's damage does not
// open it.
func TestZncEnigmaThiefProwlNeedsASphinxOrRogueToConnect(t *testing.T) {
	for _, tc := range []struct {
		name     string
		typeLine string
		attack   bool
		want     bool
	}{
		{"rogue connects", "Creature — Human Rogue", true, true},
		{"sphinx connects", "Creature — Sphinx", true, true},
		{"bear connects", "Creature — Bear", true, false},
		{"rogue stays home", "Creature — Human Rogue", false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
			dealer := effectstest.B12Creature(g, me.ID, "Dealer", tc.typeLine, 2, 2)
			if tc.attack {
				effectstest.AttackWith(t, g, opp.ID, dealer)
				effectstest.PassPriorityAroundTable(t, g)
			}
			effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
			thief := effectstest.HandCardFull(me, "Enigma Thief", "Creature — Sphinx Rogue", "{5}{U}{U}", zncEnigmaOracle, []string{"U"})
			if got := zncOffersProwl(g, me.ID, thief); got != tc.want {
				t.Errorf("prowl offered = %v, want %v", got, tc.want)
			}
			effectstest.FloatForTest(g, me, "CCCU")
			err := g.CastSpell(me.ID, thief, game.CastSpellParams{Strict: true, AlternativeCost: "prowl"})
			if tc.want && err != nil {
				t.Fatalf("prowl was refused: %v", err)
			}
			if !tc.want && err == nil {
				t.Fatal("prowl was accepted without a Sphinx or Rogue connecting")
			}
		})
	}
}

// The spell's creature types are read as it is in hand, so a type an
// effect gives it there counts (the ruling): under Leyline of
// Transformation naming Wizard, a Bear-Wizard connecting opens prowl.
func TestZncEnigmaThiefProwlReadsTheSpellsTypesInHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	// In hand before the Leyline, as a card drawn earlier would be.
	thief := effectstest.HandCardFull(me, "Enigma Thief", "Creature — Sphinx Rogue", "{5}{U}{U}", zncEnigmaOracle, []string{"U"})
	effectstest.CastCatalogSpell(t, g, "Leyline of Transformation", "Enchantment", effectstest.StdB21LeylineOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerCreatureType(t, g, me.ID, "Wizard")
	if !effectstest.B12Card(t, g, bear).HasSubtype("Wizard") {
		t.Fatal("the Leyline did not make the Bear a Wizard")
	}
	effectstest.AttackWith(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
	if !zncOffersProwl(g, me.ID, thief) {
		t.Fatal("prowl not offered: the Thief in hand is a Wizard too")
	}
	effectstest.FloatForTest(g, me, "CCCU")
	if err := g.CastSpell(me.ID, thief, game.CastSpellParams{Strict: true, AlternativeCost: "prowl"}); err != nil {
		t.Fatalf("prowl was refused: %v", err)
	}
}

// zncOffersProwl reports whether the view's offer list for the card in
// hand carries prowl.
func zncOffersProwl(g *game.Game, player, card uuid.UUID) bool {
	found := false
	g.ReadSnapshot(func() {
		c, ok := g.LookupCardForEffect(card)
		if !ok {
			return
		}
		perm := g.CastPermissionForLocked(player, c, game.ZoneHand)
		for _, o := range g.CastOffersForLocked(player, c, game.ZoneHand, perm) {
			if o != nil && o.Key == "prowl" {
				found = true
			}
		}
	})
	return found
}

// For each opponent, up to one nonland permanent that player controls
// returns to its owner's hand, all together; a land and an unpicked
// opponent's permanent stay.
func TestZncEnigmaThiefReturnsUpToOnePermanentPerOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	opp1 := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	opp2 := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	opp3 := g.Seats[(g.Turn.ActiveSeat+3)%len(g.Seats)]
	rock := effectstest.B12Permanent(g, opp1.ID, "Their Rock", "Artifact")
	land := effectstest.B12Permanent(g, opp1.ID, "Their Land", "Land")
	bear := effectstest.B12Creature(g, opp2.ID, "Their Bear", "Creature — Bear", 2, 2)
	kept := effectstest.B12Creature(g, opp3.ID, "Kept Bear", "Creature — Bear", 2, 2)
	effectstest.CastHoldCreature(t, g, "Enigma Thief", "Creature — Sphinx Rogue", zncEnigmaOracle, 5, 5)
	pick := effectstest.PickTargetPrompt(t, g)
	if effectstest.HasID(pick.PickTargetCards, land) {
		t.Error("a land was offered as a nonland permanent")
	}
	effectstest.AnswerPickTarget(t, g, rock)
	effectstest.AnswerPickTarget(t, g, bear)
	zncSkipPicks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if !opp1.Hand.Contains(rock) || !opp2.Hand.Contains(bear) {
		t.Fatal("the picks did not return to their owners' hands")
	}
	if !g.Battlefield.Contains(land) || !g.Battlefield.Contains(kept) {
		t.Error("an unpicked permanent left the battlefield")
	}
}

// zncSkipPicks declines every open "up to one" pick_target prompt.
func zncSkipPicks(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 4; i++ {
		var ch *game.PendingChoice
		for _, c := range g.PendingChoices {
			if c.Kind == game.PendingChoicePickTarget {
				ch = c
				break
			}
		}
		if ch == nil {
			return
		}
		if err := g.ResolvePickTargets(ch.ID, ch.Chooser, nil); err != nil {
			t.Fatalf("skip pick_target: %v", err)
		}
	}
}

// --- Geode Rager ------------------------------------------------------

// Landfall goads each creature the target player controls, and no one
// else's.
func TestZncGeodeRagerGoadsEachCreatureTargetPlayerControls(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	other := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	effectstest.B43Catalog(g, me.ID, "Geode Rager", "Creature — Elemental", zncGeodeOracle, 4, 3)
	a := effectstest.B12Creature(g, opp.ID, "Their A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "Their B", "Creature — Bear", 2, 2)
	c := effectstest.B12Creature(g, other.ID, "Other C", "Creature — Bear", 2, 2)
	effectstest.PlayLandFromHand(t, g, "Mountain", "")
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if got := effectstest.B33Goaded(t, g, id); got != me.ID {
			t.Errorf("the target player's creature is goaded by %s, want %s", got, me.ID)
		}
	}
	if got := effectstest.B33Goaded(t, g, c); got != uuid.Nil {
		t.Errorf("another player's creature was goaded by %s", got)
	}
}

// --- Obuun, Mul Daya Ancestor -----------------------------------------

// Landfall puts a +1/+1 counter on the target creature. At the beginning
// of combat the chosen land becomes an X/X Elemental creature with
// trample and haste, X Obuun's power as the trigger resolves and locked
// then; it is still a land, and an opponent's land is not a legal pick.
func TestZncObuunCountersOnLandfallAndAnimatesALandAtCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	obuun := effectstest.B43Catalog(g, me.ID, "Obuun, Mul Daya Ancestor", "Legendary Creature — Elf Spirit", zncObuunOracle, 3, 3)
	theirLand := effectstest.B12Permanent(g, opp.ID, "Their Forest", "Basic Land — Forest")

	land := effectstest.PlayLandFromHand(t, g, "Forest", "")
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, obuun)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, obuun, "+1/+1"); got != 1 {
		t.Fatalf("landfall: %d counters on Obuun, want 1", got)
	}

	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.B04WaitForPick(t, g, me.ID)
	if pick := effectstest.LatestPickTarget(g, me.ID); effectstest.HasID(pick.PickTargetCards, theirLand) {
		t.Error("an opponent's land was offered as a land you control")
	}
	effectstest.PickCard(t, g, me.ID, land)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.B12Card(t, g, land)
	if !c.IsCreature() || !c.IsLand() || !c.HasSubtype("Elemental") {
		t.Fatalf("the Forest is not a land Elemental creature: %q", c.TypeLine)
	}
	if got := effectstest.EffectivePower(t, g, land); got != 4 {
		t.Errorf("the land's power %d, want Obuun's 4", got)
	}
	abilities := effectstest.EffectiveAbilities(t, g, land)
	if !effects.EotHasAbility(abilities, "trample") || !effects.EotHasAbility(abilities, "haste") {
		t.Errorf("the land lacks trample or haste: %v", abilities)
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(obuun, "+1/+1", 2) })
	if got := effectstest.EffectivePower(t, g, land); got != 4 {
		t.Errorf("the land's power moved to %d with Obuun's: X is locked as the trigger resolves", got)
	}
}

// --- Trove Warden -----------------------------------------------------

// Landfall exiles a permanent card with mana value 3 or less from your
// graveyard; when the Warden dies, the cards exiled with it return under
// their owner's control.
func TestZncTroveWardenExilesSmallPermanentsAndReturnsThemWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	warden := effectstest.B43Catalog(g, me.ID, "Trove Warden", "Creature — Cat Beast", zncTroveOracle, 3, 4)
	small := zncGraveyardCard(g, me, "Small Rock", "Artifact", "{2}", nil)
	big := zncGraveyardCard(g, me, "Big Rock", "Artifact", "{4}", nil)
	sorcery := zncGraveyardCard(g, me, "Old Sorcery", "Sorcery", "{1}", nil)
	theirs := zncGraveyardCard(g, opp, "Their Rock", "Artifact", "{1}", nil)

	effectstest.PlayLandFromHand(t, g, "Plains", "")
	effectstest.B04WaitForPick(t, g, me.ID)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if !effectstest.HasID(pick.PickTargetCards, small) {
		t.Fatal("a mana value 2 artifact card in your graveyard is not a legal target")
	}
	for _, id := range []uuid.UUID{big, sorcery, theirs} {
		if effectstest.HasID(pick.PickTargetCards, id) {
			t.Errorf("card %s is offered: too big, not a permanent card, or not in your graveyard", id)
		}
	}
	effectstest.PickCard(t, g, me.ID, small)
	effectstest.PassPriorityAroundTable(t, g)
	if z := effectstest.E2Zone(g, small); z != game.ZoneExile {
		t.Fatalf("the target is in %q, want exile", z)
	}

	effectstest.B18Kill(t, g, warden)
	// A card that changes zones is a new object (CR 400.7): find it by name.
	var back []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Small Rock" {
			back = append(back, c)
		}
	}
	if len(back) != 1 {
		t.Fatalf("%d Small Rocks on the battlefield, want the one exiled with the Warden", len(back))
	}
	if back[0].Controller != me.ID || back[0].Owner != me.ID {
		t.Errorf("the returned card is controlled by %s, want its owner %s", back[0].Controller, me.ID)
	}
	if z := effectstest.E2Zone(g, big); z != game.ZoneGraveyard {
		t.Errorf("a card never exiled moved to %q", z)
	}
}

// --- Whispersteel Dagger ----------------------------------------------

// Equipped creature gets +2/+0. After it connects, its controller may
// cast ONE creature spell from that player's graveyard this turn, paying
// with mana of any colour; not a noncreature card, and not from another
// player's graveyard.
func TestZncWhispersteelDaggerOpensOneCreatureSpellFromTheDamagedGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	other := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	dagger := effectstest.B43Catalog(g, me.ID, "Whispersteel Dagger", "Artifact — Equipment", zncDaggerOracle, 0, 0)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.FloatForTest(g, me, "CCC")
	effectstest.EquipTo(t, g, me.ID, dagger, bear)
	if got := effectstest.EffectivePower(t, g, bear); got != 4 {
		t.Fatalf("equipped power %d, want 4", got)
	}

	theirBear := zncGraveyardCard(g, opp, "Their Bear", "Creature — Bear", "{1}{G}", []string{"G"})
	theirElk := zncGraveyardCard(g, opp, "Their Elk", "Creature — Elk", "{1}{G}", []string{"G"})
	theirSorcery := zncGraveyardCard(g, opp, "Their Sorcery", "Sorcery", "{G}", []string{"G"})
	otherBear := zncGraveyardCard(g, other, "Other Bear", "Creature — Bear", "{1}{G}", []string{"G"})
	if effectstest.GraveyardPermission(g, me.ID, theirBear) != nil {
		t.Fatal("the graveyard is open before the Dagger's creature connected")
	}

	effectstest.AttackWith(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.GraveyardPermission(g, me.ID, theirBear) == nil || effectstest.GraveyardPermission(g, me.ID, theirElk) == nil {
		t.Fatal("the damaged player's creature cards are not open to the Dagger's controller")
	}
	if effectstest.GraveyardPermission(g, me.ID, theirSorcery) != nil {
		t.Error("a noncreature card was opened")
	}
	if effectstest.GraveyardPermission(g, me.ID, otherBear) != nil {
		t.Error("another player's graveyard was opened")
	}

	effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
	effectstest.FloatForTest(g, me, "UU")
	if err := g.CastSpell(me.ID, theirBear, game.CastSpellParams{FromZone: "graveyard", Strict: true}); err != nil {
		t.Fatalf("cast their Bear with blue mana: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.B12Card(t, g, theirBear); !g.Battlefield.Contains(theirBear) || c.Controller != me.ID {
		t.Fatal("the Bear is not on the battlefield under the Dagger's controller")
	}
	if effectstest.GraveyardPermission(g, me.ID, theirElk) != nil {
		t.Error("a second creature spell is open: the Dagger opens one")
	}
}
