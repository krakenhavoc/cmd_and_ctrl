package testsav

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- registration --------------------------------------------------

// Every card in the batch, pinned by oracle ID → name. The two snow
// duals are rows in the typed-tapland cycle table, so a transposed row
// is invisible until somebody plays that exact land.
func TestBatch39CardsAreRegistered(t *testing.T) {
	want := map[string]string{
		effectstest.B39GlacialFloodplainOracle:  "Glacial Floodplain",
		effectstest.B39SnowfieldSinkholeOracle:  "Snowfield Sinkhole",
		effectstest.B39DragonSniperOracle:       "Dragon Sniper",
		effectstest.B39ZoZuOracle:               "Zo-Zu the Punisher",
		effectstest.B39MaskOfAvacynOracle:       "Mask of Avacyn",
		effectstest.B39MaulOfTheSkyclavesOracle: "Maul of the Skyclaves",
		effectstest.B39RoguesGlovesOracle:       "Rogue's Gloves",
		effectstest.B39FeatherOfFlightOracle:    "Feather of Flight",
		effectstest.B39KorCartographerOracle:    "Kor Cartographer",
		effectstest.B39BloodlineNecroOracle:     "Bloodline Necromancer",
		effectstest.B39SpawningKrakenOracle:     "Spawning Kraken",
		effectstest.B39NecropolisRegentOracle:   "Necropolis Regent",
		effectstest.B39MerfolkSovereignOracle:   "Merfolk Sovereign",
		effectstest.B39LuxCannonOracle:          "Lux Cannon",
		effectstest.B39NestedShamblerOracle:     "Nested Shambler",
		effectstest.B39ThrillingDiscoveryOracle: "Thrilling Discovery",
		effectstest.B39JanJansenOracle:          "Jan Jansen, Chaos Crafter",
		effectstest.B39BezaOracle:               "Beza, the Bounding Spring",
		effectstest.B39ShalaiAndHallarOracle:    "Shalai and Hallar",
		effectstest.B39AutogeneratorOracle:      "Empowered Autogenerator",
		effectstest.B39AnaraOracle:              "Anara, Wolvid Familiar",
		effectstest.B39AirbendingLessonOracle:   "Airbending Lesson",
		effectstest.B39InquisitorsFlailOracle:   "Inquisitor's Flail",
		effectstest.B39DawnOfANewAgeOracle:      "Dawn of a New Age",
		effectstest.B39MarketbackWalkerOracle:   "Marketback Walker",
		effectstest.B39AjaniOracle:              "Ajani, the Greathearted",
		effectstest.B39ObNixilisOracle:          "Ob Nixilis, the Hate-Twisted",
		effectstest.B39SorinMarkovOracle:        "Sorin Markov",
		effectstest.B39CatharticPyreOracle:      "Cathartic Pyre",
		effectstest.B39BanefireOracle:           "Banefire",
		effectstest.B39ErebosOracle:             "Erebos, Bleak-Hearted",
	}
	if len(want) != 31 {
		t.Fatalf("the batch registers 31 cards, the table lists %d", len(want))
	}
	for oracle, name := range want {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name {
			t.Errorf("oracle %s registered as %q, want %q", oracle, spec.Name, name)
		}
	}
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
		{"Glacial Floodplain", "Snow Land — Plains Island", effectstest.B39GlacialFloodplainOracle, "W", "U"},
		{"Snowfield Sinkhole", "Snow Land — Plains Swamp", effectstest.B39SnowfieldSinkholeOracle, "W", "B"},
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		land := effectstest.B12PlayFromHand(t, g, row.name, row.typeLine, row.oracle, game.CastSpellParams{})
		if !effectstest.B16Tapped(t, g, land) {
			t.Fatalf("%s: the land enters tapped", row.name)
		}
		g.WithWriteLock(func() { _ = g.UntapTargetForEffect(land) })
		effectstest.B28TapForMana(t, g, me.ID, land, row.a)
		if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != row.a {
			t.Errorf("%s: tapped for {%s}: pool %v", row.name, row.a, got)
		}
		g.WithWriteLock(func() { _ = g.UntapTargetForEffect(land) })
		effectstest.B28TapForMana(t, g, me.ID, land, row.b)
		if got := effectstest.PoolColors(me); len(got) != 2 || got[1] != row.b {
			t.Errorf("%s: also taps for {%s}: pool %v", row.name, row.b, got)
		}
	}
}

// --- vanilla keywords ----------------------------------------------

func TestB39DragonSniperHasItsThreeKeywords(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	sniper := effectstest.PushCatalogPermanent(g, me.ID, "Dragon Sniper", "Creature — Human Archer", effectstest.B39DragonSniperOracle, false)
	for _, kw := range []string{"reach", "vigilance", "deathtouch"} {
		if !effectstest.HasEffectiveKeyword(t, g, sniper, kw) {
			t.Errorf("Dragon Sniper should have %s", kw)
		}
	}
	if effectstest.HasEffectiveKeyword(t, g, sniper, "flying") {
		t.Error("reach is not flying")
	}
}

// --- Zo-Zu the Punisher --------------------------------------------

// Every land, whoever plays it, including Zo-Zu's own controller.
func TestB39ZoZuBurnsWhoeverPlaysALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Zo-Zu the Punisher", "Legendary Creature — Goblin Warrior", effectstest.B39ZoZuOracle, false)

	mine := me.Life
	effectstest.B12PlayFromHand(t, g, "Wastes", "Basic Land", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
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
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != theirs-2 {
		t.Errorf("the land's controller takes the 2, not the active player: %d → %d", theirs, opp.Life)
	}
	if me.Life != mine-2 {
		t.Errorf("you took nothing for their land: life %d", me.Life)
	}
}

