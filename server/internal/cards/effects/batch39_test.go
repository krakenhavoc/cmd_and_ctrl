package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch39_test.go — card-level coverage for the card-coverage roadmap's
// batch 39 (#402, `edhrec_rank` 4053–4153). One test per observable
// behaviour, driven through a real cast, activation, attack or combat
// damage step rather than by calling primitives.

const (
	b39GlacialFloodplainOracle  = "5d3563dd-a2c1-463c-a0ed-5ac22388bdbe"
	b39SnowfieldSinkholeOracle  = "749c2c8e-9588-4e83-b07f-3c37eb63338b"
	b39DragonSniperOracle       = "913dbce2-6805-453f-a9c3-64f1cc658627"
	b39ZoZuOracle               = "0e412b68-9179-4094-960a-95692428855b"
	b39MaskOfAvacynOracle       = "ab66f8a8-eb3d-4c2d-95e9-26a53c66b237"
	b39MaulOfTheSkyclavesOracle = "088a1f56-198c-41a1-b244-ec1c7a771041"
	b39RoguesGlovesOracle       = "308d7868-49f9-47a3-a7b1-4b0332d610f1"
	b39FeatherOfFlightOracle    = "c61b05b8-00f1-4371-9856-c65852b4ca02"
	b39KorCartographerOracle    = "a5f537f6-3254-45a3-a4e7-d4d4fd7a972d"
	b39BloodlineNecroOracle     = "f2a54e85-dd8b-462c-a811-c9553dc349a5"
	b39SpawningKrakenOracle     = "f0357833-80b5-40ba-874a-bd22ea6e4e46"
	b39NecropolisRegentOracle   = "33c3a6be-8cae-4de8-9d9b-43216d727e93"
	b39MerfolkSovereignOracle   = "16aba9af-7a12-4c3d-87eb-06278921cb7c"
	b39LuxCannonOracle          = "ebdf26a4-77ee-4635-8d52-926bdca623f7"
	b39NestedShamblerOracle     = "bfa882ee-de18-4ef8-8957-3e04ed6e3c1e"
	b39ThrillingDiscoveryOracle = "c5f5a234-c751-4976-a3e1-fbd52b3255c9"
	b39JanJansenOracle          = "cb491d8a-2e9f-46fe-9590-44c3b4a25f1b"
	b39BezaOracle               = "020de6d7-f5a2-4036-ad25-451e5977b4d4"
	b39ShalaiAndHallarOracle    = "e7604cd9-d00d-4957-82c9-46a7cdb88209"
	b39AutogeneratorOracle      = "6832f9e2-294a-44a2-8af9-eb16ccfdfc36"
	b39AnaraOracle              = "a59ff932-f758-476e-ba31-0623bd748231"
	b39AirbendingLessonOracle   = "e101b744-9175-4aa5-bf80-d6944359e538"
	b39InquisitorsFlailOracle   = "a89ae357-b5aa-4256-beb3-a2e5e7f43200"
	b39DawnOfANewAgeOracle      = "37df5ade-6a21-4b0c-89f8-e2e917589a8d"
	b39MarketbackWalkerOracle   = "0405e0a9-6d02-4691-bdb8-59c72b824dab"
	b39AjaniOracle              = "f5d9be71-91d0-4166-ba58-cbbf5d490c40"
	b39ObNixilisOracle          = "88cd12e5-ca89-4176-96fd-320fe2ccd087"
	b39SorinMarkovOracle        = "c151e8b7-4b28-4f03-8a81-9bf623f893d7"
	b39CatharticPyreOracle      = "f7e55117-69fb-412e-b16e-d92e0f95f629"
	b39BanefireOracle           = "5eff8a06-e0d6-435a-a7c0-db9f9d98636a"
	b39ErebosOracle             = "a5690dc9-f12e-4ecc-b528-f2f83c79c3c8"
)

// --- registration --------------------------------------------------

// Every card in the batch, pinned by oracle ID → name. The two snow
// duals are rows in the typed-tapland cycle table, so a transposed row
// is invisible until somebody plays that exact land.
func TestBatch39CardsAreRegistered(t *testing.T) {
	want := map[string]string{
		b39GlacialFloodplainOracle:  "Glacial Floodplain",
		b39SnowfieldSinkholeOracle:  "Snowfield Sinkhole",
		b39DragonSniperOracle:       "Dragon Sniper",
		b39ZoZuOracle:               "Zo-Zu the Punisher",
		b39MaskOfAvacynOracle:       "Mask of Avacyn",
		b39MaulOfTheSkyclavesOracle: "Maul of the Skyclaves",
		b39RoguesGlovesOracle:       "Rogue's Gloves",
		b39FeatherOfFlightOracle:    "Feather of Flight",
		b39KorCartographerOracle:    "Kor Cartographer",
		b39BloodlineNecroOracle:     "Bloodline Necromancer",
		b39SpawningKrakenOracle:     "Spawning Kraken",
		b39NecropolisRegentOracle:   "Necropolis Regent",
		b39MerfolkSovereignOracle:   "Merfolk Sovereign",
		b39LuxCannonOracle:          "Lux Cannon",
		b39NestedShamblerOracle:     "Nested Shambler",
		b39ThrillingDiscoveryOracle: "Thrilling Discovery",
		b39JanJansenOracle:          "Jan Jansen, Chaos Crafter",
		b39BezaOracle:               "Beza, the Bounding Spring",
		b39ShalaiAndHallarOracle:    "Shalai and Hallar",
		b39AutogeneratorOracle:      "Empowered Autogenerator",
		b39AnaraOracle:              "Anara, Wolvid Familiar",
		b39AirbendingLessonOracle:   "Airbending Lesson",
		b39InquisitorsFlailOracle:   "Inquisitor's Flail",
		b39DawnOfANewAgeOracle:      "Dawn of a New Age",
		b39MarketbackWalkerOracle:   "Marketback Walker",
		b39AjaniOracle:              "Ajani, the Greathearted",
		b39ObNixilisOracle:          "Ob Nixilis, the Hate-Twisted",
		b39SorinMarkovOracle:        "Sorin Markov",
		b39CatharticPyreOracle:      "Cathartic Pyre",
		b39BanefireOracle:           "Banefire",
		b39ErebosOracle:             "Erebos, Bleak-Hearted",
	}
	if len(want) != 31 {
		t.Fatalf("the batch registers 31 cards, the table lists %d", len(want))
	}
	for oracle, name := range want {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name {
			t.Errorf("oracle %s registered as %q, want %q", oracle, spec.Name, name)
		}
	}
}

