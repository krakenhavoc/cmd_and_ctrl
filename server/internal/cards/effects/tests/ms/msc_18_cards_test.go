package testsms

import (
	"errors"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestMsc18EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		effectstest.Msc18ArchnemesisOracle:  "Archnemesis",
		effectstest.Msc18CrucibleOracle:     "Cosmic Crucible",
		effectstest.Msc18ExtractPowerOracle: "Extract Power",
		effectstest.Msc18FolkHeroOracle:     "Folk Hero",
		effectstest.Msc18HangoutOracle:      "Heroes' Hangout",
		effectstest.Msc18DonaldBlakeOracle:  "Donald Blake, Guise of Thor",
		effectstest.Msc18AlexWilderOracle:   "Alex Wilder, Runaway",
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

// --- Donald Blake, Guise of Thor ---

func TestMsc18DonaldBlakePowerUpMakesHimAFlyingGodWarriorHero(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.PowerUpMain(t, g)
	me := g.Seats[0]
	donald := effectstest.PushPowerUpPermanent(g, me.ID, "Donald Blake, Guise of Thor", "Legendary Creature — Human Doctor", "{1}{W}", effectstest.Msc18DonaldBlakeOracle)
	if !slices.Contains(effectstest.EffectiveAbilities(t, g, donald), "lifelink") {
		t.Fatalf("Donald has no lifelink: %v", effectstest.EffectiveAbilities(t, g, donald))
	}
	if slices.Contains(effectstest.EffectiveAbilities(t, g, donald), "flying") {
		t.Fatal("Donald flies before the power-up")
	}
	if got := effectstest.ChargedPowerUp(t, g, me.ID, donald); got != "{4}{W}{W}" {
		t.Fatalf("price = %q, want {4}{W}{W}", got)
	}
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}{W}{W}")
	if err := g.ActivateCatalogAbility(me.ID, donald, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("power-up: %v", err)
	}
	effectstest.ResolveTopOfStack(t, g)

	if got := effectstest.PlusOnes(g, donald); got != 2 {
		t.Errorf("+1/+1 counters = %d, want 2", got)
	}
	if got := effectstest.EffectivePowerWithCounters(t, g, donald); got != 4 {
		t.Errorf("power = %d, want 4 (2 base + 2 counters)", got)
	}
	abilities := effectstest.EffectiveAbilities(t, g, donald)
	if !slices.Contains(abilities, "flying") || !slices.Contains(abilities, "lifelink") {
		t.Errorf("abilities = %v, want flying (the counter) and lifelink", abilities)
	}
	subs := effectstest.EffectiveSubtypes(t, g, donald)
	slices.Sort(subs)
	if !slices.Equal(subs, []string{"God", "Hero", "Warrior"}) {
		t.Errorf("subtypes = %v, want exactly God Warrior Hero (Human and Doctor are lost)", subs)
	}
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}{W}{W}")
	if err := g.ActivateCatalogAbility(me.ID, donald, 0, game.ActivateAbilityParams{}); !errors.Is(err, game.ErrPowerUpUsed) {
		t.Errorf("second power-up: err = %v, want ErrPowerUpUsed", err)
	}
}