// --- Equipment -----------------------------------------------------

func TestB39MaskOfAvacynPumpsAndGrantsHexproof(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, me.ID)
	mask := effectstest.SeedEquipment(g, me.ID, "Mask of Avacyn", effectstest.B39MaskOfAvacynOracle)
	effectstest.EquipTo(t, g, me.ID, mask, bear)

	if got := effectstest.EffectivePower(t, g, bear); got != 3 {
		t.Errorf("equipped power %d, want 3", got)
	}
	if got := effectstest.EffectiveToughness(t, g, bear); got != 4 {
		t.Errorf("equipped toughness %d, want 4", got)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "hexproof") {
		t.Error("the equipped creature has hexproof")
	}
}

// The entry trigger attaches the Maul to a creature YOU control, and
// it is a real targeted trigger — not an as-enters hook.
func TestB39MaulOfTheSkyclavesAttachesItselfOnEntry(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.SeedBear(g, me.ID)
	theirs := effectstest.B39Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	maul := effectstest.CastCatalogSpell(t, g, "Maul of the Skyclaves", effectstest.EquipTypeLine, effectstest.B39MaulOfTheSkyclavesOracle, nil)
	// The trigger goes on the stack as the Equipment enters and names
	// its target then — an opponent's creature is never offered.
	effectstest.B04WaitForPick(t, g, me.ID)
	if p := effectstest.LatestPickTarget(g, me.ID); effectstest.HasID(p.PickTargetCards, theirs) || !effectstest.HasID(p.PickTargetCards, bear) {
		t.Fatalf("the entry trigger targets a creature you control: %+v", p)
	}
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)

	if host := effectstest.AttachmentHostOf(t, g, maul); host.Kind != game.TargetCard || host.ID != bear {
		t.Fatalf("the Maul attached itself to the Bear: %+v", host)
	}
	if got := effectstest.EffectivePower(t, g, bear); got != 4 {
		t.Errorf("equipped power %d, want 4", got)
	}
	for _, kw := range []string{"flying", "first strike"} {
		if !effectstest.HasEffectiveKeyword(t, g, bear, kw) {
			t.Errorf("the equipped creature has %s", kw)
		}
	}
}

// Combat damage only, and the "you may" is a real prompt.
func TestB39RoguesGlovesDrawOnCombatDamageOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, me.ID)
	gloves := effectstest.SeedEquipment(g, me.ID, "Rogue's Gloves", effectstest.B39RoguesGlovesOracle)
	effectstest.EquipTo(t, g, me.ID, gloves, bear)

	hand := me.Hand.Size()
	effectstest.B27Damage(g, bear, opp.ID, 2) // noncombat
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("noncombat damage draws nothing: hand %d → %d", hand, me.Hand.Size())
	}

	effectstest.DealCombatDamageToPlayer(g, bear, opp.ID, 2)
	effectstest.B06AnswerAllTriggerPrompts(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("combat damage to a player draws one: hand %d → %d", hand, me.Hand.Size())
	}
}

// --- Auras ---------------------------------------------------------

func TestB39FeatherOfFlightCantripsAndGrantsFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.SeedBear(g, me.ID)
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Feather of Flight", "Enchantment — Aura", effectstest.B39FeatherOfFlightOracle, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)

	// The helper seeds the Aura into hand and casts it (net zero
	// against the pre-cast count); the entry trigger then draws one.
	if me.Hand.Size() != hand+1 {
		t.Errorf("the Aura replaces itself: hand %d → %d", hand, me.Hand.Size())
	}
	if got := effectstest.EffectivePower(t, g, bear); got != 3 {
		t.Errorf("enchanted power %d, want 3", got)
	}
	if got := effectstest.EffectiveToughness(t, g, bear); got != 2 {
		t.Errorf("Feather of Flight is +1/+0: toughness %d, want 2", got)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("the enchanted creature has flying")
	}
	spec, _ := effects.Lookup(effectstest.B39FeatherOfFlightOracle)
	if len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != "flash" {
		t.Errorf("flash is printed on the Aura: %v", spec.PrintedKeywords)
	}
}

// --- enters triggers ------------------------------------------------

// "A Plains card" is the subtype, not the basic — and it lands tapped.
func TestB39KorCartographerTakesAnyPlainsCardTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Hallowed Fountain", TypeLine: "Land — Plains Island"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
	)
	effectstest.CastCatalogSpell(t, g, "Kor Cartographer", "Creature — Kor Scout", effectstest.B39KorCartographerOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("\"you may search\": the chooser asks")
	}
	if effectstest.SearchOptionNamed(g, c, "Hallowed Fountain") == uuid.Nil {
		t.Error("a nonbasic Plains is a Plains card")
	}
	if effectstest.SearchOptionNamed(g, c, "Forest") != uuid.Nil {
		t.Error("a Forest is not a Plains card")
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Hallowed Fountain")
	if !g.Battlefield.Contains(ids[0]) {
		t.Fatal("the pick goes onto the battlefield")
	}
	if !effectstest.B16Tapped(t, g, ids[0]) {
		t.Error("the fetched land enters tapped")
	}
}