// --- helpers -------------------------------------------------------

// b39Creature seeds a creature under owner, free of summoning
// sickness, with a real timestamp so the layer cache sees it.
func b39Creature(g *game.Game, owner uuid.UUID, name, typeLine string, power, toughness int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

// b39CurrentPower is the combat-relevant power: the post-layer value
// PLUS the +1/+1 counters, which Effective() deliberately leaves out
// (see game.Card.CurrentPower). Every counters card in this batch is
// asserted through it.
func b39CurrentPower(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	out, seen := 0, false
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID != id {
				continue
			}
			seen, out = true, c.CurrentPower()
			return
		}
	})
	if !seen {
		t.Fatalf("card %s not on battlefield", id)
	}
	return out
}

// b39NextTurnOf advances a whole turn cycle back to `seat`'s precombat
// main phase — a planeswalker's loyalty abilities are once per turn,
// so the second activation needs a new turn and not merely a new step.
func b39NextTurnOf(t *testing.T, g *game.Game, seat int) {
	t.Helper()
	for g.Turn.ActiveSeat == seat {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	advanceToMainOf(t, g, seat)
}

// b39WaitForTriggerPrompt passes priority until `chooser` has a "you
// may" trigger prompt waiting. The entry trigger of a creature spell
// is harvested only after the spell resolves, so the answer cannot be
// given before the priority passes.
func b39WaitForTriggerPrompt(t *testing.T, g *game.Game, chooser uuid.UUID) {
	t.Helper()
	for i := 0; i < 8 && b06TriggerPromptsFor(g, chooser) == 0; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if b06TriggerPromptsFor(g, chooser) == 0 {
		t.Fatalf("no trigger prompt for %s", chooser)
	}
}

// b39IsCreature reports whether the post-layer view of a permanent is
// a creature — Erebos's devotion clause is a layer 4 type change.
func b39IsCreature(t *testing.T, g *game.Game, id uuid.UUID) bool {
	t.Helper()
	out, seen := false, false
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID != id {
				continue
			}
			seen, out = true, c.IsCreature()
			return
		}
	})
	if !seen {
		t.Fatalf("card %s not on battlefield", id)
	}
	return out
}

// b39PlayerRefs is the player target-ref form the activate and cast
// paths take.
func b39PlayerRefs(ids ...uuid.UUID) []game.TargetRef {
	out := make([]game.TargetRef, 0, len(ids))
	for _, id := range ids {
		out = append(out, game.TargetRef{Kind: game.TargetPlayer, ID: id})
	}
	return out
}

// b39CountNamed counts battlefield permanents `controller` controls
// with the given name.
func b39CountNamed(g *game.Game, controller uuid.UUID, name string) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.Name == name {
			n++
		}
	}
	return n
}

// --- lands ---------------------------------------------------------

func TestB39SnowDualsEnterTappedAndTapForTheirColours(t *testing.T) {
	for _, row := range []struct{ name, typeLine, oracle, a, b string }{
		{"Glacial Floodplain", "Snow Land — Plains Island", b39GlacialFloodplainOracle, "W", "U"},
		{"Snowfield Sinkhole", "Snow Land — Plains Swamp", b39SnowfieldSinkholeOracle, "W", "B"},
	} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		land := b12PlayFromHand(t, g, row.name, row.typeLine, row.oracle, game.CastSpellParams{})
		if !b16Tapped(t, g, land) {
			t.Fatalf("%s: the land enters tapped", row.name)
		}
		g.WithWriteLock(func() { _ = g.UntapTargetForEffect(land) })
		b28TapForMana(t, g, me.ID, land, row.a)
		if got := poolColors(me); len(got) != 1 || got[0] != row.a {
			t.Errorf("%s: tapped for {%s}: pool %v", row.name, row.a, got)
		}
		g.WithWriteLock(func() { _ = g.UntapTargetForEffect(land) })
		b28TapForMana(t, g, me.ID, land, row.b)
		if got := poolColors(me); len(got) != 2 || got[1] != row.b {
			t.Errorf("%s: also taps for {%s}: pool %v", row.name, row.b, got)
		}
	}
}

// --- vanilla keywords ----------------------------------------------

func TestB39DragonSniperHasItsThreeKeywords(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	sniper := pushCatalogPermanent(g, me.ID, "Dragon Sniper", "Creature — Human Archer", b39DragonSniperOracle, false)
	for _, kw := range []string{"reach", "vigilance", "deathtouch"} {
		if !hasEffectiveKeyword(t, g, sniper, kw) {
			t.Errorf("Dragon Sniper should have %s", kw)
		}
	}
	if hasEffectiveKeyword(t, g, sniper, "flying") {
		t.Error("reach is not flying")
	}
}

// --- Zo-Zu the Punisher --------------------------------------------

// Every land, whoever plays it, including Zo-Zu's own controller.
func TestB39ZoZuBurnsWhoeverPlaysALand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Zo-Zu the Punisher", "Legendary Creature — Goblin Warrior", b39ZoZuOracle, false)

	mine := me.Life
	b12PlayFromHand(t, g, "Wastes", "Basic Land", "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if me.Life != mine-2 {
		t.Errorf("your own land drop burns you for 2: %d → %d", mine, me.Life)
	}

	// A land entering under an opponent's control by any route —
	// here put straight onto the battlefield — burns THEM, not the
	// active player.
	theirs := opp.Life
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{
			InstanceID: uuid.New(), Name: "Their Island", TypeLine: "Basic Land — Island",
			Owner: opp.ID, Controller: opp.ID,
		})
		g.EmitEvent(game.Event{Kind: game.EventETB, CardID: g.Battlefield.Cards[len(g.Battlefield.Cards)-1].InstanceID, Actor: opp.ID})
	})
	passPriorityAroundTable(t, g)
	if opp.Life != theirs-2 {
		t.Errorf("the land's controller takes the 2, not the active player: %d → %d", theirs, opp.Life)
	}
	if me.Life != mine-2 {
		t.Errorf("you took nothing for their land: life %d", me.Life)
	}
}

