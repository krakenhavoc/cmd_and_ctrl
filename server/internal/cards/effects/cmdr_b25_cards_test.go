package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	cmdrB25TimelessDragonOracle = "a5ae93e9-0385-47fd-9d8f-306e9a6b7fe6"
	cmdrB25ChampionOfWitsOracle = "3fd8ba0c-75ec-43ea-8706-f71e6cf1d236"
)

// cmdrB25Copy finds a battlefield card named `name` that `owner` controls
// and that is not `not` (the card the copy was made from).
func cmdrB25Copy(g *game.Game, owner uuid.UUID, name string, not uuid.UUID) *game.Card {
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Name == name && c.Controller == owner && c.InstanceID != not {
			return c
		}
	}
	return nil
}

// cmdrB25Eternalize eternalizes the card from the graveyard (ability index
// `idx`) paying `mana`, and returns the token.
func cmdrB25Eternalize(t *testing.T, g *game.Game, name, typeLine, oracle string, p, tough int, idx int, mana string) (*game.Card, uuid.UUID) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	id := pushCatalogGraveyardCard(me, name, typeLine, oracle, p, tough)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, mana); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, id, idx, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("eternalize %s: %v", name, err)
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(id) {
		t.Errorf("%s: the eternalized card is not in exile", name)
	}
	tok := cmdrB25Copy(g, me.ID, name, id)
	if tok == nil {
		t.Fatalf("%s: no eternalized token", name)
	}
	return tok, id
}

func TestCmdrB25TimelessDragonPlainscyclingNeedsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	id := pushCatalogHandCard(me, "Timeless Dragon", "Creature — Dragon", cmdrB25TimelessDragonOracle)
	if err := hobB4Activate(t, g, me.ID, id, 0, "{C}", nil); err == nil {
		t.Fatal("plainscycling {2} was paid with one mana")
	}
}

func TestCmdrB25TimelessDragonEternalizesIntoABlackFourFourZombieDragon(t *testing.T) {
	g := newCatalogGame(t)
	tok, _ := cmdrB25Eternalize(t, g, "Timeless Dragon", "Creature — Dragon", cmdrB25TimelessDragonOracle, 5, 5, 1, "{C}{C}{W}{W}")
	if tok.Power != 4 || tok.Toughness != 4 {
		t.Errorf("token %d/%d, want 4/4", tok.Power, tok.Toughness)
	}
	if len(tok.Colors) != 1 || tok.Colors[0] != "B" || tok.ManaCost != "" {
		t.Errorf("token colours %v cost %q, want black with no cost", tok.Colors, tok.ManaCost)
	}
	if tok.TypeLine != "Token Creature — Zombie Dragon" {
		t.Errorf("token type line %q", tok.TypeLine)
	}
}

// castChampion casts Champion of Wits free and resolves the spell, leaving
// the enters trigger's prompt open.
func cmdrB25CastChampion(t *testing.T, g *game.Game, me *game.Player) {
	t.Helper()
	if _, err := msc17Cast(t, g, me, game.Card{
		Name: "Champion of Wits", TypeLine: "Creature — Snake Wizard", OracleID: cmdrB25ChampionOfWitsOracle,
		Power: 2, Toughness: 1,
	}, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast Champion of Wits: %v", err)
	}
	passPriorityAroundTable(t, g)
}

func TestCmdrB25ChampionOfWitsDrawsItsPowerThenDiscardsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	msc18Library(me, 8)
	cmdrB25CastChampion(t, g, me)
	ask := latestChoiceOfKind(g, game.PendingChoiceConfirm)
	if ask == nil || ask.Chooser != me.ID {
		t.Fatal("no \"draw?\" prompt on entering")
	}
	hand := me.Hand.Size()
	if err := g.ResolveConfirm(ask.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	if me.Hand.Size() != hand+2 {
		t.Fatalf("hand %d after drawing its power (2), want %d", me.Hand.Size(), hand+2)
	}
	discardFromHand(t, g, me.ID)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d after discarding two, want %d", me.Hand.Size(), hand)
	}
}