func TestB39BloodlineNecromancerReanimatesOnlyAVampireOrWizard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	vampire := effectstest.PushGraveyardPermanent(me, "Vampire Nighthawk", "Creature — Vampire Shaman", "{1}{B}{B}")
	wizard := effectstest.PushGraveyardPermanent(me, "Merfolk Looter", "Creature — Merfolk Wizard", "{1}{U}")
	bear := effectstest.PushGraveyardPermanent(me, "Bear", "Creature — Bear", "{1}{G}")

	effectstest.CastCatalogSpell(t, g, "Bloodline Necromancer", "Creature — Vampire Wizard", effectstest.B39BloodlineNecroOracle, nil)
	// The target is chosen as the trigger goes on the stack, after the
	// creature spell has resolved; the "you may" waits for the
	// trigger's own resolution (CR 603.5).
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if effectstest.HasID(p.PickTargetCards, bear) {
		t.Error("a Bear is neither a Vampire nor a Wizard")
	}
	if !effectstest.HasID(p.PickTargetCards, vampire) || !effectstest.HasID(p.PickTargetCards, wizard) {
		t.Error("either subtype on its own qualifies")
	}
	effectstest.PickCard(t, g, me.ID, vampire)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(vampire) {
		t.Error("it comes back to the battlefield, not to hand")
	}
}

// Four independent comparisons, each against any one opponent.
func TestB39BezaPaysOnlyForTheAxesAnOpponentLeadsOn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]

	// One opponent leads on lands; a different one leads on creatures.
	// Beza is on the battlefield when her own trigger resolves, so the
	// creature comparison is against a board of one — two of theirs is
	// what beats it.
	for i := 0; i < 3; i++ {
		effectstest.B12Permanent(g, opp.ID, "Wastes", "Basic Land")
	}
	effectstest.B39Creature(g, other.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.B39Creature(g, other.ID, "Their Second Bear", "Creature — Bear", 2, 2)
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
	effectstest.CastCatalogSpell(t, g, "Beza, the Bounding Spring", "Legendary Creature — Elemental Elk", effectstest.B39BezaOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	kraken := effectstest.PushCatalogPermanent(g, me.ID, "Spawning Kraken", "Creature — Kraken", effectstest.B39SpawningKrakenOracle, false)
	serpent := effectstest.B39Creature(g, me.ID, "Serpent", "Creature — Serpent", 3, 3)
	bear := effectstest.SeedBear(g, me.ID)

	effectstest.DealCombatDamageToPlayer(g, bear, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if n := b39CountNamed(g, me.ID, "Kraken"); n != 0 {
		t.Errorf("a Bear is none of the four types: %d tokens", n)
	}

	effectstest.DealCombatDamageToPlayer(g, kraken, opp.ID, 6)
	effectstest.DealCombatDamageToPlayer(g, serpent, opp.ID, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if n := b39CountNamed(g, me.ID, "Kraken"); n != 2 {
		t.Errorf("the Kraken itself and a Serpent each connect: %d tokens, want 2", n)
	}
}

// "That many" is the damage dealt, and it lands on the creature that
// dealt it — not on the Regent.
func TestB39NecropolisRegentGrowsWhicheverCreatureConnected(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	regent := effectstest.PushCatalogPermanent(g, me.ID, "Necropolis Regent", "Creature — Vampire", effectstest.B39NecropolisRegentOracle, false)
	bear := effectstest.SeedBear(g, me.ID)

	effectstest.DealCombatDamageToPlayer(g, bear, opp.ID, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterOn(g, bear, game.CounterPlusOne); got != 3 {
		t.Errorf("the Bear grew by the damage it dealt: %d counters, want 3", got)
	}
	if got := effectstest.CounterOn(g, regent, game.CounterPlusOne); got != 0 {
		t.Errorf("the counters go on the creature that connected, not the Regent: %d", got)
	}
}

// --- lords and activated abilities ---------------------------------

func TestB39MerfolkSovereignPumpsOtherMerfolkAndUnblocksAny(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sovereign := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Merfolk Sovereign", TypeLine: "Creature — Merfolk Noble",
		OracleID: effectstest.B39MerfolkSovereignOracle, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	mine := effectstest.B39Creature(g, me.ID, "Lullmage Mentor", "Creature — Merfolk Wizard", 2, 2)
	theirs := effectstest.B39Creature(g, opp.ID, "Their Merfolk", "Creature — Merfolk", 2, 2)
	bear := effectstest.SeedBear(g, me.ID)

	if got := effectstest.EffectivePower(t, g, mine); got != 3 {
		t.Errorf("your other Merfolk gets +1/+1: power %d, want 3", got)
	}
	if got := effectstest.EffectivePower(t, g, sovereign); got != 2 {
		t.Errorf("\"other\" excludes the Sovereign: power %d, want 2", got)
	}
	if got := effectstest.EffectivePower(t, g, theirs); got != 2 {
		t.Errorf("\"you control\" excludes an opponent's Merfolk: power %d, want 2", got)
	}
	if got := effectstest.EffectivePower(t, g, bear); got != 2 {
		t.Errorf("a Bear is not a Merfolk: power %d, want 2", got)
	}

	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, sovereign, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)}); err == nil {
		t.Fatal("the ability targets a Merfolk, not any creature")
	}
	// No "other" and no "you control" on the ability: an opponent's
	// Merfolk is a legal target.
	effectstest.B16Activate(t, g, me.ID, sovereign, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(theirs)})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AssertRestrictions(t, g, theirs, game.CantBeBlocked)
	// Nothing was put on anybody else.
	effectstest.AssertRestrictions(t, g, mine, 0)
}