// --- Equipment -----------------------------------------------------

func TestB39MaskOfAvacynPumpsAndGrantsHexproof(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	mask := seedEquipment(g, me.ID, "Mask of Avacyn", b39MaskOfAvacynOracle)
	equipTo(t, g, me.ID, mask, bear)

	if got := effectivePower(t, g, bear); got != 3 {
		t.Errorf("equipped power %d, want 3", got)
	}
	if got := effectiveToughness(t, g, bear); got != 4 {
		t.Errorf("equipped toughness %d, want 4", got)
	}
	if !hasEffectiveKeyword(t, g, bear, "hexproof") {
		t.Error("the equipped creature has hexproof")
	}
}

// The entry trigger attaches the Maul to a creature YOU control, and
// it is a real targeted trigger — not an as-enters hook.
func TestB39MaulOfTheSkyclavesAttachesItselfOnEntry(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := seedBear(g, me.ID)
	theirs := b39Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	maul := castCatalogSpell(t, g, "Maul of the Skyclaves", equipTypeLine, b39MaulOfTheSkyclavesOracle, nil)
	// The trigger goes on the stack as the Equipment enters and names
	// its target then — an opponent's creature is never offered.
	b04WaitForPick(t, g, me.ID)
	if p := latestPickTarget(g, me.ID); hasID(p.PickTargetCards, theirs) || !hasID(p.PickTargetCards, bear) {
		t.Fatalf("the entry trigger targets a creature you control: %+v", p)
	}
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)

	if host := attachmentHostOf(t, g, maul); host.Kind != game.TargetCard || host.ID != bear {
		t.Fatalf("the Maul attached itself to the Bear: %+v", host)
	}
	if got := effectivePower(t, g, bear); got != 4 {
		t.Errorf("equipped power %d, want 4", got)
	}
	for _, kw := range []string{"flying", "first strike"} {
		if !hasEffectiveKeyword(t, g, bear, kw) {
			t.Errorf("the equipped creature has %s", kw)
		}
	}
}

// Combat damage only, and the "you may" is a real prompt.
func TestB39RoguesGlovesDrawOnCombatDamageOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	gloves := seedEquipment(g, me.ID, "Rogue's Gloves", b39RoguesGlovesOracle)
	equipTo(t, g, me.ID, gloves, bear)

	hand := me.Hand.Size()
	b27Damage(g, bear, opp.ID, 2) // noncombat
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("noncombat damage draws nothing: hand %d → %d", hand, me.Hand.Size())
	}

	dealCombatDamageToPlayer(g, bear, opp.ID, 2)
	b06AnswerAllTriggerPrompts(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("combat damage to a player draws one: hand %d → %d", hand, me.Hand.Size())
	}
}

// --- Auras ---------------------------------------------------------

func TestB39FeatherOfFlightCantripsAndGrantsFlying(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := seedBear(g, me.ID)
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Feather of Flight", "Enchantment — Aura", b39FeatherOfFlightOracle, cardRefs(bear))
	passPriorityAroundTable(t, g)

	// The helper seeds the Aura into hand and casts it (net zero
	// against the pre-cast count); the entry trigger then draws one.
	if me.Hand.Size() != hand+1 {
		t.Errorf("the Aura replaces itself: hand %d → %d", hand, me.Hand.Size())
	}
	if got := effectivePower(t, g, bear); got != 3 {
		t.Errorf("enchanted power %d, want 3", got)
	}
	if got := effectiveToughness(t, g, bear); got != 2 {
		t.Errorf("Feather of Flight is +1/+0: toughness %d, want 2", got)
	}
	if !hasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("the enchanted creature has flying")
	}
	spec, _ := Lookup(b39FeatherOfFlightOracle)
	if len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != "flash" {
		t.Errorf("flash is printed on the Aura: %v", spec.PrintedKeywords)
	}
}

// --- enters triggers ------------------------------------------------

// "A Plains card" is the subtype, not the basic — and it lands tapped.
func TestB39KorCartographerTakesAnyPlainsCardTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ids := seedSearchLibrary(me,
		game.Card{Name: "Hallowed Fountain", TypeLine: "Land — Plains Island"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
	)
	castCatalogSpell(t, g, "Kor Cartographer", "Creature — Kor Scout", b39KorCartographerOracle, nil)
	passPriorityAroundTable(t, g)

	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("\"you may search\": the chooser asks")
	}
	if searchOptionNamed(g, c, "Hallowed Fountain") == uuid.Nil {
		t.Error("a nonbasic Plains is a Plains card")
	}
	if searchOptionNamed(g, c, "Forest") != uuid.Nil {
		t.Error("a Forest is not a Plains card")
	}
	answerSearchNamed(t, g, me.ID, "Hallowed Fountain")
	if !g.Battlefield.Contains(ids[0]) {
		t.Fatal("the pick goes onto the battlefield")
	}
	if !b16Tapped(t, g, ids[0]) {
		t.Error("the fetched land enters tapped")
	}
}

func TestB39BloodlineNecromancerReanimatesOnlyAVampireOrWizard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	vampire := pushGraveyardPermanent(me, "Vampire Nighthawk", "Creature — Vampire Shaman", "{1}{B}{B}")
	wizard := pushGraveyardPermanent(me, "Merfolk Looter", "Creature — Merfolk Wizard", "{1}{U}")
	bear := pushGraveyardPermanent(me, "Bear", "Creature — Bear", "{1}{G}")

	castCatalogSpell(t, g, "Bloodline Necromancer", "Creature — Vampire Wizard", b39BloodlineNecroOracle, nil)
	// The "you may" is asked as the trigger is harvested, which is
	// after the creature spell has resolved — so the priority passes
	// come first and the answer second.
	b39WaitForTriggerPrompt(t, g, me.ID)
	b06AnswerAllTriggerPrompts(t, g, me.ID, true)
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if hasID(p.PickTargetCards, bear) {
		t.Error("a Bear is neither a Vampire nor a Wizard")
	}
	if !hasID(p.PickTargetCards, vampire) || !hasID(p.PickTargetCards, wizard) {
		t.Error("either subtype on its own qualifies")
	}
	pickCard(t, g, me.ID, vampire)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(vampire) {
		t.Error("it comes back to the battlefield, not to hand")
	}
}