func TestCmdrB25ChampionOfWitsDeclineDrawsAndDiscardsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	msc18Library(me, 8)
	cmdrB25CastChampion(t, g, me)
	ask := latestChoiceOfKind(g, game.PendingChoiceConfirm)
	if ask == nil {
		t.Fatal("no prompt")
	}
	hand := me.Hand.Size()
	if err := g.ResolveConfirm(ask.ID, me.ID, false); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	if me.Hand.Size() != hand || discardChoiceFor(g, me.ID) != nil {
		t.Errorf("declining changed the hand (%d, want %d) or asked for a discard", me.Hand.Size(), hand)
	}
}

func TestCmdrB25ChampionOfWitsEternalizedTokenDrawsFour(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	msc18Library(me, 8)
	hand := me.Hand.Size()
	tok, _ := cmdrB25Eternalize(t, g, "Champion of Wits", "Creature — Snake Wizard", cmdrB25ChampionOfWitsOracle, 2, 1, 0, "{C}{C}{C}{C}{C}{U}{U}")
	if tok.Power != 4 || tok.Toughness != 4 {
		t.Errorf("token %d/%d, want 4/4", tok.Power, tok.Toughness)
	}
	ask := latestChoiceOfKind(g, game.PendingChoiceConfirm)
	if ask == nil {
		t.Fatal("the token's enters trigger asked nothing")
	}
	if err := g.ResolveConfirm(ask.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	if me.Hand.Size() != hand+4 {
		t.Errorf("hand %d, want %d (drew four)", me.Hand.Size(), hand+4)
	}
}

const (
	cmdrB25BindingMummyOracle    = "5adc9d0d-3393-4d91-be85-1f9dd344698b"
	cmdrB25CastOutOracle         = "f90b00f6-36e0-4988-9409-57297483a952"
	cmdrB25LordAccursedOracle    = "3f372107-6c52-4f1f-8bf0-7b3fa21b2cae"
	cmdrB25GleamingOverseerOracl = "f7716dff-5de7-4b52-966e-72ee3adc8d9a"
	cmdrB25LazotepOracle         = "c3e67b8f-1c91-4a2f-b924-35bbaff4bab9"
)

func TestCmdrB25BindingMummyTapsOnAnotherZombieEntering(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Binding Mummy", "Creature — Zombie", cmdrB25BindingMummyOracle, 2, 2)
	victim := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	msc17Play(t, g, me, game.Card{Name: "Walker", TypeLine: "Creature — Zombie", Power: 2, Toughness: 2})
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, victim)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if c := findTestCard(g, victim); c == nil || !c.Tapped {
		t.Error("the targeted creature is not tapped")
	}
}

func TestCmdrB25BindingMummyIgnoresNonZombiesAndItself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	msc17Play(t, g, me, game.Card{
		Name: "Binding Mummy", TypeLine: "Creature — Zombie", OracleID: cmdrB25BindingMummyOracle, Power: 2, Toughness: 2,
	})
	if latestPickTarget(g, me.ID) != nil {
		t.Error("the Mummy's own entry triggered it (\"another\")")
	}
	msc17Play(t, g, me, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	if latestPickTarget(g, me.ID) != nil {
		t.Error("a non-Zombie entering triggered it")
	}
}

func TestCmdrB25BindingMummyMayDecline(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Binding Mummy", "Creature — Zombie", cmdrB25BindingMummyOracle, 2, 2)
	victim := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	msc17Play(t, g, me, game.Card{Name: "Walker", TypeLine: "Creature — Zombie", Power: 2, Toughness: 2})
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, victim)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if c := findTestCard(g, victim); c == nil || c.Tapped {
		t.Error("declining still tapped the target")
	}
}