func TestMsc18DonaldBlakePowerUpIsCheaperTheTurnHeEntered(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.PowerUpMain(t, g)
	me := g.Seats[0]
	donald := effectstest.CastPowerUpCreature(t, g, me, "Donald Blake, Guise of Thor", "Legendary Creature — Human Doctor", "{1}{W}", "{C}{W}", effectstest.Msc18DonaldBlakeOracle, 1)
	if got := effectstest.ChargedPowerUp(t, g, me.ID, donald); got != "{3}{W}" {
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
	pay := effectstest.SeedGraveyardFodder(t, g, fodder)
	effectstest.FloatMana(t, g, me, "{C}{C}{R}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: "escape", AltCostIDs: pay, Strict: true,
	}); err != nil {
		t.Fatalf("escape cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestMsc18AlexWilderEscapedPumpsAndHastesHimself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	alex := msc18GraveyardCreature(g, me, "Alex Wilder, Runaway", "Legendary Creature — Human Hero Villain", effectstest.Msc18AlexWilderOracle, 1, 3)
	msc18Escape(t, g, me, alex, 3)
	if !g.Battlefield.Contains(alex) {
		t.Fatal("Alex did not escape onto the battlefield")
	}
	if got := effectstest.EffectivePower(t, g, alex); got != 3 {
		t.Errorf("power = %d, want 3 (1 + 2)", got)
	}
	if !slices.Contains(effectstest.EffectiveAbilities(t, g, alex), "haste") {
		t.Errorf("Alex has no haste: %v", effectstest.EffectiveAbilities(t, g, alex))
	}
}

func TestMsc18AlexWilderPumpsAnotherCreatureYouCastFromTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	alex := effectstest.B12Push(g, me.ID, "Alex Wilder, Runaway", "Legendary Creature — Human Hero Villain", effectstest.Msc18AlexWilderOracle, 1, 3)
	typhon := msc18GraveyardCreature(g, me, "Voracious Typhon", "Creature — Snake Beast", effectstest.Msc18VoraciousTyphonOrcl, 4, 4)
	pay := effectstest.SeedGraveyardFodder(t, g, 4)
	effectstest.FloatMana(t, g, me, "{C}{C}{B}{B}{B}")
	if err := g.CastSpell(me.ID, typhon, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: "escape", AltCostIDs: pay,
	}); err != nil {
		t.Fatalf("escape the Typhon: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(typhon) {
		t.Fatal("the Typhon did not enter")
	}
	if got, base := effectstest.EffectivePowerWithCounters(t, g, typhon), 4+3+2; got != base {
		t.Errorf("Typhon power = %d, want %d (4 + three counters + 2)", got, base)
	}
	if !slices.Contains(effectstest.EffectiveAbilities(t, g, typhon), "haste") {
		t.Errorf("the Typhon has no haste: %v", effectstest.EffectiveAbilities(t, g, typhon))
	}
	if got := effectstest.EffectivePower(t, g, alex); got != 1 {
		t.Errorf("Alex's own power = %d, want 1: he did not enter", got)
	}
}

func TestMsc18AlexWilderIgnoresACreatureCastFromYourHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Alex Wilder, Runaway", "Legendary Creature — Human Hero Villain", effectstest.Msc18AlexWilderOracle, 1, 3)
	bear := effectstest.CastHoldCreature(t, g, "Grizzly Bears", "Creature — Bear", "", 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, bear); got != 2 {
		t.Errorf("hand-cast Bear power = %d, want 2", got)
	}
	if slices.Contains(effectstest.EffectiveAbilities(t, g, bear), "haste") {
		t.Error("the hand-cast Bear got haste")
	}
}

