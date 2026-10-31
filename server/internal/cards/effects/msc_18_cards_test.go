package effects

import (
	"errors"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// msc_18_cards_test.go — Marvel Super Heroes Commander, A–L, batch 18 (#4455).
// Each test plays the real catalog entry and asserts what its text says.

const (
	msc18ArchnemesisOracle   = "c408717c-e07b-484f-9c81-5666595adf81"
	msc18CrucibleOracle      = "2469c0b7-a045-4733-bf53-28f3fdbe3797"
	msc18ExtractPowerOracle  = "5c30a92b-2e88-4e66-b0a5-fa05304136e7"
	msc18FolkHeroOracle      = "c5126d6e-969c-4034-98e3-f8653163346f"
	msc18HangoutOracle       = "43306a78-85b5-4df4-a018-75dd75aa8997"
	msc18DonaldBlakeOracle   = "f966ae40-4717-4f41-9437-0e38b01a8be9"
	msc18AlexWilderOracle    = "f3031397-c2d5-40c6-8eca-88faed00a5ea"
	msc18DivinationOracle    = "273b339c-964b-4a18-8eb5-ceb8abcdfd9e"
	msc18VoraciousTyphonOrcl = "f62784c0-9c67-4a05-a09b-aabf03a9390f"
)

func TestMsc18EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		msc18ArchnemesisOracle:  "Archnemesis",
		msc18CrucibleOracle:     "Cosmic Crucible",
		msc18ExtractPowerOracle: "Extract Power",
		msc18FolkHeroOracle:     "Folk Hero",
		msc18HangoutOracle:      "Heroes' Hangout",
		msc18DonaldBlakeOracle:  "Donald Blake, Guise of Thor",
		msc18AlexWilderOracle:   "Alex Wilder, Runaway",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- Donald Blake, Guise of Thor ---

func TestMsc18DonaldBlakePowerUpMakesHimAFlyingGodWarriorHero(t *testing.T) {
	g := newCatalogGame(t)
	powerUpMain(t, g)
	me := g.Seats[0]
	donald := pushPowerUpPermanent(g, me.ID, "Donald Blake, Guise of Thor", "Legendary Creature — Human Doctor", "{1}{W}", msc18DonaldBlakeOracle)
	if !slices.Contains(effectiveAbilities(t, g, donald), "lifelink") {
		t.Fatalf("Donald has no lifelink: %v", effectiveAbilities(t, g, donald))
	}
	if slices.Contains(effectiveAbilities(t, g, donald), "flying") {
		t.Fatal("Donald flies before the power-up")
	}
	if got := chargedPowerUp(t, g, me.ID, donald); got != "{4}{W}{W}" {
		t.Fatalf("price = %q, want {4}{W}{W}", got)
	}
	floatMana(t, g, me, "{C}{C}{C}{C}{W}{W}")
	if err := g.ActivateCatalogAbility(me.ID, donald, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("power-up: %v", err)
	}
	resolveTopOfStack(t, g)

	if got := plusOnes(g, donald); got != 2 {
		t.Errorf("+1/+1 counters = %d, want 2", got)
	}
	if got := effectivePowerWithCounters(t, g, donald); got != 4 {
		t.Errorf("power = %d, want 4 (2 base + 2 counters)", got)
	}
	abilities := effectiveAbilities(t, g, donald)
	if !slices.Contains(abilities, "flying") || !slices.Contains(abilities, "lifelink") {
		t.Errorf("abilities = %v, want flying (the counter) and lifelink", abilities)
	}
	subs := effectiveSubtypes(t, g, donald)
	slices.Sort(subs)
	if !slices.Equal(subs, []string{"God", "Hero", "Warrior"}) {
		t.Errorf("subtypes = %v, want exactly God Warrior Hero (Human and Doctor are lost)", subs)
	}
	floatMana(t, g, me, "{C}{C}{C}{C}{W}{W}")
	if err := g.ActivateCatalogAbility(me.ID, donald, 0, game.ActivateAbilityParams{}); !errors.Is(err, game.ErrPowerUpUsed) {
		t.Errorf("second power-up: err = %v, want ErrPowerUpUsed", err)
	}
}

func TestMsc18DonaldBlakePowerUpIsCheaperTheTurnHeEntered(t *testing.T) {
	g := newCatalogGame(t)
	powerUpMain(t, g)
	me := g.Seats[0]
	donald := castPowerUpCreature(t, g, me, "Donald Blake, Guise of Thor", "Legendary Creature — Human Doctor", "{1}{W}", "{C}{W}", msc18DonaldBlakeOracle, 1)
	if got := chargedPowerUp(t, g, me.ID, donald); got != "{3}{W}" {
		t.Errorf("price the turn he entered = %q, want {3}{W} ({4}{W}{W} less {1}{W})", got)
	}
}

// --- Alex Wilder, Runaway ---

func msc18GraveyardCreature(g *game.Game, owner *game.Player, name, typeLine, oracle string, p, tough int) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		owner.Graveyard.PushTop(game.Card{
			InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
			Power: p, Toughness: tough, Owner: owner.ID, Controller: owner.ID,
		})
	})
	return id
}