// Four independent comparisons, each against any one opponent.
func TestB39BezaPaysOnlyForTheAxesAnOpponentLeadsOn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]

	// One opponent leads on lands; a different one leads on creatures.
	// Beza is on the battlefield when her own trigger resolves, so the
	// creature comparison is against a board of one — two of theirs is
	// what beats it.
	for i := 0; i < 3; i++ {
		b12Permanent(g, opp.ID, "Wastes", "Basic Land")
	}
	b39Creature(g, other.ID, "Their Bear", "Creature — Bear", 2, 2)
	b39Creature(g, other.ID, "Their Second Bear", "Creature — Bear", 2, 2)
	// Nobody leads on life or on cards in hand.
	opp.Life, other.Life = me.Life-1, me.Life-1
	for opp.Hand.Size() > 0 {
		opp.Hand.PopTop()
	}
	for other.Hand.Size() > 0 {
		other.Hand.PopTop()
	}

	life, hand := me.Life, me.Hand.Size()
	treasures := b39CountNamed(g, me.ID, "Treasure")
	castCatalogSpell(t, g, "Beza, the Bounding Spring", "Legendary Creature — Elemental Elk", b39BezaOracle, nil)
	passPriorityAroundTable(t, g)

	if got := b39CountNamed(g, me.ID, "Treasure") - treasures; got != 1 {
		t.Errorf("an opponent leads on lands: %d Treasures, want 1", got)
	}
	if got := b39CountNamed(g, me.ID, "Fish"); got != 2 {
		t.Errorf("an opponent leads on creatures: %d Fish, want 2", got)
	}
	if me.Life != life {
		t.Errorf("nobody leads on life: %d → %d, want no change", life, me.Life)
	}
	// The helper seeded Beza into hand and cast her: net zero, and no
	// draw, because nobody leads on cards in hand.
	if me.Hand.Size() != hand {
		t.Errorf("nobody leads on cards in hand: hand %d → %d, want %d", hand, me.Hand.Size(), hand)
	}
}

// --- combat-damage triggers ----------------------------------------

// The Kraken counts itself, one trigger per creature that connects,
// and a non-sea-monster makes nothing.
func TestB39SpawningKrakenMakesAKrakenPerSeaMonsterThatConnects(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	kraken := pushCatalogPermanent(g, me.ID, "Spawning Kraken", "Creature — Kraken", b39SpawningKrakenOracle, false)
	serpent := b39Creature(g, me.ID, "Serpent", "Creature — Serpent", 3, 3)
	bear := seedBear(g, me.ID)

	dealCombatDamageToPlayer(g, bear, opp.ID, 2)
	passPriorityAroundTable(t, g)
	if n := b39CountNamed(g, me.ID, "Kraken"); n != 0 {
		t.Errorf("a Bear is none of the four types: %d tokens", n)
	}

	dealCombatDamageToPlayer(g, kraken, opp.ID, 6)
	dealCombatDamageToPlayer(g, serpent, opp.ID, 3)
	passPriorityAroundTable(t, g)
	if n := b39CountNamed(g, me.ID, "Kraken"); n != 2 {
		t.Errorf("the Kraken itself and a Serpent each connect: %d tokens, want 2", n)
	}
}

// "That many" is the damage dealt, and it lands on the creature that
// dealt it — not on the Regent.
func TestB39NecropolisRegentGrowsWhicheverCreatureConnected(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	regent := pushCatalogPermanent(g, me.ID, "Necropolis Regent", "Creature — Vampire", b39NecropolisRegentOracle, false)
	bear := seedBear(g, me.ID)

	dealCombatDamageToPlayer(g, bear, opp.ID, 3)
	passPriorityAroundTable(t, g)
	if got := counterOn(g, bear, game.CounterPlusOne); got != 3 {
		t.Errorf("the Bear grew by the damage it dealt: %d counters, want 3", got)
	}
	if got := counterOn(g, regent, game.CounterPlusOne); got != 0 {
		t.Errorf("the counters go on the creature that connected, not the Regent: %d", got)
	}
}

// --- lords and activated abilities ---------------------------------

func TestB39MerfolkSovereignPumpsOtherMerfolkAndUnblocksAny(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sovereign := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Merfolk Sovereign", TypeLine: "Creature — Merfolk Noble",
		OracleID: b39MerfolkSovereignOracle, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	mine := b39Creature(g, me.ID, "Lullmage Mentor", "Creature — Merfolk Wizard", 2, 2)
	theirs := b39Creature(g, opp.ID, "Their Merfolk", "Creature — Merfolk", 2, 2)
	bear := seedBear(g, me.ID)

	if got := effectivePower(t, g, mine); got != 3 {
		t.Errorf("your other Merfolk gets +1/+1: power %d, want 3", got)
	}
	if got := effectivePower(t, g, sovereign); got != 2 {
		t.Errorf("\"other\" excludes the Sovereign: power %d, want 2", got)
	}
	if got := effectivePower(t, g, theirs); got != 2 {
		t.Errorf("\"you control\" excludes an opponent's Merfolk: power %d, want 2", got)
	}
	if got := effectivePower(t, g, bear); got != 2 {
		t.Errorf("a Bear is not a Merfolk: power %d, want 2", got)
	}

	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, sovereign, 0, game.ActivateAbilityParams{Targets: cardRefs(bear)}); err == nil {
		t.Fatal("the ability targets a Merfolk, not any creature")
	}
	// No "other" and no "you control" on the ability: an opponent's
	// Merfolk is a legal target.
	b16Activate(t, g, me.ID, sovereign, 0, game.ActivateAbilityParams{Targets: cardRefs(theirs)})
	passPriorityAroundTable(t, g)
	assertRestrictions(t, g, theirs, game.CantBeBlocked)
	// Nothing was put on anybody else.
	assertRestrictions(t, g, mine, 0)
}