func TestCmdrB25CastOutFlashExilesUntilItLeavesAndCycles(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := b12Creature(g, opp.ID, "Their Threat", "Creature — Bear", 4, 4)
	out := castCatalogSpell(t, g, "Cast Out", "Enchantment", cmdrB25CastOutOracle, nil)
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Fatal("the creature is not exiled")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(out) })
	g.RunStateChecksForTest()
	if n := namedOnBattlefield(g, "Their Threat", opp.ID); n != 1 {
		t.Errorf("after Cast Out left, %d copies under its owner's control, want 1", n)
	}

	g2 := newCatalogGame(t)
	me2 := g2.Seats[g2.Turn.ActiveSeat]
	advanceToMain(t, g2)
	id := pushCatalogHandCard(me2, "Cast Out", "Enchantment", cmdrB25CastOutOracle)
	hand := me2.Hand.Size()
	if err := hobB4Activate(t, g2, me2.ID, id, 0, "{W}", nil); err != nil {
		t.Fatalf("cycling: %v", err)
	}
	passPriorityAroundTable(t, g2)
	if me2.Hand.Size() != hand || !me2.Graveyard.Contains(id) {
		t.Errorf("cycling: hand %d (want %d), in graveyard %v", me2.Hand.Size(), hand, me2.Graveyard.Contains(id))
	}
}

func TestCmdrB25CastOutOnlyTargetsOpponentsNonlandPermanents(t *testing.T) {
	spec, _ := Lookup(cmdrB25CastOutOracle)
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirLand := b12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	rfPermAExpectLegal(t, g, me.ID, spec.Triggered[0].Targets, map[uuid.UUID]bool{mine: false, theirLand: false, theirs: true})
}

func TestCmdrB25LordOfTheAccursedAnthemAndMenace(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lord := b12Push(g, me.ID, "Lord of the Accursed", "Creature — Zombie", cmdrB25LordAccursedOracle, 2, 3)
	mine := b12Creature(g, me.ID, "Walker", "Creature — Zombie", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Walker", "Creature — Zombie", 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if got := effectiveOf(t, g, mine).Power; got != 3 {
		t.Errorf("my other Zombie power %d, want 3", got)
	}
	for name, id := range map[string]uuid.UUID{"the lord": lord, "their Zombie": theirs, "a Bear": bear} {
		if got := effectiveOf(t, g, id).Power; got != 2 {
			t.Errorf("%s power %d, want 2 (the anthem is other Zombies you control)", name, got)
		}
	}
	if err := hobB4Activate(t, g, me.ID, lord, 0, "{C}{B}", nil); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	for name, id := range map[string]uuid.UUID{"the lord": lord, "my Zombie": mine, "their Zombie": theirs} {
		if !containsString(effectiveAbilities(t, g, id), "menace") {
			t.Errorf("%s lacks menace (all Zombies gain it)", name)
		}
	}
	if containsString(effectiveAbilities(t, g, bear), "menace") {
		t.Error("a Bear gained menace")
	}
}

func TestCmdrB25GleamingOverseerAmassesAndGivesZombieTokensHexproofAndMenace(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	msc17Play(t, g, me, game.Card{
		Name: "Gleaming Overseer", TypeLine: "Creature — Zombie Wizard", OracleID: cmdrB25GleamingOverseerOracl, Power: 1, Toughness: 4,
	})
	army := armyOf(t, g, me.ID)
	if got := countersOn(g, army.InstanceID, game.CounterPlusOne); got != 1 {
		t.Errorf("+1/+1 counters %d, want 1", got)
	}
	abs := effectiveAbilities(t, g, army.InstanceID)
	if !containsString(abs, "hexproof") || !containsString(abs, "menace") {
		t.Errorf("the Zombie Army has %v, want hexproof and menace", abs)
	}
}

func TestCmdrB25LazotepChancellorPaysOneToAmassOnDiscard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Lazotep Chancellor", "Creature — Zombie Wizard", cmdrB25LazotepOracle, 1, 3)
	advanceToMain(t, g)
	id := pushCatalogHandCard(me, "Spare", "Sorcery", "")
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}"); err != nil {
		t.Fatal(err)
	}
	_ = id
	g.WithWriteLock(func() { g.DiscardChoiceForEffect(me.ID, 1) })
	discardFromHand(t, g, me.ID)
	passPriorityAroundTable(t, g)
	answerPayUnless(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	army := armyOf(t, g, me.ID)
	if got := countersOn(g, army.InstanceID, game.CounterPlusOne); got != 2 {
		t.Errorf("+1/+1 counters %d, want 2", got)
	}
}