func msc18Escape(t *testing.T, g *game.Game, me *game.Player, id uuid.UUID, fodder int) {
	t.Helper()
	pay := seedGraveyardFodder(t, g, fodder)
	floatMana(t, g, me, "{C}{C}{R}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: "escape", AltCostIDs: pay, Strict: true,
	}); err != nil {
		t.Fatalf("escape cast: %v", err)
	}
	passPriorityAroundTable(t, g)
}

func TestMsc18AlexWilderEscapedPumpsAndHastesHimself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	alex := msc18GraveyardCreature(g, me, "Alex Wilder, Runaway", "Legendary Creature — Human Hero Villain", msc18AlexWilderOracle, 1, 3)
	msc18Escape(t, g, me, alex, 3)
	if !g.Battlefield.Contains(alex) {
		t.Fatal("Alex did not escape onto the battlefield")
	}
	if got := effectivePower(t, g, alex); got != 3 {
		t.Errorf("power = %d, want 3 (1 + 2)", got)
	}
	if !slices.Contains(effectiveAbilities(t, g, alex), "haste") {
		t.Errorf("Alex has no haste: %v", effectiveAbilities(t, g, alex))
	}
}

func TestMsc18AlexWilderPumpsAnotherCreatureYouCastFromTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	alex := b12Push(g, me.ID, "Alex Wilder, Runaway", "Legendary Creature — Human Hero Villain", msc18AlexWilderOracle, 1, 3)
	typhon := msc18GraveyardCreature(g, me, "Voracious Typhon", "Creature — Snake Beast", msc18VoraciousTyphonOrcl, 4, 4)
	pay := seedGraveyardFodder(t, g, 4)
	floatMana(t, g, me, "{C}{C}{B}{B}{B}")
	if err := g.CastSpell(me.ID, typhon, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: "escape", AltCostIDs: pay,
	}); err != nil {
		t.Fatalf("escape the Typhon: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(typhon) {
		t.Fatal("the Typhon did not enter")
	}
	if got, base := effectivePowerWithCounters(t, g, typhon), 4+3+2; got != base {
		t.Errorf("Typhon power = %d, want %d (4 + three counters + 2)", got, base)
	}
	if !slices.Contains(effectiveAbilities(t, g, typhon), "haste") {
		t.Errorf("the Typhon has no haste: %v", effectiveAbilities(t, g, typhon))
	}
	if got := effectivePower(t, g, alex); got != 1 {
		t.Errorf("Alex's own power = %d, want 1: he did not enter", got)
	}
}

func TestMsc18AlexWilderIgnoresACreatureCastFromYourHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Alex Wilder, Runaway", "Legendary Creature — Human Hero Villain", msc18AlexWilderOracle, 1, 3)
	bear := castHoldCreature(t, g, "Grizzly Bears", "Creature — Bear", "", 2, 2)
	passPriorityAroundTable(t, g)
	if got := effectivePower(t, g, bear); got != 2 {
		t.Errorf("hand-cast Bear power = %d, want 2", got)
	}
	if slices.Contains(effectiveAbilities(t, g, bear), "haste") {
		t.Error("the hand-cast Bear got haste")
	}
}