// Charge counters are a cost, paid at announce, and three of them buy
// the destruction of anything.
func TestB39LuxCannonChargesThenDestroysAnyPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cannon := effectstest.PushCatalogPermanent(g, me.ID, "Lux Cannon", "Artifact", effectstest.B39LuxCannonOracle, false)
	land := effectstest.B12Permanent(g, opp.ID, "Their Wastes", "Basic Land")

	effectstest.AdvanceToMain(t, g)
	for i := 0; i < 3; i++ {
		effectstest.B16Activate(t, g, me.ID, cannon, 0, game.ActivateAbilityParams{})
		effectstest.PassPriorityAroundTable(t, g)
		g.WithWriteLock(func() { _ = g.UntapTargetForEffect(cannon) })
		effectstest.B39NextTurnOf(t, g, 0)
	}
	if got := effectstest.CounterOn(g, cannon, game.CounterCharge); got != 3 {
		t.Fatalf("three activations, three charge counters: %d", got)
	}
	effectstest.B16Activate(t, g, me.ID, cannon, 1, game.ActivateAbilityParams{Targets: effectstest.CardRefs(land)})
	// The counters are spent at announce, before the ability resolves.
	if got := effectstest.CounterOn(g, cannon, game.CounterCharge); got != 0 {
		t.Errorf("the three counters are a cost paid at announce: %d left", got)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) {
		t.Error("target permanent is destroyed — a land included")
	}
}

func TestB39LuxCannonCannotFireBelowThreeCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cannon := effectstest.PushCatalogPermanent(g, me.ID, "Lux Cannon", "Artifact", effectstest.B39LuxCannonOracle, false)
	land := effectstest.B12Permanent(g, opp.ID, "Their Wastes", "Basic Land")
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(cannon, game.CounterCharge, 2) })
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, cannon, 1, game.ActivateAbilityParams{Targets: effectstest.CardRefs(land)}); err == nil {
		t.Fatal("two counters cannot pay a cost of three")
	}
}

// Two disjoint sacrifice costs, and both tap.
func TestB39JanJansenEatsTheRightArtifactForEachAbility(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	jan := effectstest.PushCatalogPermanent(g, me.ID, "Jan Jansen, Chaos Crafter", "Legendary Creature — Gnome Artificer", effectstest.B39JanJansenOracle, false)
	servo := effectstest.B39Creature(g, me.ID, "Servo", "Artifact Creature — Servo", 1, 1)
	rock := effectstest.B12Permanent(g, me.ID, "Sol Ring", "Artifact")

	effectstest.AdvanceToMain(t, g)
	// A noncreature artifact cannot pay the artifact-creature cost.
	if err := g.ActivateCatalogAbility(me.ID, jan, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{rock}}); err == nil {
		t.Fatal("a Sol Ring is not an artifact creature")
	}
	before := b39CountNamed(g, me.ID, "Treasure")
	effectstest.B16Activate(t, g, me.ID, jan, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{servo}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := b39CountNamed(g, me.ID, "Treasure") - before; got != 2 {
		t.Errorf("two Treasures: got %d", got)
	}

	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(jan) })
	effectstest.B39NextTurnOf(t, g, 0)
	effectstest.B16Activate(t, g, me.ID, jan, 1, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{rock}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := b39CountNamed(g, me.ID, "Construct"); got != 2 {
		t.Errorf("two Constructs: got %d", got)
	}
}

// --- Shalai and Hallar ----------------------------------------------

func TestB39ShalaiAndHallarPingsForCountersOnYourCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Shalai and Hallar", "Legendary Creature — Angel Elf", effectstest.B39ShalaiAndHallarOracle, false)
	bear := effectstest.SeedBear(g, me.ID)
	theirs := effectstest.B39Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	rock := effectstest.B12Permanent(g, me.ID, "Sol Ring", "Artifact")

	before := opp.Life
	// The EventResolve stamps set the scene as a real resolution would;
	// who placed the counters does not matter to this card (#3863).
	//
	// Counters on an artifact you control: not a creature, no trigger.
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventResolve, Actor: me.ID})
		_ = g.AddCounterForEffect(rock, game.CounterPlusOne, 2)
	})
	effectstest.PassPriorityAroundTable(t, g)
	// Counters on an opponent's creature: not yours, no trigger.
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventResolve, Actor: me.ID})
		_ = g.AddCounterForEffect(theirs, game.CounterPlusOne, 2)
	})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before {
		t.Fatalf("neither an artifact nor an opponent's creature fires it: %d → %d", before, opp.Life)
	}

	// Three counters at once on your own creature is ONE trigger for
	// three damage, not three triggers for one.
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventResolve, Actor: me.ID})
		_ = g.AddCounterForEffect(bear, game.CounterPlusOne, 3)
	})
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-3 {
		t.Errorf("three counters on your creature is three damage: %d → %d", before, opp.Life)
	}
}

// #3863: the clause is passive ("are put on"), so an OPPONENT'S effect
// growing your creature triggers it too, and the count is what this
// placement added, not the creature's total.
func TestB39ShalaiAndHallarCountsCountersAnOpponentPuts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Shalai and Hallar", "Legendary Creature — Angel Elf", effectstest.B39ShalaiAndHallarOracle, false)
	bear := effectstest.SeedBear(g, me.ID)
	before := opp.Life
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventResolve, Actor: opp.ID})
		_ = g.AddCounterByForEffect(opp.ID, bear, game.CounterPlusOne, 2)
	})
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-2 {
		t.Fatalf("two counters an opponent put on your creature are two damage: %d → %d", before, opp.Life)
	}
	// A second placement of one more is one damage, not three.
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventResolve, Actor: opp.ID})
		_ = g.AddCounterForEffect(bear, game.CounterPlusOne, 1)
	})
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-3 {
		t.Errorf("one more counter is one more damage: %d → %d", before, opp.Life)
	}
}