const (
	cmdrB25WhirlerRogueOracle = "7060f2c8-fca0-4b7a-bddf-37682f434596"
	cmdrB25PiaKiranOracle     = "86084ed9-b8bb-4289-9579-6056194787bc"
)

// cmdrB25Thopters counts the Thopter tokens `owner` controls.
func cmdrB25Thopters(g *game.Game, owner uuid.UUID) int {
	return namedOnBattlefield(g, "Thopter", owner)
}

func TestCmdrB25WhirlerRogueMakesThoptersAndTapsTwoArtifactsForUnblockable(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	msc17Play(t, g, me, game.Card{
		Name: "Whirler Rogue", TypeLine: "Creature — Human Rogue Artificer", OracleID: cmdrB25WhirlerRogueOracle, Power: 2, Toughness: 2,
	})
	if got := cmdrB25Thopters(g, me.ID); got != 2 {
		t.Fatalf("%d Thopters, want 2", got)
	}
	var rogue uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Whirler Rogue" {
			rogue = c.InstanceID
		}
	}
	target := b12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	var thopters []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Thopter" {
			thopters = append(thopters, c.InstanceID)
		}
	}
	if err := g.ActivateCatalogAbility(me.ID, rogue, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}}, TapIDs: thopters, Strict: true,
	}); err != nil {
		t.Fatalf("tap two Thopters: %v", err)
	}
	passPriorityAroundTable(t, g)
	tapped := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Thopter" && c.Tapped {
			tapped++
		}
	}
	if tapped != 2 {
		t.Errorf("%d Thopters tapped, want 2", tapped)
	}
}

func TestCmdrB25WhirlerRogueNeedsTwoUntappedArtifacts(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	rogue := b12Push(g, me.ID, "Whirler Rogue", "Creature — Human Rogue Artificer", cmdrB25WhirlerRogueOracle, 2, 2)
	trinket := b12Permanent(g, me.ID, "Lone Trinket", "Artifact")
	target := b12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, rogue, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}}, TapIDs: []uuid.UUID{trinket}, Strict: true,
	}); err == nil {
		t.Fatal("activated with only one artifact to tap")
	}
}

func TestCmdrB25PiaAndKiranNalaarMakesThoptersAndSacrificesForDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	msc17Play(t, g, me, game.Card{
		Name: "Pia and Kiran Nalaar", TypeLine: "Legendary Creature — Human Artificer", OracleID: cmdrB25PiaKiranOracle, Power: 2, Toughness: 2,
	})
	if got := cmdrB25Thopters(g, me.ID); got != 2 {
		t.Fatalf("%d Thopters, want 2", got)
	}
	var pia uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Pia and Kiran Nalaar" {
			pia = c.InstanceID
		}
	}
	life := opp.Life
	var thopter uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Thopter" {
			thopter = c.InstanceID
		}
	}
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{R}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, pia, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}, SacrificeIDs: []uuid.UUID{thopter}, Strict: true,
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("life %d, want %d", opp.Life, life-2)
	}
	if got := cmdrB25Thopters(g, me.ID); got != 1 {
		t.Errorf("%d Thopters left, want 1 (one was sacrificed)", got)
	}
}

func TestCmdrB25PiaAndKiranNalaarNeedsAnArtifactToSacrifice(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pia := b12Push(g, me.ID, "Pia and Kiran Nalaar", "Legendary Creature — Human Artificer", cmdrB25PiaKiranOracle, 2, 2)
	if err := hobB4Activate(t, g, me.ID, pia, 0, "{C}{C}{R}", []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}); err == nil {
		t.Fatal("activated with no artifact to sacrifice")
	}
}

const cmdrB25VizierOracle = "55bceb83-bbba-477f-98ce-6e0a80b91080"