func TestMsc18AlexWilderIgnoresACreatureThatWasNotCast(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Alex Wilder, Runaway", "Legendary Creature — Human Hero Villain", msc18AlexWilderOracle, 1, 3)
	corpse := msc18GraveyardCreature(g, me, "Corpse", "Creature — Zombie", "", 2, 2)
	g.WithWriteLock(func() {
		if err := g.ReturnFromGraveyardForEffect(corpse, game.ZoneBattlefield); err != nil {
			t.Errorf("reanimate: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(corpse) {
		t.Fatal("the creature did not return")
	}
	if got := effectivePower(t, g, corpse); got != 2 {
		t.Errorf("reanimated power = %d, want 2: it was never cast", got)
	}
}

func TestMsc18AlexWilderEscapeNeedsThreeOtherCards(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	alex := msc18GraveyardCreature(g, me, "Alex Wilder, Runaway", "Legendary Creature — Human Hero Villain", msc18AlexWilderOracle, 1, 3)
	pay := seedGraveyardFodder(t, g, 2)
	floatMana(t, g, me, "{C}{C}{R}")
	err := g.CastSpell(me.ID, alex, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: "escape", AltCostIDs: pay, Strict: true,
	})
	if err == nil {
		t.Fatal("escaped with only two other cards")
	}
	if g.Battlefield.Contains(alex) {
		t.Error("Alex is on the battlefield")
	}
}

// --- Cosmic Crucible ---

// msc18SettleCrucibleMana walks to the first main phase and answers the
// Crucible's four colour picks, so the stack is empty for a sorcery.
func msc18SettleCrucibleMana(t *testing.T, g *game.Game, me *game.Player) {
	t.Helper()
	powerUpMain(t, g)
	passPriorityAroundTable(t, g)
	if n := riderAnswerManaPicks(t, g, me.ID, "W"); n != 4 {
		t.Fatalf("answered %d colour picks, want 4", n)
	}
}

func msc18DivinationCast(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	return castCatalogSpell(t, g, "Divination", "Sorcery", msc18DivinationOracle, nil)
}

func msc18Library(me *game.Player, n int) {
	names := make([]string, n)
	for i := range names {
		names[i] = "Library Card"
	}
	seedLibrary(me, names...)
}

func TestMsc18CosmicCrucibleAddsFourManaInAnyCombinationAtYourFirstMain(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Cosmic Crucible", "Enchantment", msc18CrucibleOracle, 0, 0)
	powerUpMain(t, g)
	passPriorityAroundTable(t, g)
	first := riderLatestManaPick(g, me.ID)
	if first == nil {
		t.Fatal("no colour pick for the Crucible's mana")
	}
	if len(first.ColorOptions) != 5 {
		t.Errorf("colour options %v, want all five", first.ColorOptions)
	}
	if err := g.ResolveManaChoice(first.ID, me.ID, "W"); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
	// The remaining three picks differ: independent slots.
	for _, c := range []string{"U", "B", "R"} {
		pick := riderLatestManaPick(g, me.ID)
		if pick == nil {
			t.Fatalf("missing a colour pick for %s", c)
		}
		if err := g.ResolveManaChoice(pick.ID, me.ID, c); err != nil {
			t.Fatalf("ResolveManaChoice: %v", err)
		}
	}
	if me.ManaPool.Total() != 4 {
		t.Fatalf("pool has %d mana, want 4", me.ManaPool.Total())
	}
	for _, c := range []string{"W", "U", "B", "R"} {
		if msc12PoolOf(me, c) != 1 {
			t.Errorf("pool has %d %s, want 1", msc12PoolOf(me, c), c)
		}
	}
}

func TestMsc18CosmicCrucibleAddsNothingOnAnOpponentsTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Cosmic Crucible", "Enchantment", msc18CrucibleOracle, 0, 0)
	msc18SettleCrucibleMana(t, g, me)
	advanceToNextSeatsTurn(t, g)
	powerUpMain(t, g)
	passPriorityAroundTable(t, g)
	if riderLatestManaPick(g, me.ID) != nil || me.ManaPool.Total() != 0 {
		t.Errorf("mana on an opponent's turn: pool %d", me.ManaPool.Total())
	}
}

func TestMsc18CosmicCrucibleCopiesANoncreatureSpellOnceEachTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	msc18Library(me, 12)
	b12Push(g, me.ID, "Cosmic Crucible", "Enchantment", msc18CrucibleOracle, 0, 0)
	msc18SettleCrucibleMana(t, g, me)
	hand := me.Hand.Size()

	msc18DivinationCast(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 4 {
		t.Fatalf("drew %d cards, want 4: the Divination and its copy", got)
	}

	// The second noncreature spell this turn is not copied.
	msc18DivinationCast(t, g)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 6 {
		t.Errorf("after a second Divination drew %d cards in all, want 6 (no second copy)", got)
	}
}

func TestMsc18CosmicCrucibleDecliningDoesNotUseTheTurnUp(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	msc18Library(me, 12)
	b12Push(g, me.ID, "Cosmic Crucible", "Enchantment", msc18CrucibleOracle, 0, 0)
	msc18SettleCrucibleMana(t, g, me)
	hand := me.Hand.Size()

	msc18DivinationCast(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Fatalf("drew %d, want 2: declined", got)
	}
	msc18DivinationCast(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 6 {
		t.Errorf("drew %d in all, want 6: the second spell is still offered", got)
	}
}

func TestMsc18CosmicCrucibleIgnoresCreatureSpellsAndOtherPlayers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Cosmic Crucible", "Enchantment", msc18CrucibleOracle, 0, 0)
	msc18SettleCrucibleMana(t, g, me)
	castHoldCreature(t, g, "Grizzly Bears", "Creature — Bear", "", 2, 2)
	if p := latestTriggerPrompt(g, me.ID); p != nil {
		t.Fatal("a creature spell offered a copy")
	}
	passPriorityAroundTable(t, g)
	msc18Library(opp, 6)
	advanceToNextSeatsTurn(t, g)
	castFor(t, g, opp, "Divination", "Sorcery", msc18DivinationOracle, game.CastSpellParams{})
	if p := latestTriggerPrompt(g, me.ID); p != nil {
		t.Fatal("an opponent's spell offered a copy")
	}
}

// --- Folk Hero ---

func msc18Commander(g *game.Game, owner *game.Player, typeLine string) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id, Name: "Test Commander", TypeLine: typeLine,
		Power: 2, Toughness: 2, Owner: owner.ID, Controller: owner.ID, IsCommander: true,
	})
	return id
}