func TestMsc18AlexWilderIgnoresACreatureThatWasNotCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Alex Wilder, Runaway", "Legendary Creature — Human Hero Villain", effectstest.Msc18AlexWilderOracle, 1, 3)
	corpse := msc18GraveyardCreature(g, me, "Corpse", "Creature — Zombie", "", 2, 2)
	g.WithWriteLock(func() {
		if err := g.ReturnFromGraveyardForEffect(corpse, game.ZoneBattlefield); err != nil {
			t.Errorf("reanimate: %v", err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(corpse) {
		t.Fatal("the creature did not return")
	}
	if got := effectstest.EffectivePower(t, g, corpse); got != 2 {
		t.Errorf("reanimated power = %d, want 2: it was never cast", got)
	}
}

func TestMsc18AlexWilderEscapeNeedsThreeOtherCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	alex := msc18GraveyardCreature(g, me, "Alex Wilder, Runaway", "Legendary Creature — Human Hero Villain", effectstest.Msc18AlexWilderOracle, 1, 3)
	pay := effectstest.SeedGraveyardFodder(t, g, 2)
	effectstest.FloatMana(t, g, me, "{C}{C}{R}")
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
	effectstest.PowerUpMain(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.RiderAnswerManaPicks(t, g, me.ID, "W"); n != 4 {
		t.Fatalf("answered %d colour picks, want 4", n)
	}
}

func TestMsc18CosmicCrucibleAddsFourManaInAnyCombinationAtYourFirstMain(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Cosmic Crucible", "Enchantment", effectstest.Msc18CrucibleOracle, 0, 0)
	effectstest.PowerUpMain(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	first := effectstest.RiderLatestManaPick(g, me.ID)
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
		pick := effectstest.RiderLatestManaPick(g, me.ID)
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
		if effectstest.Msc12PoolOf(me, c) != 1 {
			t.Errorf("pool has %d %s, want 1", effectstest.Msc12PoolOf(me, c), c)
		}
	}
}

func TestMsc18CosmicCrucibleAddsNothingOnAnOpponentsTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Cosmic Crucible", "Enchantment", effectstest.Msc18CrucibleOracle, 0, 0)
	msc18SettleCrucibleMana(t, g, me)
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.PowerUpMain(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.RiderLatestManaPick(g, me.ID) != nil || me.ManaPool.Total() != 0 {
		t.Errorf("mana on an opponent's turn: pool %d", me.ManaPool.Total())
	}
}

func TestMsc18CosmicCrucibleCopiesANoncreatureSpellOnceEachTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.Msc18Library(me, 12)
	effectstest.B12Push(g, me.ID, "Cosmic Crucible", "Enchantment", effectstest.Msc18CrucibleOracle, 0, 0)
	msc18SettleCrucibleMana(t, g, me)
	hand := me.Hand.Size()

	effectstest.Msc18DivinationCast(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 4 {
		t.Fatalf("drew %d cards, want 4: the Divination and its copy", got)
	}

	// The second noncreature spell this turn is not copied.
	effectstest.Msc18DivinationCast(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 6 {
		t.Errorf("after a second Divination drew %d cards in all, want 6 (no second copy)", got)
	}
}

func TestMsc18CosmicCrucibleDecliningDoesNotUseTheTurnUp(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.Msc18Library(me, 12)
	effectstest.B12Push(g, me.ID, "Cosmic Crucible", "Enchantment", effectstest.Msc18CrucibleOracle, 0, 0)
	msc18SettleCrucibleMana(t, g, me)
	hand := me.Hand.Size()

	effectstest.Msc18DivinationCast(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Fatalf("drew %d, want 2: declined", got)
	}
	effectstest.Msc18DivinationCast(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 6 {
		t.Errorf("drew %d in all, want 6: the second spell is still offered", got)
	}
}

func TestMsc18CosmicCrucibleIgnoresCreatureSpellsAndOtherPlayers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Cosmic Crucible", "Enchantment", effectstest.Msc18CrucibleOracle, 0, 0)
	msc18SettleCrucibleMana(t, g, me)
	effectstest.CastHoldCreature(t, g, "Grizzly Bears", "Creature — Bear", "", 2, 2)
	if p := effectstest.LatestTriggerPrompt(g, me.ID); p != nil {
		t.Fatal("a creature spell offered a copy")
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.Msc18Library(opp, 6)
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.CastFor(t, g, opp, "Divination", "Sorcery", effectstest.Msc18DivinationOracle, game.CastSpellParams{})
	if p := effectstest.LatestTriggerPrompt(g, me.ID); p != nil {
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
	effectstest.CastFor(t, g, p, "Spell "+typeLine, typeLine, "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
}

func TestMsc18FolkHeroGrantsTheDrawToYourCommanderOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.Msc18Library(me, 10)
	effectstest.B12Push(g, me.ID, "Folk Hero", "Legendary Enchantment — Background", effectstest.Msc18FolkHeroOracle, 0, 0)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.Msc18Library(me, 10)
	effectstest.B12Push(g, me.ID, "Folk Hero", "Legendary Enchantment — Background", effectstest.Msc18FolkHeroOracle, 0, 0)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.Msc18Library(me, 10)
	effectstest.Msc18Library(opp, 10)
	effectstest.B12Push(g, me.ID, "Folk Hero", "Legendary Enchantment — Background", effectstest.Msc18FolkHeroOracle, 0, 0)
	effectstest.B12Creature(g, me.ID, "Plain Human", "Creature — Human Warrior", 2, 2)
	hand := me.Hand.Size()

	msc18CastCreature(t, g, me, "Creature — Human")
	if got := me.Hand.Size() - hand; got != 0 {
		t.Errorf("a non-commander creature drew %d", got)
	}
	msc18Commander(g, me, "Legendary Creature — Human")
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.PowerUpMain(t, g)
	oppHand := opp.Hand.Size()
	hand = me.Hand.Size()
	effectstest.CastFor(t, g, opp, "Their Human", "Creature — Human", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size()-hand != 0 || opp.Hand.Size()-oppHand != 0 {
		t.Errorf("an opponent's spell drew: me %d, opp %d", me.Hand.Size()-hand, opp.Hand.Size()-oppHand)
	}
}

// --- Archnemesis ---

func msc18CastArchnemesis(t *testing.T, g *game.Game, me *game.Player, host uuid.UUID) uuid.UUID {
	t.Helper()
	id := effectstest.CastCatalogSpell(t, g, "Archnemesis", "Enchantment — Aura", effectstest.Msc18ArchnemesisOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: host}})
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, id)
	if !ok {
		t.Fatal("Archnemesis did not resolve")
	}
	if c.AttachedTo.Kind != game.TargetPlayer || c.AttachedTo.ID != host {
		t.Fatalf("attached to %+v, want player %s", c.AttachedTo, host)
	}
	return id
}

func TestMsc18ArchnemesisEnchantsOnlyAnOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	err := effectstest.CastCatalogSpellErr(t, g, "Archnemesis", "Enchantment — Aura", effectstest.Msc18ArchnemesisOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	if err == nil {
		t.Fatal("Archnemesis was cast on its own controller")
	}
}

func TestMsc18ArchnemesisDrainsDrawsAndGainsWhenYouAttackTheEnchantedPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.Msc18Library(me, 6)
	msc18CastArchnemesis(t, g, me, opp.ID)
	a := effectstest.B12Creature(g, me.ID, "Attacker A", "Creature — Soldier", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Attacker B", "Creature — Soldier", 2, 2)
	life, oppLife, hand := me.Life, opp.Life, me.Hand.Size()

	effectstest.DeclareAttack(t, g, opp.ID, a, b)
	effectstest.PassPriorityAroundTable(t, g)

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
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.Msc18Library(me, 6)
	msc18CastArchnemesis(t, g, me, opp.ID)
	a := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Soldier", 2, 2)
	life, otherLife, hand := me.Life, other.Life, me.Hand.Size()

	effectstest.DeclareAttack(t, g, other.ID, a)
	effectstest.PassPriorityAroundTable(t, g)

	if other.Life != otherLife || me.Life != life || me.Hand.Size() != hand {
		t.Errorf("an attack on a player who is not enchanted triggered: other %d→%d, me %d→%d, hand %d→%d",
			otherLife, other.Life, life, me.Life, hand, me.Hand.Size())
	}
}

// #4839, CR 508.3e: attacking the enchanted player's planeswalker is not
// attacking the enchanted player.
func TestMsc18ArchnemesisIgnoresAnAttackOnTheEnchantedPlayersPlaneswalker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.Msc18Library(me, 6)
	aura := msc18CastArchnemesis(t, g, me, opp.ID)
	walker := effectstest.P1g2Planeswalker(g, opp.ID, 5)
	a := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Soldier", 2, 2)
	life, oppLife, hand := me.Life, opp.Life, me.Hand.Size()

	effectstest.DeclareAttack(t, g, walker, a)
	if n := effectstest.B784Triggers(g, aura); n != 0 {
		t.Fatalf("an attack on the enchanted player's planeswalker: %d Archnemesis triggers, want 0", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != oppLife || me.Life != life || me.Hand.Size() != hand {
		t.Errorf("the drain ran: opp %d→%d, me %d→%d, hand %d→%d",
			oppLife, opp.Life, life, me.Life, hand, me.Hand.Size())
	}
}

// #4839, CR 508.3e: a player who attacks only your planeswalker has not
// attacked you, so the Aura is not offered to them.
func TestMsc18ArchnemesisDoesNotMoveWhenOnlyYourPlaneswalkerIsAttacked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	aura := msc18CastArchnemesis(t, g, me, other.ID)
	walker := effectstest.P1g2Planeswalker(g, me.ID, 5)
	for g.Turn.ActiveSeat != 1 {
		effectstest.AdvanceToNextSeatsTurn(t, g)
	}
	their := effectstest.B12Creature(g, opp.ID, "Their Attacker", "Creature — Soldier", 2, 2)
	effectstest.DeclareAttack(t, g, walker, their)
	if n := effectstest.B784Triggers(g, aura); n != 0 {
		t.Fatalf("an attack on your planeswalker: %d Archnemesis triggers, want 0", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, aura); c.AttachedTo.ID != other.ID {
		t.Errorf("attached to %+v, want it to stay on %s", c.AttachedTo, other.ID)
	}
}

func TestMsc18ArchnemesisMovesToAPlayerWhoAttacksYou(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	aura := msc18CastArchnemesis(t, g, me, other.ID)
	for g.Turn.ActiveSeat != 1 {
		effectstest.AdvanceToNextSeatsTurn(t, g)
	}
	their := effectstest.B12Creature(g, opp.ID, "Their Attacker", "Creature — Soldier", 2, 2)
	effectstest.DeclareAttack(t, g, me.ID, their)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)

	c, ok := effectstest.BattlefieldCard(g, aura)
	if !ok {
		t.Fatal("Archnemesis left the battlefield")
	}
	if c.AttachedTo.Kind != game.TargetPlayer || c.AttachedTo.ID != opp.ID {
		t.Errorf("attached to %+v, want the attacking player %s", c.AttachedTo, opp.ID)
	}
}

func TestMsc18ArchnemesisStaysWhenYouDeclineTheMove(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	aura := msc18CastArchnemesis(t, g, me, other.ID)
	for g.Turn.ActiveSeat != 1 {
		effectstest.AdvanceToNextSeatsTurn(t, g)
	}
	their := effectstest.B12Creature(g, opp.ID, "Their Attacker", "Creature — Soldier", 2, 2)
	effectstest.DeclareAttack(t, g, me.ID, their)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)

	c, _ := effectstest.BattlefieldCard(g, aura)
	if c.AttachedTo.ID != other.ID {
		t.Errorf("attached to %+v, want it to stay on %s", c.AttachedTo, other.ID)
	}
}

// --- Extract Power ---

func TestMsc18ExtractPowerExilesEachTopCardFaceDownAndLetsYouPlayThemFree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	mine := effectstest.SeedSearchLibrary(me, game.Card{Name: "My Bear", TypeLine: "Creature — Bear", ManaCost: "{2}{G}", Power: 2, Toughness: 2})
	theirs := effectstest.SeedSearchLibrary(opp, game.Card{Name: "Their Forest", TypeLine: "Basic Land — Forest"})
	third := g.Seats[2]
	thirdCard := effectstest.SeedSearchLibrary(third, game.Card{Name: "Third Bolt", TypeLine: "Instant", ManaCost: "{R}"})
	effectstest.CastCatalogSpell(t, g, "Extract Power", "Sorcery", effectstest.Msc18ExtractPowerOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

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
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	mine := effectstest.SeedSearchLibrary(me, game.Card{Name: "My Bear", TypeLine: "Creature — Bear", ManaCost: "{2}{G}", Power: 2, Toughness: 2})
	effectstest.CastCatalogSpell(t, g, "Extract Power", "Sorcery", effectstest.Msc18ExtractPowerOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.CastSpell(opp.ID, mine[0], game.CastSpellParams{FromZone: "exile", Strict: true}); err == nil {
		t.Error("an opponent cast a card exiled by my Extract Power")
	}
}

// --- Heroes' Hangout ---

func TestMsc18HeroesHangoutDateNightPlaysOnlyTheChosenCardUntilYourNextTurnEnds(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	// seedSearchLibrary puts the FIRST card on top.
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Top Card", TypeLine: "Basic Land — Island"},
		game.Card{Name: "Second Card", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Deep Card", TypeLine: "Creature — Bear", Power: 1, Toughness: 1},
	)
	top, second, deep := ids[0], ids[1], ids[2]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Heroes' Hangout", TypeLine: "Sorcery", OracleID: effectstest.Msc18HangoutOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{0}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range []uuid.UUID{top, second} {
		if !g.Exile.Contains(c) {
			t.Fatalf("%s was not exiled", c)
		}
	}
	if !me.Library.Contains(deep) {
		t.Error("the third card left the library")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, second)
	effectstest.PassPriorityAroundTable(t, g)

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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Other Land", TypeLine: "Basic Land — Island"},
		game.Card{Name: "Chosen Land", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Filler", TypeLine: "Creature — Bear", Power: 1, Toughness: 1},
	)
	chosen := ids[1]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Heroes' Hangout", TypeLine: "Sorcery", OracleID: effectstest.Msc18HangoutOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{0}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, chosen)
	effectstest.PassPriorityAroundTable(t, g)

	// Through another seat's turn, to my next turn: still playable there.
	effectstest.PowerUpPassTo(t, g, 0)
	if err := g.CastSpell(me.ID, chosen, game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("play the chosen card on my next turn: %v", err)
	}
	if !g.Battlefield.Contains(chosen) {
		t.Error("the chosen land did not enter on the next turn")
	}
}

func TestMsc18HeroesHangoutDateNightLapsesAfterYourNextTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Other Land", TypeLine: "Basic Land — Island"},
		game.Card{Name: "Chosen Land", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Filler", TypeLine: "Creature — Bear", Power: 1, Toughness: 1},
	)
	chosen := ids[1]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Heroes' Hangout", TypeLine: "Sorcery", OracleID: effectstest.Msc18HangoutOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{0}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, chosen)
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.PowerUpPassTo(t, g, 0)
	effectstest.PowerUpPassTo(t, g, 0)
	if err := g.CastSpell(me.ID, chosen, game.CastSpellParams{FromZone: "exile", Strict: true}); err == nil {
		t.Error("the chosen card was still playable the turn after my next turn")
	}
}

func TestMsc18HeroesHangoutPatrolNightPumpsUpToTwoCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	a := effectstest.B12Creature(g, me.ID, "Hero A", "Creature — Hero", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "Their Guy", "Creature — Soldier", 3, 3)
	c := effectstest.B12Creature(g, me.ID, "Bystander", "Creature — Hero", 2, 2)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Heroes' Hangout", TypeLine: "Sorcery", OracleID: effectstest.Msc18HangoutOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Modes:   []int{1},
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}},
	}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, a); got != 3 {
		t.Errorf("Hero A power = %d, want 3", got)
	}
	if got := effectstest.EffectivePower(t, g, b); got != 4 {
		t.Errorf("their creature power = %d, want 4", got)
	}
	for _, x := range []uuid.UUID{a, b} {
		if !slices.Contains(effectstest.EffectiveAbilities(t, g, x), "first strike") {
			t.Errorf("%s has no first strike: %v", x, effectstest.EffectiveAbilities(t, g, x))
		}
	}
	if got := effectstest.EffectivePower(t, g, c); got != 2 || slices.Contains(effectstest.EffectiveAbilities(t, g, c), "first strike") {
		t.Errorf("the untargeted creature changed: power %d", got)
	}
}