// #3863: counters a creature enters with are "put on" it (CR 122.6).
func TestB39ShalaiAndHallarCountsCountersACreatureEntersWith(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Shalai and Hallar", "Legendary Creature — Angel Elf", effectstest.B39ShalaiAndHallarOracle, false)
	before := opp.Life
	effectstest.CastCatalogSpell(t, g, "Phantom Centaur", "Creature — Centaur Spirit", "9bb8c54b-1228-4b7c-8651-52cb5b0f6e72", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-3 {
		t.Errorf("a creature entering with three counters is three damage: %d → %d", before, opp.Life)
	}
}

// --- Empowered Autogenerator ----------------------------------------

// The counter goes on first, so the first activation adds ONE.
func TestB39AutogeneratorAddsOneOnItsFirstTapAndGrows(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	gen := effectstest.CastCatalogSpell(t, g, "Empowered Autogenerator", "Artifact", effectstest.B39AutogeneratorOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B16Tapped(t, g, gen) {
		t.Fatal("the Autogenerator enters tapped")
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(gen) })
	effectstest.B28TapForMana(t, g, me.ID, gen, "U")
	if got := effectstest.PoolColors(me); len(got) != 1 {
		t.Errorf("the first tap adds exactly one mana: pool %v", got)
	}
	if got := effectstest.CounterOn(g, gen, game.CounterCharge); got != 1 {
		t.Errorf("one charge counter after the first tap: %d", got)
	}

	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(gen) })
	effectstest.B28TapForMana(t, g, me.ID, gen, "U")
	if got := effectstest.PoolColors(me); len(got) != 3 {
		t.Errorf("the second tap adds two more: pool %v", got)
	}
	if got := effectstest.CounterOn(g, gen, game.CounterCharge); got != 2 {
		t.Errorf("two charge counters after the second tap: %d", got)
	}
}

// #1370: with a Doubling Season on the same board, the counter
// placement lands doubled — 2 charge counters on the first tap, not
// 1 — and X has to read that, not guess "existing + 1". Before the
// fix, ProducedFunc computed the pre-doubler guess while the Rider
// placed the doubled count afterward, so the pool got 1 mana for a
// card carrying 2 charge counters.
func TestB39AutogeneratorReadsCounterCountAfterDoublingSeason(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	_ = effectstest.SeedReplacementPermanent(g, effectstest.DoublingSeasonOracle, "Doubling Season", me.ID)
	gen := effectstest.CastCatalogSpell(t, g, "Empowered Autogenerator", "Artifact", effectstest.B39AutogeneratorOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(gen) })

	effectstest.B28TapForMana(t, g, me.ID, gen, "U")
	if got := effectstest.CounterOn(g, gen, game.CounterCharge); got != 2 {
		t.Fatalf("first tap places 1×2 (Doubling Season) = 2 charge counters, got %d", got)
	}
	if got := effectstest.PoolColors(me); len(got) != 2 {
		t.Errorf("X must read the doubled count: pool %v, want 2 mana", got)
	}

	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(gen) })
	effectstest.B28TapForMana(t, g, me.ID, gen, "U")
	if got := effectstest.CounterOn(g, gen, game.CounterCharge); got != 4 {
		t.Fatalf("second tap places 1×2 more = 4 charge counters total, got %d", got)
	}
	if got := effectstest.PoolColors(me); len(got) != 6 {
		t.Errorf("second activation adds 4 more mana on top of the first 2: pool %v, want 6", got)
	}
}

// #1370: Hardened Scales only replaces a +1/+1 counter placed on a
// creature (see hardened_scales.go's AppliesTo). The Autogenerator
// places a CHARGE counter on an ARTIFACT, so Hardened Scales alone on
// the board must be a complete no-op — the card behaves exactly as it
// does with no doubler in play, pinning that the ordering fix did not
// accidentally widen what counts as a doubler here.
func TestB39AutogeneratorUnaffectedByHardenedScalesAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	_ = effectstest.SeedReplacementPermanent(g, effectstest.HardenedScalesOracle, "Hardened Scales", me.ID)
	gen := effectstest.CastCatalogSpell(t, g, "Empowered Autogenerator", "Artifact", effectstest.B39AutogeneratorOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(gen) })

	effectstest.B28TapForMana(t, g, me.ID, gen, "U")
	if got := effectstest.CounterOn(g, gen, game.CounterCharge); got != 1 {
		t.Fatalf("Hardened Scales does not touch a charge counter on an artifact: got %d charge counters, want 1", got)
	}
	if got := effectstest.PoolColors(me); len(got) != 1 {
		t.Errorf("unaffected first tap still adds exactly one mana: pool %v", got)
	}

	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(gen) })
	effectstest.B28TapForMana(t, g, me.ID, gen, "U")
	if got := effectstest.CounterOn(g, gen, game.CounterCharge); got != 2 {
		t.Fatalf("two charge counters after the second tap, want 2, got %d", got)
	}
	if got := effectstest.PoolColors(me); len(got) != 3 {
		t.Errorf("unaffected second tap adds two more: pool %v, want 3 total", got)
	}
}