func msc18CastCreature(t *testing.T, g *game.Game, p *game.Player, typeLine string) {
	t.Helper()
	castFor(t, g, p, "Spell "+typeLine, typeLine, "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
}

func TestMsc18FolkHeroGrantsTheDrawToYourCommanderOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	msc18Library(me, 10)
	b12Push(g, me.ID, "Folk Hero", "Legendary Enchantment — Background", msc18FolkHeroOracle, 0, 0)
	cmdr := msc18Commander(g, me, "Legendary Creature — Human Warrior")
	hand := me.Hand.Size()

	msc18CastCreature(t, g, me, "Creature — Human Cleric")
	if got := me.Hand.Size() - hand; got != 1 {
		t.Fatalf("a spell sharing Human drew %d, want 1", got)
	}
	// Only once each turn.
	msc18CastCreature(t, g, me, "Creature — Warrior")
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("a second sharing spell drew again: %d in all, want 1", got)
	}
	_ = cmdr
}

func TestMsc18FolkHeroNeedsASharedCreatureType(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	msc18Library(me, 10)
	b12Push(g, me.ID, "Folk Hero", "Legendary Enchantment — Background", msc18FolkHeroOracle, 0, 0)
	msc18Commander(g, me, "Legendary Creature — Human Warrior")
	hand := me.Hand.Size()

	msc18CastCreature(t, g, me, "Creature — Elf")
	if got := me.Hand.Size() - hand; got != 0 {
		t.Errorf("a spell sharing nothing drew %d", got)
	}
	// The unshared spell did not use the turn's trigger up.
	msc18CastCreature(t, g, me, "Creature — Warrior")
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("a Warrior spell drew %d, want 1", got)
	}
}