// Charge counters are a cost, paid at announce, and three of them buy
// the destruction of anything.
func TestB39LuxCannonChargesThenDestroysAnyPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cannon := pushCatalogPermanent(g, me.ID, "Lux Cannon", "Artifact", b39LuxCannonOracle, false)
	land := b12Permanent(g, opp.ID, "Their Wastes", "Basic Land")

	advanceToMain(t, g)
	for i := 0; i < 3; i++ {
		b16Activate(t, g, me.ID, cannon, 0, game.ActivateAbilityParams{})
		passPriorityAroundTable(t, g)
		g.WithWriteLock(func() { _ = g.UntapTargetForEffect(cannon) })
		b39NextTurnOf(t, g, 0)
	}
	if got := counterOn(g, cannon, game.CounterCharge); got != 3 {
		t.Fatalf("three activations, three charge counters: %d", got)
	}
	b16Activate(t, g, me.ID, cannon, 1, game.ActivateAbilityParams{Targets: cardRefs(land)})
	// The counters are spent at announce, before the ability resolves.
	if got := counterOn(g, cannon, game.CounterCharge); got != 0 {
		t.Errorf("the three counters are a cost paid at announce: %d left", got)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) {
		t.Error("target permanent is destroyed — a land included")
	}
}

func TestB39LuxCannonCannotFireBelowThreeCounters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cannon := pushCatalogPermanent(g, me.ID, "Lux Cannon", "Artifact", b39LuxCannonOracle, false)
	land := b12Permanent(g, opp.ID, "Their Wastes", "Basic Land")
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(cannon, game.CounterCharge, 2) })
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, cannon, 1, game.ActivateAbilityParams{Targets: cardRefs(land)}); err == nil {
		t.Fatal("two counters cannot pay a cost of three")
	}
}

// Two disjoint sacrifice costs, and both tap.
func TestB39JanJansenEatsTheRightArtifactForEachAbility(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	jan := pushCatalogPermanent(g, me.ID, "Jan Jansen, Chaos Crafter", "Legendary Creature — Gnome Artificer", b39JanJansenOracle, false)
	servo := b39Creature(g, me.ID, "Servo", "Artifact Creature — Servo", 1, 1)
	rock := b12Permanent(g, me.ID, "Sol Ring", "Artifact")

	advanceToMain(t, g)
	// A noncreature artifact cannot pay the artifact-creature cost.
	if err := g.ActivateCatalogAbility(me.ID, jan, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{rock}}); err == nil {
		t.Fatal("a Sol Ring is not an artifact creature")
	}
	before := b39CountNamed(g, me.ID, "Treasure")
	b16Activate(t, g, me.ID, jan, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{servo}})
	passPriorityAroundTable(t, g)
	if got := b39CountNamed(g, me.ID, "Treasure") - before; got != 2 {
		t.Errorf("two Treasures: got %d", got)
	}

	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(jan) })
	b39NextTurnOf(t, g, 0)
	b16Activate(t, g, me.ID, jan, 1, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{rock}})
	passPriorityAroundTable(t, g)
	if got := b39CountNamed(g, me.ID, "Construct"); got != 2 {
		t.Errorf("two Constructs: got %d", got)
	}
}

// --- Shalai and Hallar ----------------------------------------------

func TestB39ShalaiAndHallarPingsForCountersOnYourCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Shalai and Hallar", "Legendary Creature — Angel Elf", b39ShalaiAndHallarOracle, false)
	bear := seedBear(g, me.ID)
	theirs := b39Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	rock := b12Permanent(g, me.ID, "Sol Ring", "Artifact")

	before := opp.Life
	// The EventResolve stamp is how the attribution read
	// (b12CountersPlacedByYou) learns whose ability is placing the
	// counters — All Will Be One's test sets the scene the same way.
	//
	// Counters on an artifact you control: not a creature, no trigger.
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventResolve, Actor: me.ID})
		_ = g.AddCounterForEffect(rock, game.CounterPlusOne, 2)
	})
	passPriorityAroundTable(t, g)
	// Counters on an opponent's creature: not yours, no trigger.
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventResolve, Actor: me.ID})
		_ = g.AddCounterForEffect(theirs, game.CounterPlusOne, 2)
	})
	passPriorityAroundTable(t, g)
	if opp.Life != before {
		t.Fatalf("neither an artifact nor an opponent's creature fires it: %d → %d", before, opp.Life)
	}

	// Three counters at once on your own creature is ONE trigger for
	// three damage, not three triggers for one.
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventResolve, Actor: me.ID})
		_ = g.AddCounterForEffect(bear, game.CounterPlusOne, 3)
	})
	b04WaitForPick(t, g, me.ID)
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != before-3 {
		t.Errorf("three counters on your creature is three damage: %d → %d", before, opp.Life)
	}
}

// --- Empowered Autogenerator ----------------------------------------

// The counter goes on first, so the first activation adds ONE.
func TestB39AutogeneratorAddsOneOnItsFirstTapAndGrows(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	gen := castCatalogSpell(t, g, "Empowered Autogenerator", "Artifact", b39AutogeneratorOracle, nil)
	passPriorityAroundTable(t, g)
	if !b16Tapped(t, g, gen) {
		t.Fatal("the Autogenerator enters tapped")
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(gen) })
	b28TapForMana(t, g, me.ID, gen, "U")
	if got := poolColors(me); len(got) != 1 {
		t.Errorf("the first tap adds exactly one mana: pool %v", got)
	}
	if got := counterOn(g, gen, game.CounterCharge); got != 1 {
		t.Errorf("one charge counter after the first tap: %d", got)
	}

	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(gen) })
	b28TapForMana(t, g, me.ID, gen, "U")
	if got := poolColors(me); len(got) != 3 {
		t.Errorf("the second tap adds two more: pool %v", got)
	}
	if got := counterOn(g, gen, game.CounterCharge); got != 2 {
		t.Errorf("two charge counters after the second tap: %d", got)
	}
}

// --- Anara, Wolvid Familiar -----------------------------------------