// --- Anara, Wolvid Familiar -----------------------------------------

func TestB39AnaraGrantsIndestructibleOnlyOnYourOwnTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Anara, Wolvid Familiar", "Legendary Creature — Wolf Beast", effectstest.B39AnaraOracle, false)
	mine := effectstest.B36Commander(g, me.ID, "My Commander")
	theirs := effectstest.B36Commander(g, opp.ID, "Their Commander")
	bear := effectstest.SeedBear(g, me.ID)

	effectstest.B39NextTurnOf(t, g, 0)
	if !effectstest.HasEffectiveKeyword(t, g, mine, "indestructible") {
		t.Error("on your turn your commander is indestructible")
	}
	if effectstest.HasEffectiveKeyword(t, g, theirs, "indestructible") {
		t.Error("\"commanders you control\" does not cover an opponent's")
	}
	if effectstest.HasEffectiveKeyword(t, g, bear, "indestructible") {
		t.Error("a creature that is not a commander gets nothing")
	}

	effectstest.AdvanceToMainOf(t, g, 1)
	if effectstest.HasEffectiveKeyword(t, g, mine, "indestructible") {
		t.Error("on somebody else's turn the grant is gone")
	}
}

// --- Airbending Lesson -----------------------------------------------

func TestB39AirbendingLessonExilesANonlandAndCantrips(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.B39Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	land := effectstest.B12Permanent(g, opp.ID, "Their Wastes", "Basic Land")

	spec, _ := effects.Lookup(effectstest.B39AirbendingLessonOracle)
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
	effectstest.CastCatalogSpell(t, g, "Airbending Lesson", "Instant — Lesson", effectstest.B39AirbendingLessonOracle, effectstest.CardRefs(theirs))
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, me.ID)
	flail := effectstest.SeedEquipment(g, me.ID, "Inquisitor's Flail", effectstest.B39InquisitorsFlailOracle)
	effectstest.EquipTo(t, g, me.ID, flail, bear)

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
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-4 {
		t.Errorf("the equipped creature's 2 power hits for 4: %d → %d, want %d", before, opp.Life, before-4)
	}
}

// Inbound: another creature's combat damage to the equipped creature
// is doubled too, which is what makes the Flail a liability on a
// blocker. A 1/1 blocker kills a Flailed 2/2.
func TestB39InquisitorsFlailDoublesTheDamageComingBack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, me.ID)
	flail := effectstest.SeedEquipment(g, me.ID, "Inquisitor's Flail", effectstest.B39InquisitorsFlailOracle)
	effectstest.EquipTo(t, g, me.ID, flail, bear)
	blocker := effectstest.B39Creature(g, opp.ID, "Their Wall", "Creature — Wall", 1, 4)

	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.B35Block(t, g, blocker, bear)
	effectstest.LockInBlocks(t, g)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)

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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, me.ID)
	blocker := effectstest.B39Creature(g, opp.ID, "Their Wall", "Creature — Wall", 1, 4)

	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.B35Block(t, g, blocker, bear)
	effectstest.LockInBlocks(t, g)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(blocker) || !g.Battlefield.Contains(bear) {
		t.Error("2 into a 1/4 and 1 into a 2/2 kills nothing without the Flail")
	}
}

// --- Dawn of a New Age ------------------------------------------------

func TestB39DawnOfANewAgeDrawsPerCreatureThenCashesIn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedBear(g, me.ID)
	effectstest.B39Creature(g, me.ID, "Second Bear", "Creature — Bear", 2, 2)

	dawn := effectstest.CastCatalogSpell(t, g, "Dawn of a New Age", "Enchantment", effectstest.B39DawnOfANewAgeOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterOn(g, dawn, "hope"); got != 2 {
		t.Fatalf("a hope counter for each of the two creatures: %d", got)
	}
	// A creature arriving afterwards adds nothing — it is an entry
	// replacement, not a static.
	effectstest.B39Creature(g, me.ID, "Third Bear", "Creature — Bear", 2, 2)
	if got := effectstest.CounterOn(g, dawn, "hope"); got != 2 {
		t.Errorf("the count is taken once, as it enters: %d", got)
	}

	hand, life := me.Hand.Size(), me.Life
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterOn(g, dawn, "hope"); got != 1 {
		t.Errorf("one counter comes off at your end step: %d", got)
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("removing a counter draws a card: hand %d → %d", hand, me.Hand.Size())
	}
	if !g.Battlefield.Contains(dawn) {
		t.Fatal("a counter is left, so it is not sacrificed yet")
	}

	effectstest.B39NextTurnOf(t, g, 0)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(dawn) {
		t.Error("the turn that spends the last counter sacrifices it")
	}
	if me.Life != life+4 {
		t.Errorf("and gains 4 life: %d → %d", life, me.Life)
	}
}

// --- Marketback Walker ------------------------------------------------

func TestB39MarketbackWalkerEntersWithXAndDrawsThatManyWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	walker := effectstest.B19CastXModal(t, g, "Marketback Walker", "Artifact Creature — Construct",
		effectstest.B39MarketbackWalkerOracle, "{X}{X}", 3, nil, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterOn(g, walker, game.CounterPlusOne); got != 3 {
		t.Fatalf("cast for X=3, three +1/+1 counters: %d", got)
	}
	if got := effectstest.B39CurrentPower(t, g, walker); got != 3 {
		t.Errorf("a 0/0 with three counters is a 3/3: power %d", got)
	}

	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C", "C", "C", "C")
	effectstest.B16Activate(t, g, me.ID, walker, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterOn(g, walker, game.CounterPlusOne); got != 4 {
		t.Fatalf("{4} adds a fourth counter: %d", got)
	}

	hand := me.Hand.Size()
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(walker) })
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+4 {
		t.Errorf("it draws a card for each counter it had: hand %d → %d, want %d", hand, me.Hand.Size(), hand+4)
	}
}