func TestMsc18FolkHeroGivesNothingToANonCommanderOrAnOpponentsSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	msc18Library(me, 10)
	msc18Library(opp, 10)
	b12Push(g, me.ID, "Folk Hero", "Legendary Enchantment — Background", msc18FolkHeroOracle, 0, 0)
	b12Creature(g, me.ID, "Plain Human", "Creature — Human Warrior", 2, 2)
	hand := me.Hand.Size()

	msc18CastCreature(t, g, me, "Creature — Human")
	if got := me.Hand.Size() - hand; got != 0 {
		t.Errorf("a non-commander creature drew %d", got)
	}
	msc18Commander(g, me, "Legendary Creature — Human")
	advanceToNextSeatsTurn(t, g)
	powerUpMain(t, g)
	oppHand := opp.Hand.Size()
	hand = me.Hand.Size()
	castFor(t, g, opp, "Their Human", "Creature — Human", "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if me.Hand.Size()-hand != 0 || opp.Hand.Size()-oppHand != 0 {
		t.Errorf("an opponent's spell drew: me %d, opp %d", me.Hand.Size()-hand, opp.Hand.Size()-oppHand)
	}
}

// --- Archnemesis ---

func msc18CastArchnemesis(t *testing.T, g *game.Game, me *game.Player, host uuid.UUID) uuid.UUID {
	t.Helper()
	id := castCatalogSpell(t, g, "Archnemesis", "Enchantment — Aura", msc18ArchnemesisOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: host}})
	passPriorityAroundTable(t, g)
	c, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatal("Archnemesis did not resolve")
	}
	if c.AttachedTo.Kind != game.TargetPlayer || c.AttachedTo.ID != host {
		t.Fatalf("attached to %+v, want player %s", c.AttachedTo, host)
	}
	return id
}

func TestMsc18ArchnemesisEnchantsOnlyAnOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	err := castCatalogSpellErr(t, g, "Archnemesis", "Enchantment — Aura", msc18ArchnemesisOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	if err == nil {
		t.Fatal("Archnemesis was cast on its own controller")
	}
}

func TestMsc18ArchnemesisDrainsDrawsAndGainsWhenYouAttackTheEnchantedPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	msc18Library(me, 6)
	msc18CastArchnemesis(t, g, me, opp.ID)
	a := b12Creature(g, me.ID, "Attacker A", "Creature — Soldier", 2, 2)
	b := b12Creature(g, me.ID, "Attacker B", "Creature — Soldier", 2, 2)
	life, oppLife, hand := me.Life, opp.Life, me.Hand.Size()

	declareAttack(t, g, opp.ID, a, b)
	passPriorityAroundTable(t, g)

	if got := oppLife - opp.Life; got != 2 {
		t.Errorf("the enchanted player lost %d, want 2 (once for two attackers)", got)
	}
	if got := me.Life - life; got != 2 {
		t.Errorf("gained %d life, want 2", got)
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d cards, want 1", got)
	}
}

func TestMsc18ArchnemesisIgnoresAnAttackOnAnotherPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	msc18Library(me, 6)
	msc18CastArchnemesis(t, g, me, opp.ID)
	a := b12Creature(g, me.ID, "Attacker", "Creature — Soldier", 2, 2)
	life, otherLife, hand := me.Life, other.Life, me.Hand.Size()

	declareAttack(t, g, other.ID, a)
	passPriorityAroundTable(t, g)

	if other.Life != otherLife || me.Life != life || me.Hand.Size() != hand {
		t.Errorf("an attack on a player who is not enchanted triggered: other %d→%d, me %d→%d, hand %d→%d",
			otherLife, other.Life, life, me.Life, hand, me.Hand.Size())
	}
}

func TestMsc18ArchnemesisMovesToAPlayerWhoAttacksYou(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	aura := msc18CastArchnemesis(t, g, me, other.ID)
	for g.Turn.ActiveSeat != 1 {
		advanceToNextSeatsTurn(t, g)
	}
	their := b12Creature(g, opp.ID, "Their Attacker", "Creature — Soldier", 2, 2)
	declareAttack(t, g, me.ID, their)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)

	c, ok := battlefieldCard(g, aura)
	if !ok {
		t.Fatal("Archnemesis left the battlefield")
	}
	if c.AttachedTo.Kind != game.TargetPlayer || c.AttachedTo.ID != opp.ID {
		t.Errorf("attached to %+v, want the attacking player %s", c.AttachedTo, opp.ID)
	}
}