func TestCmdrB25VizierOfManyFacesCastCopiesWithoutAnException(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bears := seedCopyableCreature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	id := castCatalogSpell(t, g, "Vizier of Many Faces", "Creature — Shapeshifter Cleric", cmdrB25VizierOracle, nil)
	resolveWithCopyChoice(t, g, bears)
	got := copyBattlefieldCard(t, g, id)
	if got.Name != "Grizzly Bears" || got.TypeLine != "Creature — Bear" {
		t.Errorf("copy is %q %q, want a plain Grizzly Bears", got.Name, got.TypeLine)
	}
}

func TestCmdrB25VizierOfManyFacesEmbalmedTokenIsAWhiteZombieCopyWithNoManaCost(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	bears := seedCopyableCreature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	// The Bears are green so the white exception is visible.
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == bears {
				g.Battlefield.Cards[i].Colors = []string{"G"}
				g.Battlefield.Cards[i].ManaCost = "{1}{G}"
			}
		}
	})
	gy := pushCatalogGraveyardCard(me, "Vizier of Many Faces", "Creature — Shapeshifter Cleric", cmdrB25VizierOracle, 0, 0)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{C}{U}{U}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, gy, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("embalm: %v", err)
	}
	resolveWithCopyChoice(t, g, bears)
	tok := cmdrB25Copy(g, me.ID, "Grizzly Bears", bears)
	if tok == nil {
		t.Fatal("the embalm token did not enter as a Grizzly Bears")
	}
	if !tok.IsToken() {
		t.Error("not a token")
	}
	if tok.ManaCost != "" {
		t.Errorf("mana cost %q, want none", tok.ManaCost)
	}
	if len(tok.Colors) != 1 || tok.Colors[0] != "W" {
		t.Errorf("colours %v, want white", tok.Colors)
	}
	if !tok.HasSubtype("Zombie") || !tok.HasSubtype("Bear") {
		t.Errorf("type line %q, want a Zombie in addition to Bear", tok.TypeLine)
	}
}

const cmdrB25NeverReturnOracle = "bf276236-18a6-4a48-b47d-5227832d26e6"

func cmdrB25NeverReturnCard(owner uuid.UUID) game.Card {
	c := game.Card{
		InstanceID: uuid.New(),
		OracleID:   cmdrB25NeverReturnOracle,
		Layout:     game.LayoutSplit,
		Owner:      owner,
		Controller: owner,
		Faces: []game.Face{
			{Name: "Never", TypeLine: "Sorcery", ManaCost: "{1}{B}{B}", OracleText: "Destroy target creature or planeswalker."},
			{Name: "Return", TypeLine: "Sorcery", ManaCost: "{3}{B}",
				OracleText: "Aftermath (Cast this spell only from your graveyard. Then exile it.)\nExile target card from a graveyard. Create a 2/2 black Zombie creature token."},
		},
	}
	c.SettleImported()
	return c
}

func TestCmdrB25NeverDestroysACreatureOrPlaneswalkerOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	land := b12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	advanceToMain(t, g)
	c := cmdrB25NeverReturnCard(me.ID)
	me.Hand.PushTop(c)
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: land}}}); err == nil {
		t.Fatal("Never targeted a land")
	}
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{B}{B}"); err != nil {
		t.Fatal(err)
	}
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}); err != nil {
		t.Fatalf("cast Never: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the creature survived Never")
	}
}

func TestCmdrB25ReturnExilesACardAndMakesAZombieThenIsExiled(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	advanceToMain(t, g)
	victim := pushCatalogGraveyardCard(opp, "Their Card", "Creature — Bear", "", 2, 2)
	c := cmdrB25NeverReturnCard(me.ID)
	me.Graveyard.PushTop(c)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{C}{B}"); err != nil {
		t.Fatal(err)
	}
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Face: 1, FromZone: "graveyard", Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}}}); err != nil {
		t.Fatalf("cast Return: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(victim) {
		t.Error("the targeted card is not exiled")
	}
	if namedOnBattlefield(g, "Zombie", me.ID) != 1 {
		t.Error("no 2/2 Zombie token")
	}
	if !g.Exile.Contains(c.InstanceID) {
		t.Error("Never // Return is not exiled after aftermath")
	}
}