// --- planeswalkers ----------------------------------------------------

func TestB39AjaniGrantsVigilanceAndTicksTheOtherWalkers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ajani := effectstest.PushCatalogPermanent(g, me.ID, "Ajani, the Greathearted", "Legendary Planeswalker — Ajani", effectstest.B39AjaniOracle, false)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(ajani, game.CounterLoyalty, 5) })
	bear := effectstest.SeedBear(g, me.ID)
	theirs := effectstest.B39Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	other := effectstest.PushCatalogPermanent(g, me.ID, "Other Walker", "Legendary Planeswalker — Teferi", "", false)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(other, game.CounterLoyalty, 3) })

	if !effectstest.HasEffectiveKeyword(t, g, bear, "vigilance") {
		t.Error("creatures you control have vigilance")
	}
	if effectstest.HasEffectiveKeyword(t, g, theirs, "vigilance") {
		t.Error("an opponent's creature gets nothing")
	}

	effectstest.AdvanceToMain(t, g)
	life := me.Life
	effectstest.B16Activate(t, g, me.ID, ajani, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("+1 gains 3 life: %d → %d", life, me.Life)
	}

	effectstest.B39NextTurnOf(t, g, 0)
	effectstest.B16Activate(t, g, me.ID, ajani, 1, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterOn(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("a +1/+1 counter on each creature you control: %d", got)
	}
	if got := effectstest.CounterOn(g, theirs, game.CounterPlusOne); got != 0 {
		t.Errorf("not on an opponent's: %d", got)
	}
	if got := effectstest.CounterOn(g, other, game.CounterLoyalty); got != 4 {
		t.Errorf("a loyalty counter on each OTHER walker you control: %d, want 4", got)
	}
}

func TestB39ObNixilisBurnsEveryOpponentDrawAndHelpsTheVictim(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ob := effectstest.PushCatalogPermanent(g, me.ID, "Ob Nixilis, the Hate-Twisted", "Legendary Planeswalker — Nixilis", effectstest.B39ObNixilisOracle, false)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(ob, game.CounterLoyalty, 5) })
	theirs := effectstest.B39Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	// Your own draws are free.
	mine := me.Life
	hand := me.Hand.Size()
	g.WithWriteLock(func() { _ = g.DrawNForEffect(me.ID, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != mine {
		t.Errorf("Ob Nixilis does not burn his own controller: %d → %d", mine, me.Life)
	}

	effectstest.AdvanceToMain(t, g)
	before := opp.Life
	oppHand := opp.Hand.Size()
	effectstest.B16Activate(t, g, me.ID, ob, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(theirs)})
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sorin := effectstest.PushCatalogPermanent(g, me.ID, "Sorin Markov", "Legendary Planeswalker — Sorin", effectstest.B39SorinMarkovOracle, false)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(sorin, game.CounterLoyalty, 10) })

	effectstest.AdvanceToMain(t, g)
	opp.Life = 38
	effectstest.B16Activate(t, g, me.ID, sorin, 1, game.ActivateAbilityParams{Targets: b39PlayerRefs(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != 10 {
		t.Errorf("38 becomes 10: %d", opp.Life)
	}

	effectstest.B39NextTurnOf(t, g, 0)
	opp.Life = 4
	effectstest.B16Activate(t, g, me.ID, sorin, 1, game.ActivateAbilityParams{Targets: b39PlayerRefs(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != 10 {
		t.Errorf("4 becomes 10 — the clause is a set, not a loss: %d", opp.Life)
	}

	spec, _ := effects.Lookup(effectstest.B39SorinMarkovOracle)
	if len(spec.Activated) != 3 {
		t.Errorf("Sorin offers +2, -3 and the -7 turn control (#3857): %d abilities", len(spec.Activated))
	}
}

// --- spells -----------------------------------------------------------

func TestB39ThrillingDiscoveryGainsTwoThenTradesTwoForThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	life, hand := me.Life, me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Thrilling Discovery", "Sorcery", effectstest.B39ThrillingDiscoveryOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("the 2 life is unconditional: %d → %d", life, me.Life)
	}
	if got := effectstest.DiscardOwed(g, me.ID); got != 2 {
		t.Fatalf("the prompt offers up to two: owes %d", got)
	}
	effectstest.AnswerDiscard(t, g, me.ID, effectstest.B39HandIDs(me, 2)...)
	effectstest.PassPriorityAroundTable(t, g)
	// The helper seeded and cast the sorcery (net zero), then −2
	// discarded and +3 drawn.
	if me.Hand.Size() != hand+1 {
		t.Errorf("two for three: hand %d → %d, want %d", hand, me.Hand.Size(), hand+1)
	}
}

// Declining the discard is legal and draws nothing — the "if you do"
// is a real condition.
func TestB39ThrillingDiscoveryDeclinedStillGainsTheTwoLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	life, hand := me.Life, me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Thrilling Discovery", "Sorcery", effectstest.B39ThrillingDiscoveryOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerDiscard(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("the life is unconditional: %d → %d", life, me.Life)
	}
	if me.Hand.Size() != hand {
		t.Errorf("declining draws nothing: hand %d → %d, want %d", hand, me.Hand.Size(), hand)
	}
}

func TestB39CatharticPyreTakesEitherModeAndNeverAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	theirs := effectstest.B39Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	spec, _ := effects.Lookup(effectstest.B39CatharticPyreOracle)
	if spec.Modes == nil || len(spec.Modes.Options) != 2 || spec.Modes.Min != 1 || spec.Modes.Max != 1 {
		t.Fatalf("Cathartic Pyre is choose one of two: %+v", spec.Modes)
	}
	if spec.Modes.Options[0].Targets == nil || spec.Modes.Options[0].Targets.Players {
		t.Error("the damage mode targets a creature or planeswalker, never a player")
	}

	before := opp.Life
	effectstest.B19CastXModal(t, g, "Cathartic Pyre", "Instant", effectstest.B39CatharticPyreOracle, "{1}{R}", 0, []int{0}, effectstest.CardRefs(theirs))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("3 damage kills a 2/2")
	}
	if opp.Life != before {
		t.Errorf("the mode cannot hit a player: %d → %d", before, opp.Life)
	}
}