func TestMsc18ArchnemesisStaysWhenYouDeclineTheMove(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	aura := msc18CastArchnemesis(t, g, me, other.ID)
	for g.Turn.ActiveSeat != 1 {
		advanceToNextSeatsTurn(t, g)
	}
	their := b12Creature(g, opp.ID, "Their Attacker", "Creature — Soldier", 2, 2)
	declareAttack(t, g, me.ID, their)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)

	c, _ := battlefieldCard(g, aura)
	if c.AttachedTo.ID != other.ID {
		t.Errorf("attached to %+v, want it to stay on %s", c.AttachedTo, other.ID)
	}
}

// --- Extract Power ---

func TestMsc18ExtractPowerExilesEachTopCardFaceDownAndLetsYouPlayThemFree(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	mine := seedSearchLibrary(me, game.Card{Name: "My Bear", TypeLine: "Creature — Bear", ManaCost: "{2}{G}", Power: 2, Toughness: 2})
	theirs := seedSearchLibrary(opp, game.Card{Name: "Their Forest", TypeLine: "Basic Land — Forest"})
	third := g.Seats[2]
	thirdCard := seedSearchLibrary(third, game.Card{Name: "Third Bolt", TypeLine: "Instant", ManaCost: "{R}"})
	castCatalogSpell(t, g, "Extract Power", "Sorcery", msc18ExtractPowerOracle, nil)
	passPriorityAroundTable(t, g)

	for _, id := range []uuid.UUID{mine[0], theirs[0], thirdCard[0]} {
		if !g.Exile.Contains(id) {
			t.Fatalf("top card %s was not exiled", id)
		}
		c, _ := g.LookupCardForEffect(id)
		if c.FaceDownKind == game.FaceDownNone {
			t.Errorf("%s was exiled face up", c.Name)
		}
	}
	// A creature from my own library, for free: no mana floated.
	if me.ManaPool.Total() != 0 {
		t.Fatal("pool is not empty")
	}
	if err := g.CastSpell(me.ID, mine[0], game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("cast my exiled Bear for free: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(mine[0]) {
		t.Error("the Bear did not enter")
	}
	// An opponent's land: played, with a land drop.
	if err := g.CastSpell(me.ID, theirs[0], game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("play the opponent's exiled land: %v", err)
	}
	if !g.Battlefield.Contains(theirs[0]) {
		t.Error("their Forest did not enter")
	}
}

func TestMsc18ExtractPowerGivesTheOpponentsNoPermission(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	mine := seedSearchLibrary(me, game.Card{Name: "My Bear", TypeLine: "Creature — Bear", ManaCost: "{2}{G}", Power: 2, Toughness: 2})
	castCatalogSpell(t, g, "Extract Power", "Sorcery", msc18ExtractPowerOracle, nil)
	passPriorityAroundTable(t, g)
	if err := g.CastSpell(opp.ID, mine[0], game.CastSpellParams{FromZone: "exile", Strict: true}); err == nil {
		t.Error("an opponent cast a card exiled by my Extract Power")
	}
}

// --- Heroes' Hangout ---

func TestMsc18HeroesHangoutDateNightPlaysOnlyTheChosenCardUntilYourNextTurnEnds(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	// seedSearchLibrary puts the FIRST card on top.
	ids := seedSearchLibrary(me,
		game.Card{Name: "Top Card", TypeLine: "Basic Land — Island"},
		game.Card{Name: "Second Card", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Deep Card", TypeLine: "Creature — Bear", Power: 1, Toughness: 1},
	)
	top, second, deep := ids[0], ids[1], ids[2]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Heroes' Hangout", TypeLine: "Sorcery", OracleID: msc18HangoutOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{0}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, c := range []uuid.UUID{top, second} {
		if !g.Exile.Contains(c) {
			t.Fatalf("%s was not exiled", c)
		}
	}
	if !me.Library.Contains(deep) {
		t.Error("the third card left the library")
	}
	answerChooseCards(t, g, me.ID, second)
	passPriorityAroundTable(t, g)

	if err := g.CastSpell(me.ID, top, game.CastSpellParams{FromZone: "exile", Strict: true}); err == nil {
		t.Error("the card that was not chosen can be played")
	}
	if err := g.CastSpell(me.ID, second, game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("play the chosen card: %v", err)
	}
	if !g.Battlefield.Contains(second) {
		t.Error("the chosen land did not enter")
	}
}

func TestMsc18HeroesHangoutDateNightWindowRunsThroughYourNextTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	ids := seedSearchLibrary(me,
		game.Card{Name: "Other Land", TypeLine: "Basic Land — Island"},
		game.Card{Name: "Chosen Land", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Filler", TypeLine: "Creature — Bear", Power: 1, Toughness: 1},
	)
	chosen := ids[1]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Heroes' Hangout", TypeLine: "Sorcery", OracleID: msc18HangoutOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{0}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID, chosen)
	passPriorityAroundTable(t, g)

	// Through another seat's turn, to my next turn: still playable there.
	powerUpPassTo(t, g, 0)
	if err := g.CastSpell(me.ID, chosen, game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("play the chosen card on my next turn: %v", err)
	}
	if !g.Battlefield.Contains(chosen) {
		t.Error("the chosen land did not enter on the next turn")
	}
}

func TestMsc18HeroesHangoutDateNightLapsesAfterYourNextTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	ids := seedSearchLibrary(me,
		game.Card{Name: "Other Land", TypeLine: "Basic Land — Island"},
		game.Card{Name: "Chosen Land", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Filler", TypeLine: "Creature — Bear", Power: 1, Toughness: 1},
	)
	chosen := ids[1]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Heroes' Hangout", TypeLine: "Sorcery", OracleID: msc18HangoutOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{0}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID, chosen)
	passPriorityAroundTable(t, g)

	powerUpPassTo(t, g, 0)
	powerUpPassTo(t, g, 0)
	if err := g.CastSpell(me.ID, chosen, game.CastSpellParams{FromZone: "exile", Strict: true}); err == nil {
		t.Error("the chosen card was still playable the turn after my next turn")
	}
}

func TestMsc18HeroesHangoutPatrolNightPumpsUpToTwoCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	a := b12Creature(g, me.ID, "Hero A", "Creature — Hero", 2, 2)
	b := b12Creature(g, opp.ID, "Their Guy", "Creature — Soldier", 3, 3)
	c := b12Creature(g, me.ID, "Bystander", "Creature — Hero", 2, 2)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Heroes' Hangout", TypeLine: "Sorcery", OracleID: msc18HangoutOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Modes:   []int{1},
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}},
	}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := effectivePower(t, g, a); got != 3 {
		t.Errorf("Hero A power = %d, want 3", got)
	}
	if got := effectivePower(t, g, b); got != 4 {
		t.Errorf("their creature power = %d, want 4", got)
	}
	for _, x := range []uuid.UUID{a, b} {
		if !slices.Contains(effectiveAbilities(t, g, x), "first strike") {
			t.Errorf("%s has no first strike: %v", x, effectiveAbilities(t, g, x))
		}
	}
	if got := effectivePower(t, g, c); got != 2 || slices.Contains(effectiveAbilities(t, g, c), "first strike") {
		t.Errorf("the untargeted creature changed: power %d", got)
	}
}