func TestB39AnaraGrantsIndestructibleOnlyOnYourOwnTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Anara, Wolvid Familiar", "Legendary Creature — Wolf Beast", b39AnaraOracle, false)
	mine := b36Commander(g, me.ID, "My Commander")
	theirs := b36Commander(g, opp.ID, "Their Commander")
	bear := seedBear(g, me.ID)

	b39NextTurnOf(t, g, 0)
	if !hasEffectiveKeyword(t, g, mine, "indestructible") {
		t.Error("on your turn your commander is indestructible")
	}
	if hasEffectiveKeyword(t, g, theirs, "indestructible") {
		t.Error("\"commanders you control\" does not cover an opponent's")
	}
	if hasEffectiveKeyword(t, g, bear, "indestructible") {
		t.Error("a creature that is not a commander gets nothing")
	}

	advanceToMainOf(t, g, 1)
	if hasEffectiveKeyword(t, g, mine, "indestructible") {
		t.Error("on somebody else's turn the grant is gone")
	}
}

// --- Airbending Lesson -----------------------------------------------

func TestB39AirbendingLessonExilesANonlandAndCantrips(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := b39Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	land := b12Permanent(g, opp.ID, "Their Wastes", "Basic Land")

	spec, _ := Lookup(b39AirbendingLessonOracle)
	if spec.Targets == nil || spec.Targets.CardOK == nil {
		t.Fatal("the Lesson has a structured target clause")
	}
	var landLegal, creatureLegal bool
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			switch c.InstanceID {
			case land:
				landLegal = spec.Targets.CardOK(g, me.ID, c, game.ZoneBattlefield)
			case theirs:
				creatureLegal = spec.Targets.CardOK(g, me.ID, c, game.ZoneBattlefield)
			}
		}
	})
	if landLegal {
		t.Error("a land is never a legal target")
	}
	if !creatureLegal {
		t.Fatal("a nonland permanent is")
	}

	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Airbending Lesson", "Instant — Lesson", b39AirbendingLessonOracle, cardRefs(theirs))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("the targeted permanent is exiled")
	}
	// The helper seeds the card into hand and casts it (net zero
	// against `hand`); the spell then draws one.
	if me.Hand.Size() != hand+1 {
		t.Errorf("the Lesson draws a card: hand %d → %d", hand, me.Hand.Size())
	}
}

// --- Inquisitor's Flail ----------------------------------------------

func TestB39InquisitorsFlailDoublesBothDirectionsOfCombatDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	flail := seedEquipment(g, me.ID, "Inquisitor's Flail", b39InquisitorsFlailOracle)
	equipTo(t, g, me.ID, flail, bear)

	// Noncombat damage is never doubled — the second clause of each
	// replacement is IsCombatDamage.
	life := opp.Life
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(bear, opp.ID, 3) })
	if opp.Life != life-3 {
		t.Errorf("noncombat damage is untouched: %d → %d, want %d", life, opp.Life, life-3)
	}

	// Outbound: a real combat damage step through the declare-attack
	// path, so the replacement pipeline actually runs.
	before := opp.Life
	declareAttack(t, g, opp.ID, bear)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if opp.Life != before-4 {
		t.Errorf("the equipped creature's 2 power hits for 4: %d → %d, want %d", before, opp.Life, before-4)
	}
}

// Inbound: another creature's combat damage to the equipped creature
// is doubled too, which is what makes the Flail a liability on a
// blocker. A 1/1 blocker kills a Flailed 2/2.
func TestB39InquisitorsFlailDoublesTheDamageComingBack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	flail := seedEquipment(g, me.ID, "Inquisitor's Flail", b39InquisitorsFlailOracle)
	equipTo(t, g, me.ID, flail, bear)
	blocker := b39Creature(g, opp.ID, "Their Wall", "Creature — Wall", 1, 4)

	declareAttack(t, g, opp.ID, bear)
	b35Block(t, g, blocker, bear)
	lockInBlocks(t, g)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)

	// The Bear's 2 is doubled to 4, which kills a 1/4; the blocker's 1
	// is doubled to 2, which kills the 2/2 wearing the Flail. Both
	// halves fire, and the trade the Flail turns into a double kill is
	// the whole point of the second clause.
	if g.Battlefield.Contains(blocker) {
		t.Error("the Flailed Bear's doubled 4 kills a 1/4 blocker")
	}
	if g.Battlefield.Contains(bear) {
		t.Error("the blocker's doubled 1 kills the Flailed 2/2 right back")
	}
}

// Without the Flail the same exchange kills nothing: the doubling is
// the Equipment's, not the combat step's.
func TestB39AFlaillessTradeKillsNeitherCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	blocker := b39Creature(g, opp.ID, "Their Wall", "Creature — Wall", 1, 4)

	declareAttack(t, g, opp.ID, bear)
	b35Block(t, g, blocker, bear)
	lockInBlocks(t, g)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(blocker) || !g.Battlefield.Contains(bear) {
		t.Error("2 into a 1/4 and 1 into a 2/2 kills nothing without the Flail")
	}
}

// --- Dawn of a New Age ------------------------------------------------

func TestB39DawnOfANewAgeDrawsPerCreatureThenCashesIn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedBear(g, me.ID)
	b39Creature(g, me.ID, "Second Bear", "Creature — Bear", 2, 2)

	dawn := castCatalogSpell(t, g, "Dawn of a New Age", "Enchantment", b39DawnOfANewAgeOracle, nil)
	passPriorityAroundTable(t, g)
	if got := counterOn(g, dawn, "hope"); got != 2 {
		t.Fatalf("a hope counter for each of the two creatures: %d", got)
	}
	// A creature arriving afterwards adds nothing — it is an entry
	// replacement, not a static.
	b39Creature(g, me.ID, "Third Bear", "Creature — Bear", 2, 2)
	if got := counterOn(g, dawn, "hope"); got != 2 {
		t.Errorf("the count is taken once, as it enters: %d", got)
	}

	hand, life := me.Hand.Size(), me.Life
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if got := counterOn(g, dawn, "hope"); got != 1 {
		t.Errorf("one counter comes off at your end step: %d", got)
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("removing a counter draws a card: hand %d → %d", hand, me.Hand.Size())
	}
	if !g.Battlefield.Contains(dawn) {
		t.Fatal("a counter is left, so it is not sacrificed yet")
	}

	b39NextTurnOf(t, g, 0)
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(dawn) {
		t.Error("the turn that spends the last counter sacrifices it")
	}
	if me.Life != life+4 {
		t.Errorf("and gains 4 life: %d → %d", life, me.Life)
	}
}

// --- Marketback Walker ------------------------------------------------