// --- Doomsday ---

const msc18DoomsdayOracle = "721eb5a2-d7cf-4db0-8013-ef3f596c52a5"

func msc18DoomsdayCast(t *testing.T, g *game.Game) {
	t.Helper()
	effectstest.CastCatalogSpell(t, g, "Doomsday", "Sorcery", msc18DoomsdayOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	lib := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "L1", TypeLine: "Instant"}, game.Card{Name: "L2", TypeLine: "Instant"},
		game.Card{Name: "L3", TypeLine: "Instant"}, game.Card{Name: "L4", TypeLine: "Instant"},
		game.Card{Name: "L5", TypeLine: "Instant"}, game.Card{Name: "L6", TypeLine: "Instant"},
	)
	gy := msc18GraveyardCards(me, "G1", "G2", "G3")
	me.Life = 41
	searches := effectstest.CountEvents(g, game.EventSearchLibrary)

	msc18DoomsdayCast(t, g)
	if effectstest.CountEvents(g, game.EventSearchLibrary) != searches+1 {
		t.Error("Doomsday did not search the library")
	}
	keep := []uuid.UUID{lib[0], lib[3], lib[5], gy[0], gy[2]}
	effectstest.AnswerChooseCards(t, g, me.ID, keep...)
	order := effectstest.PutInLibraryChoiceFor(g, me.ID)
	if order == nil {
		t.Fatal("no ordering prompt for the five cards")
	}
	want := []uuid.UUID{gy[2], lib[5], gy[0], lib[0], lib[3]}
	if err := g.ResolvePutInLibrary(order.ID, me.ID, nil, want); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	lib := effectstest.SeedSearchLibrary(me, game.Card{Name: "L1", TypeLine: "Instant"}, game.Card{Name: "L2", TypeLine: "Instant"})
	gy := msc18GraveyardCards(me, "G1")
	me.Life = 7

	msc18DoomsdayCast(t, g)
	prompt := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if prompt == nil || prompt.ChooseMin != 3 || prompt.ChooseMax != 3 {
		t.Fatalf("prompt = %+v, want exactly the 3 cards that exist", prompt)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, lib[0], lib[1], gy[0])
	order := effectstest.PutInLibraryChoiceFor(g, me.ID)
	if order == nil {
		t.Fatal("no ordering prompt")
	}
	if err := g.ResolvePutInLibrary(order.ID, me.ID, nil, []uuid.UUID{lib[0], lib[1], gy[0]}); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Library.Size() != 3 || g.Exile.Size() != 0 {
		t.Errorf("library %d, exile %d, want 3 and 0", me.Library.Size(), g.Exile.Size())
	}
	if me.Life != 3 {
		t.Errorf("life = %d, want 3 (7 less 4)", me.Life)
	}
}

func TestMsc18DoomsdayLeavesOtherPlayersZonesAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	lib := effectstest.SeedSearchLibrary(me, game.Card{Name: "L1", TypeLine: "Instant"})
	theirLib := effectstest.SeedSearchLibrary(opp, game.Card{Name: "T1", TypeLine: "Instant"})
	theirGy := msc18GraveyardCards(opp, "TG")

	msc18DoomsdayCast(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, lib[0])
	effectstest.PassPriorityAroundTable(t, g)
	if g.Exile.Contains(theirLib[0]) || g.Exile.Contains(theirGy[0]) || !opp.Library.Contains(theirLib[0]) {
		t.Error("an opponent's library or graveyard was touched")
	}
}