// --- Doomsday ---

const msc18DoomsdayOracle = "721eb5a2-d7cf-4db0-8013-ef3f596c52a5"

func msc18DoomsdayCast(t *testing.T, g *game.Game) {
	t.Helper()
	castCatalogSpell(t, g, "Doomsday", "Sorcery", msc18DoomsdayOracle, nil)
	passPriorityAroundTable(t, g)
}

func msc18GraveyardCards(me *game.Player, names ...string) []uuid.UUID {
	out := make([]uuid.UUID, 0, len(names))
	for _, n := range names {
		c := game.NewCard(n, me.ID)
		c.TypeLine = "Instant"
		me.Graveyard.PushTop(c)
		out = append(out, c.InstanceID)
	}
	return out
}

func TestMsc18DoomsdayKeepsFiveFromLibraryAndGraveyardInYourOrderAndHalvesYourLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	lib := seedSearchLibrary(me,
		game.Card{Name: "L1", TypeLine: "Instant"}, game.Card{Name: "L2", TypeLine: "Instant"},
		game.Card{Name: "L3", TypeLine: "Instant"}, game.Card{Name: "L4", TypeLine: "Instant"},
		game.Card{Name: "L5", TypeLine: "Instant"}, game.Card{Name: "L6", TypeLine: "Instant"},
	)
	gy := msc18GraveyardCards(me, "G1", "G2", "G3")
	me.Life = 41
	searches := countEvents(g, game.EventSearchLibrary)

	msc18DoomsdayCast(t, g)
	if countEvents(g, game.EventSearchLibrary) != searches+1 {
		t.Error("Doomsday did not search the library")
	}
	keep := []uuid.UUID{lib[0], lib[3], lib[5], gy[0], gy[2]}
	answerChooseCards(t, g, me.ID, keep...)
	order := putInLibraryChoiceFor(g, me.ID)
	if order == nil {
		t.Fatal("no ordering prompt for the five cards")
	}
	want := []uuid.UUID{gy[2], lib[5], gy[0], lib[0], lib[3]}
	if err := g.ResolvePutInLibrary(order.ID, me.ID, nil, want); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	passPriorityAroundTable(t, g)

	if me.Library.Size() != 5 {
		t.Fatalf("library has %d cards, want exactly the 5 chosen", me.Library.Size())
	}
	for i, id := range want {
		// The top of the library is the last element.
		if got := me.Library.Cards[len(me.Library.Cards)-1-i].InstanceID; got != id {
			t.Errorf("library position %d from the top is %s, want %s", i, got, id)
		}
	}
	if me.Graveyard.Size() != 1 { // only Doomsday itself
		t.Errorf("graveyard has %d cards, want only the resolved Doomsday", me.Graveyard.Size())
	}
	for _, id := range []uuid.UUID{lib[1], lib[2], lib[4], gy[1]} {
		if !g.Exile.Contains(id) {
			t.Errorf("an unchosen card %s was not exiled", id)
		}
	}
	if me.Life != 20 {
		t.Errorf("life = %d, want 20 (41 less half rounded up, 21)", me.Life)
	}
}