func TestB39MarketbackWalkerEntersWithXAndDrawsThatManyWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	walker := b19CastXModal(t, g, "Marketback Walker", "Artifact Creature — Construct",
		b39MarketbackWalkerOracle, "{X}{X}", 3, nil, nil)
	passPriorityAroundTable(t, g)
	if got := counterOn(g, walker, game.CounterPlusOne); got != 3 {
		t.Fatalf("cast for X=3, three +1/+1 counters: %d", got)
	}
	if got := b39CurrentPower(t, g, walker); got != 3 {
		t.Errorf("a 0/0 with three counters is a 3/3: power %d", got)
	}

	advanceToMain(t, g)
	b06AddMana(me, "C", "C", "C", "C")
	b16Activate(t, g, me.ID, walker, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if got := counterOn(g, walker, game.CounterPlusOne); got != 4 {
		t.Fatalf("{4} adds a fourth counter: %d", got)
	}

	hand := me.Hand.Size()
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(walker) })
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+4 {
		t.Errorf("it draws a card for each counter it had: hand %d → %d, want %d", hand, me.Hand.Size(), hand+4)
	}
}

// --- planeswalkers ----------------------------------------------------

func TestB39AjaniGrantsVigilanceAndTicksTheOtherWalkers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ajani := pushCatalogPermanent(g, me.ID, "Ajani, the Greathearted", "Legendary Planeswalker — Ajani", b39AjaniOracle, false)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(ajani, game.CounterLoyalty, 5) })
	bear := seedBear(g, me.ID)
	theirs := b39Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	other := pushCatalogPermanent(g, me.ID, "Other Walker", "Legendary Planeswalker — Teferi", "", false)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(other, game.CounterLoyalty, 3) })

	if !hasEffectiveKeyword(t, g, bear, "vigilance") {
		t.Error("creatures you control have vigilance")
	}
	if hasEffectiveKeyword(t, g, theirs, "vigilance") {
		t.Error("an opponent's creature gets nothing")
	}

	advanceToMain(t, g)
	life := me.Life
	b16Activate(t, g, me.ID, ajani, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("+1 gains 3 life: %d → %d", life, me.Life)
	}

	b39NextTurnOf(t, g, 0)
	b16Activate(t, g, me.ID, ajani, 1, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if got := counterOn(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("a +1/+1 counter on each creature you control: %d", got)
	}
	if got := counterOn(g, theirs, game.CounterPlusOne); got != 0 {
		t.Errorf("not on an opponent's: %d", got)
	}
	if got := counterOn(g, other, game.CounterLoyalty); got != 4 {
		t.Errorf("a loyalty counter on each OTHER walker you control: %d, want 4", got)
	}
}

func TestB39ObNixilisBurnsEveryOpponentDrawAndHelpsTheVictim(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ob := pushCatalogPermanent(g, me.ID, "Ob Nixilis, the Hate-Twisted", "Legendary Planeswalker — Nixilis", b39ObNixilisOracle, false)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(ob, game.CounterLoyalty, 5) })
	theirs := b39Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	// Your own draws are free.
	mine := me.Life
	hand := me.Hand.Size()
	g.WithWriteLock(func() { _ = g.DrawNForEffect(me.ID, 1) })
	passPriorityAroundTable(t, g)
	if me.Life != mine {
		t.Errorf("Ob Nixilis does not burn his own controller: %d → %d", mine, me.Life)
	}

	advanceToMain(t, g)
	before := opp.Life
	oppHand := opp.Hand.Size()
	b16Activate(t, g, me.ID, ob, 0, game.ActivateAbilityParams{Targets: cardRefs(theirs)})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("the -2 destroys the target")
	}
	if opp.Hand.Size() != oppHand+2 {
		t.Errorf("its controller draws two: hand %d → %d", oppHand, opp.Hand.Size())
	}
	// Two draws, two pings from the passive.
	if opp.Life != before-2 {
		t.Errorf("each of those draws is a ping: %d → %d, want %d", before, opp.Life, before-2)
	}
	_ = hand
}

// "Becomes 10" is a set, not a loss: a player below 10 gains.
func TestB39SorinMinusThreeSetsTheLifeTotalInBothDirections(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sorin := pushCatalogPermanent(g, me.ID, "Sorin Markov", "Legendary Planeswalker — Sorin", b39SorinMarkovOracle, false)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(sorin, game.CounterLoyalty, 10) })

	advanceToMain(t, g)
	opp.Life = 38
	b16Activate(t, g, me.ID, sorin, 1, game.ActivateAbilityParams{Targets: b39PlayerRefs(opp.ID)})
	passPriorityAroundTable(t, g)
	if opp.Life != 10 {
		t.Errorf("38 becomes 10: %d", opp.Life)
	}

	b39NextTurnOf(t, g, 0)
	opp.Life = 4
	b16Activate(t, g, me.ID, sorin, 1, game.ActivateAbilityParams{Targets: b39PlayerRefs(opp.ID)})
	passPriorityAroundTable(t, g)
	if opp.Life != 10 {
		t.Errorf("4 becomes 10 — the clause is a set, not a loss: %d", opp.Life)
	}

	spec, _ := Lookup(b39SorinMarkovOracle)
	if len(spec.Activated) != 2 {
		t.Errorf("the -7 is deliberately not offered: %d abilities", len(spec.Activated))
	}
}

// --- spells -----------------------------------------------------------

func TestB39ThrillingDiscoveryGainsTwoThenTradesTwoForThree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	life, hand := me.Life, me.Hand.Size()
	castCatalogSpell(t, g, "Thrilling Discovery", "Sorcery", b39ThrillingDiscoveryOracle, nil)
	passPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("the 2 life is unconditional: %d → %d", life, me.Life)
	}
	if got := discardOwed(g, me.ID); got != 2 {
		t.Fatalf("the prompt offers up to two: owes %d", got)
	}
	answerDiscard(t, g, me.ID, b39HandIDs(me, 2)...)
	passPriorityAroundTable(t, g)
	// The helper seeded and cast the sorcery (net zero), then −2
	// discarded and +3 drawn.
	if me.Hand.Size() != hand+1 {
		t.Errorf("two for three: hand %d → %d, want %d", hand, me.Hand.Size(), hand+1)
	}
}