func TestB39BanefireDealsXAndDeclaresItsRiders(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	before := opp.Life
	effectstest.B19CastXModal(t, g, "Banefire", "Sorcery", effectstest.B39BanefireOracle, "{X}{R}", 6, nil, b39PlayerRefs(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-6 {
		t.Errorf("X damage to any target: %d → %d, want %d", before, opp.Life, before-6)
	}

	spec, _ := effects.Lookup(effectstest.B39BanefireOracle)
	if spec.CantBeCountered {
		t.Error("the uncounterable rider is conditional on X; the unconditional flag would make a small Banefire uncounterable, which is stronger than printed")
	}
	// ADR 0107 §5: both riders are conditional spell riders now.
	if spec.CantBeCounteredIf == nil || spec.SpellDamageCantBePrevented == nil || spec.Completeness != effects.CompletenessFull {
		t.Errorf("both riders are declared on X: %v %v", spec.Completeness, spec.Caveats)
	}
	_ = me
}

// --- Nested Shambler ---------------------------------------------------

// X is the LAST-KNOWN power, so counters and anthems count.
func TestB39NestedShamblerMakesTappedSquirrelsForItsLastKnownPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	shambler := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Nested Shambler", TypeLine: "Creature — Zombie",
		OracleID: effectstest.B39NestedShamblerOracle, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(shambler, game.CounterPlusOne, 2) })
	if got := effectstest.B39CurrentPower(t, g, shambler); got != 3 {
		t.Fatalf("setup: a 1/1 with two counters is a 3/3, got %d", got)
	}

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(shambler) })
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	erebos := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Erebos, Bleak-Hearted",
		TypeLine: "Legendary Enchantment Creature — God", OracleID: effectstest.B39ErebosOracle,
		ManaCost: "{3}{B}", Power: 5, Toughness: 6, Owner: me.ID, Controller: me.ID,
	})
	// His own {B} is devotion 1 — well short of five.
	if effectstest.B39IsCreature(t, g, erebos) {
		t.Error("devotion 1 is less than five: Erebos isn't a creature")
	}
	for i := 0; i < 4; i++ {
		effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: "Black Rock", TypeLine: "Artifact",
			ManaCost: "{B}", Owner: me.ID, Controller: me.ID,
		})
	}
	if !effectstest.B39IsCreature(t, g, erebos) {
		t.Error("devotion 5 turns him on")
	}
	if !effectstest.HasEffectiveKeyword(t, g, erebos, "indestructible") {
		t.Error("indestructible is printed and is live either way")
	}
}

func TestB39ErebosDrawsForTwoLifeAndShrinksForASacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	erebos := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Erebos, Bleak-Hearted",
		TypeLine: "Legendary Enchantment Creature — God", OracleID: effectstest.B39ErebosOracle,
		ManaCost: "{3}{B}", Power: 5, Toughness: 6, Owner: me.ID, Controller: me.ID,
	})
	bear := effectstest.SeedBear(g, me.ID)
	fodder := effectstest.B39Creature(g, me.ID, "Fodder", "Creature — Goblin", 1, 1)

	hand, life := me.Hand.Size(), me.Life
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(fodder) })
	effectstest.PassPriorityAroundTable(t, g)
	// The decision is made as the trigger resolves, not as it is put on
	// the stack: nothing was paid or asked until now.
	if me.Life != life {
		t.Fatalf("life %d → %d before the trigger resolved", life, me.Life)
	}
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 || me.Life != life-2 {
		t.Errorf("another creature dying pays 2 life for a card: hand %d → %d, life %d → %d",
			hand, me.Hand.Size(), life, me.Life)
	}

	effectstest.AdvanceToMain(t, g)
	victim := effectstest.B39Creature(g, me.ID, "Second Fodder", "Creature — Goblin", 1, 1)
	effectstest.B06AddMana(me, "C", "B")
	effectstest.B16Activate(t, g, me.ID, erebos, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{victim},
		Targets:      effectstest.CardRefs(bear),
	})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, bear); got != 0 {
		t.Errorf("-2/-1 on a 2/2 leaves a 0/1: power %d", got)
	}
	if got := effectstest.EffectiveToughness(t, g, bear); got != 1 {
		t.Errorf("-2/-1 on a 2/2 leaves a 0/1: toughness %d", got)
	}
}