func TestMsc18DoomsdayFindsAsManyAsPossibleAndRoundsLifeUp(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	lib := seedSearchLibrary(me, game.Card{Name: "L1", TypeLine: "Instant"}, game.Card{Name: "L2", TypeLine: "Instant"})
	gy := msc18GraveyardCards(me, "G1")
	me.Life = 7

	msc18DoomsdayCast(t, g)
	prompt := chooseCardsChoiceFor(g, me.ID)
	if prompt == nil || prompt.ChooseMin != 3 || prompt.ChooseMax != 3 {
		t.Fatalf("prompt = %+v, want exactly the 3 cards that exist", prompt)
	}
	answerChooseCards(t, g, me.ID, lib[0], lib[1], gy[0])
	order := putInLibraryChoiceFor(g, me.ID)
	if order == nil {
		t.Fatal("no ordering prompt")
	}
	if err := g.ResolvePutInLibrary(order.ID, me.ID, nil, []uuid.UUID{lib[0], lib[1], gy[0]}); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Library.Size() != 3 || g.Exile.Size() != 0 {
		t.Errorf("library %d, exile %d, want 3 and 0", me.Library.Size(), g.Exile.Size())
	}
	if me.Life != 3 {
		t.Errorf("life = %d, want 3 (7 less 4)", me.Life)
	}
}

func TestMsc18DoomsdayLeavesOtherPlayersZonesAlone(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	lib := seedSearchLibrary(me, game.Card{Name: "L1", TypeLine: "Instant"})
	theirLib := seedSearchLibrary(opp, game.Card{Name: "T1", TypeLine: "Instant"})
	theirGy := msc18GraveyardCards(opp, "TG")

	msc18DoomsdayCast(t, g)
	answerChooseCards(t, g, me.ID, lib[0])
	passPriorityAroundTable(t, g)
	if g.Exile.Contains(theirLib[0]) || g.Exile.Contains(theirGy[0]) || !opp.Library.Contains(theirLib[0]) {
		t.Error("an opponent's library or graveyard was touched")
	}
}