// Declining the discard is legal and draws nothing — the "if you do"
// is a real condition.
func TestB39ThrillingDiscoveryDeclinedStillGainsTheTwoLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	life, hand := me.Life, me.Hand.Size()
	castCatalogSpell(t, g, "Thrilling Discovery", "Sorcery", b39ThrillingDiscoveryOracle, nil)
	passPriorityAroundTable(t, g)
	answerDiscard(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("the life is unconditional: %d → %d", life, me.Life)
	}
	if me.Hand.Size() != hand {
		t.Errorf("declining draws nothing: hand %d → %d, want %d", hand, me.Hand.Size(), hand)
	}
}

func TestB39CatharticPyreTakesEitherModeAndNeverAPlayer(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	theirs := b39Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	spec, _ := Lookup(b39CatharticPyreOracle)
	if spec.Modes == nil || len(spec.Modes.Options) != 2 || spec.Modes.Min != 1 || spec.Modes.Max != 1 {
		t.Fatalf("Cathartic Pyre is choose one of two: %+v", spec.Modes)
	}
	if spec.Modes.Options[0].Targets == nil || spec.Modes.Options[0].Targets.Players {
		t.Error("the damage mode targets a creature or planeswalker, never a player")
	}

	before := opp.Life
	b19CastXModal(t, g, "Cathartic Pyre", "Instant", b39CatharticPyreOracle, "{1}{R}", 0, []int{0}, cardRefs(theirs))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("3 damage kills a 2/2")
	}
	if opp.Life != before {
		t.Errorf("the mode cannot hit a player: %d → %d", before, opp.Life)
	}
}

func TestB39BanefireDealsXAndDeclaresItsMissingRiders(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	before := opp.Life
	b19CastXModal(t, g, "Banefire", "Sorcery", b39BanefireOracle, "{X}{R}", 6, nil, b39PlayerRefs(opp.ID))
	passPriorityAroundTable(t, g)
	if opp.Life != before-6 {
		t.Errorf("X damage to any target: %d → %d, want %d", before, opp.Life, before-6)
	}

	spec, _ := Lookup(b39BanefireOracle)
	if spec.CantBeCountered {
		t.Error("the uncounterable rider is conditional on X and is deliberately NOT set — setting it would make a small Banefire uncounterable, which is stronger than printed")
	}
	if spec.Completeness != CompletenessCaveats || len(spec.Caveats) != 2 {
		t.Errorf("both missing riders are declared: %v %v", spec.Completeness, spec.Caveats)
	}
	_ = me
}

// --- Nested Shambler ---------------------------------------------------

// X is the LAST-KNOWN power, so counters and anthems count.
func TestB39NestedShamblerMakesTappedSquirrelsForItsLastKnownPower(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	shambler := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Nested Shambler", TypeLine: "Creature — Zombie",
		OracleID: b39NestedShamblerOracle, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(shambler, game.CounterPlusOne, 2) })
	if got := b39CurrentPower(t, g, shambler); got != 3 {
		t.Fatalf("setup: a 1/1 with two counters is a 3/3, got %d", got)
	}

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(shambler) })
	passPriorityAroundTable(t, g)
	squirrels := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Squirrel" {
			squirrels++
			if !c.Tapped {
				t.Error("the Squirrels enter tapped")
			}
		}
	}
	if squirrels != 3 {
		t.Errorf("three Squirrels for its last-known power of 3: got %d", squirrels)
	}
}

// --- Erebos, Bleak-Hearted ---------------------------------------------

func TestB39ErebosIsNotACreatureBelowFiveDevotion(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	erebos := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Erebos, Bleak-Hearted",
		TypeLine: "Legendary Enchantment Creature — God", OracleID: b39ErebosOracle,
		ManaCost: "{3}{B}", Power: 5, Toughness: 6, Owner: me.ID, Controller: me.ID,
	})
	// His own {B} is devotion 1 — well short of five.
	if b39IsCreature(t, g, erebos) {
		t.Error("devotion 1 is less than five: Erebos isn't a creature")
	}
	for i := 0; i < 4; i++ {
		pushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: "Black Rock", TypeLine: "Artifact",
			ManaCost: "{B}", Owner: me.ID, Controller: me.ID,
		})
	}
	if !b39IsCreature(t, g, erebos) {
		t.Error("devotion 5 turns him on")
	}
	if !hasEffectiveKeyword(t, g, erebos, "indestructible") {
		t.Error("indestructible is printed and is live either way")
	}
}

func TestB39ErebosDrawsForTwoLifeAndShrinksForASacrifice(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	erebos := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Erebos, Bleak-Hearted",
		TypeLine: "Legendary Enchantment Creature — God", OracleID: b39ErebosOracle,
		ManaCost: "{3}{B}", Power: 5, Toughness: 6, Owner: me.ID, Controller: me.ID,
	})
	bear := seedBear(g, me.ID)
	fodder := b39Creature(g, me.ID, "Fodder", "Creature — Goblin", 1, 1)

	hand, life := me.Hand.Size(), me.Life
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(fodder) })
	b06AnswerAllTriggerPrompts(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 || me.Life != life-2 {
		t.Errorf("another creature dying pays 2 life for a card: hand %d → %d, life %d → %d",
			hand, me.Hand.Size(), life, me.Life)
	}

	advanceToMain(t, g)
	victim := b39Creature(g, me.ID, "Second Fodder", "Creature — Goblin", 1, 1)
	b06AddMana(me, "C", "B")
	b16Activate(t, g, me.ID, erebos, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{victim},
		Targets:      cardRefs(bear),
	})
	b06AnswerAllTriggerPrompts(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if got := effectivePower(t, g, bear); got != 0 {
		t.Errorf("-2/-1 on a 2/2 leaves a 0/1: power %d", got)
	}
	if got := effectiveToughness(t, g, bear); got != 1 {
		t.Errorf("-2/-1 on a 2/2 leaves a 0/1: toughness %d", got)
	}
}

// b39HandIDs returns the first n card IDs in a player's hand.
func b39HandIDs(p *game.Player, n int) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range p.Hand.Cards {
		if len(out) == n {
			break
		}
		out = append(out, c.InstanceID)
	}
	return out
}
