package effects

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// aura_host_cards_test.go — CR 303.4f / 303.4g (#3158): an Aura put onto
// the battlefield other than by resolving as an Aura spell chooses what
// it enchants as it enters, and stays where it was when nothing is
// legal. The engine half is game/aura_host.go; these are the cards.

// ahGraveyardAura puts a catalogued Aura card into p's graveyard.
func ahGraveyardAura(p *game.Player, name, oracle, cost, color string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: auraTypeLine, OracleID: oracle,
		ManaCost: cost, Colors: []string{color}, Owner: p.ID, Controller: p.ID,
	})
	return id
}

func ahPacifism(p *game.Player) uuid.UUID {
	return ahGraveyardAura(p, "Pacifism", pacifismOracle, "{1}{W}", "W")
}

func ahAnimateDead(p *game.Player) uuid.UUID {
	return ahGraveyardAura(p, "Animate Dead", animateDeadOracle, "{1}{B}", "B")
}

// auraHostChoiceFor is the open aura_host prompt addressed to chooser.
func auraHostChoiceFor(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceAuraHost && c.Chooser == chooser {
			return c
		}
	}
	return nil
}

func answerAuraHost(t *testing.T, g *game.Game, chooser uuid.UUID, host game.TargetRef) {
	t.Helper()
	c := auraHostChoiceFor(g, chooser)
	if c == nil {
		t.Fatalf("no aura_host prompt for %s: %+v", chooser, g.PendingChoices)
	}
	if err := g.ResolveAuraHost(c.ID, chooser, host); err != nil {
		t.Fatalf("ResolveAuraHost: %v", err)
	}
}

// sunTitanReturns casts Sun Titan and returns `card` from the caster's
// graveyard with its enters trigger, stopping at whatever the entry asks.
func sunTitanReturns(t *testing.T, g *game.Game, card uuid.UUID) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	titan := castAndResolveCreature(t, g, "Sun Titan", "Creature — Giant", sunTitanOracle)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	pickCard(t, g, me.ID, card)
	passPriorityAroundTable(t, g)
	return titan
}

// The headline case: Sun Titan returns Animate Dead, which chooses a
// creature card in any graveyard as it enters (not a battlefield
// creature, not itself), enters attached to it, and its enters trigger
// then returns that card under its controller's control (ADR 0154).
func TestSunTitanReturnsAnimateDeadWhichChoosesAGraveyardCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	aura := ahAnimateDead(me)
	mine := adGraveyardCreature(me, "My Ghoul")
	theirs := adGraveyardCreature(opp, "Their Zombie")
	bear := pushTribalCreature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)

	titan := sunTitanReturns(t, g, aura)
	c := auraHostChoiceFor(g, me.ID)
	if c == nil {
		t.Fatalf("Animate Dead did not ask what it enchants: %+v", g.PendingChoices)
	}
	if !hasID(c.PickTargetCards, mine) || !hasID(c.PickTargetCards, theirs) {
		t.Errorf("both creature cards in graveyards are offered: %v", c.PickTargetCards)
	}
	for _, id := range []uuid.UUID{bear, titan, aura} {
		if hasID(c.PickTargetCards, id) {
			t.Errorf("%s is not a creature card in a graveyard", cardByID(g, id).Name)
		}
	}
	if adZone(g, aura) != game.ZoneGraveyard {
		t.Fatal("nothing moves until the question is answered")
	}

	answerAuraHost(t, g, me.ID, cardRef(theirs))
	a := adCard(t, g, aura)
	if adZone(g, aura) != game.ZoneBattlefield || !a.IsAttachedTo(theirs) {
		t.Fatalf("Animate Dead enters attached to the chosen card: zone=%v host=%+v", adZone(g, aura), a.AttachedTo)
	}
	if triggerOnStack(g, aura) == nil {
		t.Fatal("its enters trigger waits on the stack")
	}
	passPriorityAroundTable(t, g)
	z := adCard(t, g, theirs)
	if adZone(g, theirs) != game.ZoneBattlefield || z.Controller != me.ID {
		t.Fatalf("the opponent's creature returns under my control: zone=%v controller=%v", adZone(g, theirs), z.Controller)
	}
	if a := adCard(t, g, aura); !a.IsAttachedTo(theirs) || adZone(g, aura) != game.ZoneBattlefield {
		t.Fatal("Animate Dead stays attached to the returned creature")
	}
	if got := effectivePower(t, g, theirs); got != 2 {
		t.Errorf("enchanted creature gets -1/-0: power %d, want 2", got)
	}
}

// CR 303.4g: with no creature card in any graveyard there is nothing it
// can enchant, so it stays in the graveyard — no prompt, no trip through
// the battlefield, no enters trigger.
func TestAnimateDeadWithNothingToEnchantStaysInTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	aura := ahAnimateDead(me)
	pushTribalCreature(g, me.ID, "Bear", "Creature — Bear", 2, 2)

	sunTitanReturns(t, g, aura)
	if auraHostChoiceFor(g, me.ID) != nil {
		t.Fatal("nothing to choose, nothing to ask")
	}
	if adZone(g, aura) != game.ZoneGraveyard {
		t.Fatalf("Animate Dead stays in the graveyard, is in %v", adZone(g, aura))
	}
	for _, ev := range g.Events {
		if ev.Kind == game.EventETB && ev.CardID == aura {
			t.Fatal("it never entered")
		}
	}
}

// The predicate: the Aura's own enchant clause (Pacifism's "enchant
// creature": no land, no artifact), protection refuses a host (CR
// 702.16c), and hexproof does not — choosing a host is not targeting.
// One legal host is a forced choice and asks nothing.
func TestAuraHostFollowsTheEnchantClauseAndProtectionButNotHexproof(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	aura := ahPacifism(me)
	hexproof := pushTribalCreature(g, opp.ID, "Hexproof Bear", "Creature — Bear", 2, 2, "hexproof")
	proWhite := pushTribalCreature(g, opp.ID, "White Ward Bear", "Creature — Bear", 2, 2, "protection from white")
	rock := pushCatalogPermanent(g, opp.ID, "Rock", "Artifact", "", false)
	land := pushCatalogPermanent(g, opp.ID, "Forest", "Basic Land — Forest", "", false)

	titan := sunTitanReturns(t, g, aura)
	c := auraHostChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("two legal hosts: the player chooses")
	}
	if !hasID(c.PickTargetCards, hexproof) || !hasID(c.PickTargetCards, titan) {
		t.Errorf("a hexproof creature and my own creature are legal hosts: %v", c.PickTargetCards)
	}
	for _, id := range []uuid.UUID{proWhite, rock, land} {
		if hasID(c.PickTargetCards, id) {
			t.Errorf("%s is not a legal host for a white enchant-creature Aura", cardByID(g, id).Name)
		}
	}
	if len(c.PickTargetPlayers) != 0 {
		t.Errorf("enchant creature offers no player: %v", c.PickTargetPlayers)
	}
	if err := g.ResolveAuraHost(c.ID, me.ID, cardRef(proWhite)); err == nil {
		t.Fatal("an answer outside the legal set is refused")
	}
	answerAuraHost(t, g, me.ID, cardRef(hexproof))
	if host := attachmentHostOf(t, g, aura); host != cardRef(hexproof) {
		t.Fatalf("Pacifism enchants the hexproof creature, got %+v", host)
	}
}

// Replenish's two Auras each ask their own question, among the
// permanents already there — not the enchantment creature Replenish is
// returning at the same time — and each enters attached to its answer.
func TestReplenishReturnsTwoAurasAndEachChoosesItsHost(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	first := ahPacifism(me)
	second := ahPacifism(me)
	god := b17GraveyardCard(me, "God", "Legendary Enchantment Creature — God", "{3}{W}")
	a := pushTribalCreature(g, opp.ID, "Bear A", "Creature — Bear", 2, 2)
	b := pushTribalCreature(g, opp.ID, "Bear B", "Creature — Bear", 3, 3)

	castCatalogSpell(t, g, "Replenish", "Sorcery", b36ReplenishOracle, nil)
	passPriorityAroundTable(t, g)
	c := auraHostChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the first Aura asks")
	}
	if hasID(c.PickTargetCards, god) {
		t.Error("an Aura can't enchant another permanent Replenish returns")
	}
	if g.Battlefield.Contains(god) {
		t.Fatal("nothing lands until every Aura has chosen")
	}
	answerAuraHost(t, g, me.ID, cardRef(a))
	if auraHostChoiceFor(g, me.ID) == nil {
		t.Fatal("the second Aura asks its own question")
	}
	answerAuraHost(t, g, me.ID, cardRef(b))
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(god) {
		t.Error("the enchantment creature returns")
	}
	hosts := map[game.TargetRef]bool{attachmentHostOf(t, g, first): true, attachmentHostOf(t, g, second): true}
	if !hosts[cardRef(a)] || !hosts[cardRef(b)] {
		t.Fatalf("each Pacifism enchants the creature chosen for it: %v", hosts)
	}
}

// A graveyard clause and a simultaneous entry: Replenish returns Animate
// Dead and an enchantment creature card together. That card is leaving
// the graveyard in the same entry, so it is not a host; the creature card
// in the opponent's graveyard is the only one, a forced choice, and it
// comes back under Animate Dead's controller.
func TestReplenishedAnimateDeadCantEnchantACardReplenishReturns(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	aura := ahAnimateDead(me)
	god := b17GraveyardCard(me, "God", "Legendary Enchantment Creature — God", "{3}{W}")
	ghoul := adGraveyardCreature(opp, "Their Ghoul")
	castCatalogSpell(t, g, "Replenish", "Sorcery", b36ReplenishOracle, nil)
	passPriorityAroundTable(t, g)
	if c := auraHostChoiceFor(g, me.ID); c != nil {
		t.Fatalf("one legal host, nothing to ask: %v", c.PickTargetCards)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(god) {
		t.Fatal("the enchantment creature returns")
	}
	if adZone(g, ghoul) != game.ZoneBattlefield || adCard(t, g, ghoul).Controller != me.ID {
		t.Fatalf("Animate Dead returns the opponent's Ghoul under my control: zone=%v", adZone(g, ghoul))
	}
	if !adCard(t, g, aura).IsAttachedTo(ghoul) {
		t.Fatal("Animate Dead stays on the Ghoul")
	}
}

// Undo mid-choice: a clone taken while the question is open restores
// the question, and answering it differently is honoured. The paused
// entry is a continuation the persistence snapshot cannot carry, and its
// census says so (every entry question's frame is).
func TestAuraHostSurvivesAnUndoMidChoice(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	aura := ahAnimateDead(me)
	mine := adGraveyardCreature(me, "My Ghoul")
	theirs := adGraveyardCreature(opp, "Their Zombie")
	sunTitanReturns(t, g, aura)
	if auraHostChoiceFor(g, me.ID) == nil {
		t.Fatal("setup: Animate Dead asks")
	}
	if g.CaptureSnapshot().Restorable() {
		t.Error("a paused entry is not a restore point")
	}

	snap := g.Clone()
	answerAuraHost(t, g, me.ID, cardRef(theirs))
	if !adCard(t, g, aura).IsAttachedTo(theirs) {
		t.Fatal("the first answer attaches")
	}

	g.RestoreFrom(snap)
	me = g.Seats[0]
	if adZone(g, aura) != game.ZoneGraveyard || auraHostChoiceFor(g, me.ID) == nil {
		t.Fatalf("undo puts the question back with the Aura in the graveyard: zone=%v", adZone(g, aura))
	}
	answerAuraHost(t, g, me.ID, cardRef(mine))
	passPriorityAroundTable(t, g)
	if adZone(g, mine) != game.ZoneBattlefield || adZone(g, theirs) != game.ZoneGraveyard {
		t.Fatalf("the restored answer returns my Ghoul: mine=%v theirs=%v", adZone(g, mine), adZone(g, theirs))
	}
}

// The bot's move list: one answer per legal host, each accepted.
func TestAuraHostEnumeratorOffersEveryHost(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	aura := ahAnimateDead(me)
	mine := adGraveyardCreature(me, "My Ghoul")
	theirs := adGraveyardCreature(opp, "Their Zombie")
	sunTitanReturns(t, g, aura)
	c := auraHostChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("setup: Animate Dead asks")
	}
	offered := map[uuid.UUID]bool{}
	var mineMove *legal.Move
	moves := legal.EnumerateFor(g, me.ID)
	for i, m := range moves {
		if m.Type != legal.TypeResolveChoice {
			continue
		}
		var p struct {
			ChoiceID string `json:"choice_id"`
			Target   *struct {
				Kind string `json:"kind"`
				ID   string `json:"id"`
			} `json:"target"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil || p.ChoiceID != c.ID.String() || p.Target == nil {
			continue
		}
		id := uuid.MustParse(p.Target.ID)
		offered[id] = true
		if id == mine {
			mineMove = &moves[i]
		}
	}
	if len(offered) != 2 || !offered[mine] || !offered[theirs] || mineMove == nil {
		t.Fatalf("one answer per creature card in a graveyard, got %v", offered)
	}
	// The plural form the client's banner sends is routed too; a
	// two-ref answer is refused and leaves the prompt open.
	two, _ := json.Marshal(map[string]any{"choice_id": c.ID.String(), "targets": []map[string]string{
		{"kind": "card", "id": mine.String()}, {"kind": "card", "id": theirs.String()},
	}})
	if err := actions.Dispatch(g, actions.Action{Type: actions.Type(legal.TypeResolveChoice), Player: me.ID, Caller: me.ID, Params: two}); err == nil {
		t.Fatal("an Aura enchants one host")
	}
	if err := actions.Dispatch(g, actions.Action{Type: actions.Type(mineMove.Type), Player: me.ID, Caller: me.ID, Params: mineMove.Params}); err != nil {
		t.Fatalf("the enumerated answer is refused: %v", err)
	}
	if !adCard(t, g, aura).IsAttachedTo(mine) {
		t.Fatal("an enumerated answer is accepted")
	}
}

// Brilliant Restoration: the Aura returns attached. One legal host is a
// forced choice, so it asks nothing.
func TestBrilliantRestorationReturnsAnAuraAttached(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	aura := ahPacifism(me)
	bear := pushTribalCreature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Brilliant Restoration", "Sorcery", b21BrilliantRestorationOracle, nil)
	passPriorityAroundTable(t, g)
	if auraHostChoiceFor(g, me.ID) != nil {
		t.Fatal("one host: nothing to ask")
	}
	if host := attachmentHostOf(t, g, aura); host != cardRef(bear) {
		t.Fatalf("Pacifism returns attached to the only creature, got %+v", host)
	}
}

// Open the Vaults: an opponent's Aura comes back under its owner, and
// its owner chooses what it enchants.
func TestOpenTheVaultsLetsEachAurasOwnerChoose(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	aura := ahPacifism(opp)
	mineBear := pushTribalCreature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	pushTribalCreature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Open the Vaults", "Sorcery", b38OpenTheVaultsOracle, nil)
	passPriorityAroundTable(t, g)
	if auraHostChoiceFor(g, me.ID) != nil || auraHostChoiceFor(g, opp.ID) == nil {
		t.Fatal("the Aura's owner, not the caster, chooses")
	}
	answerAuraHost(t, g, opp.ID, cardRef(mineBear))
	passPriorityAroundTable(t, g)
	a := adCard(t, g, aura)
	if a.Controller != opp.ID || !a.IsAttachedTo(mineBear) {
		t.Fatalf("their Pacifism returns under them on my Bear: controller=%v host=%+v", a.Controller, a.AttachedTo)
	}
}

// Carmen's attack returns an Aura card, which chooses its host.
func TestCarmenReturnsAnAuraThatChoosesItsHost(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	carmen := b12Push(g, me.ID, "Carmen, Cruel Skymarcher", "Legendary Creature — Vampire Soldier", b26CarmenTestOracle, 2, 2)
	aura := ahPacifism(me)
	bear := pushTribalCreature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, carmen)
	pickCard(t, g, me.ID, aura)
	passPriorityAroundTable(t, g)
	answerAuraHost(t, g, me.ID, cardRef(bear))
	if host := attachmentHostOf(t, g, aura); host != cardRef(bear) {
		t.Fatalf("Pacifism enters on the chosen Bear, got %+v", host)
	}
}

// It That Betrays takes an opponent's sacrificed Aura, and ITS controller
// — the player it enters under — chooses the host.
func TestItThatBetraysTakesASacrificedAuraAndChoosesItsHost(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	betrays := pushAnnCatalogCreature(g, me.ID, "It That Betrays", itThatBetraysOracle, 11, 11)
	theirBear := pushTribalCreature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	aura := pushCatalogPermanent(g, opp.ID, "Pacifism", auraTypeLine, pacifismOracle, false)
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(aura, cardRef(betrays)); err != nil {
			t.Fatalf("attach: %v", err)
		}
	})
	if err := g.SacrificePermanent(opp.ID, aura); err != nil {
		t.Fatalf("sacrifice: %v", err)
	}
	passPriorityAroundTable(t, g)
	if auraHostChoiceFor(g, opp.ID) != nil {
		t.Fatal("the Aura enters under my control, so I choose")
	}
	answerAuraHost(t, g, me.ID, cardRef(theirBear))
	a := adCard(t, g, aura)
	if a.Controller != me.ID || !a.IsAttachedTo(theirBear) {
		t.Fatalf("their Pacifism is mine, on their Bear: controller=%v host=%+v", a.Controller, a.AttachedTo)
	}
}

// Nazgûl Battle-Mace: the same, through the equipped creature's granted
// trigger and the "unless that player pays 3 life" prompt.
func TestNazgulBattleMaceTakesASacrificedAuraThatChoosesItsHost(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	host := pushAnnCatalogCreature(g, me.ID, "Knight", "", 3, 3)
	mace := pushCatalogPermanent(g, me.ID, "Nazgûl Battle-Mace", "Artifact — Equipment", nazgulBattleMaceOracle, false)
	theirBear := pushTribalCreature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	aura := pushCatalogPermanent(g, opp.ID, "Pacifism", auraTypeLine, pacifismOracle, false)
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(mace, cardRef(host)); err != nil {
			t.Fatalf("attach: %v", err)
		}
		if err := g.AttachForEffect(aura, cardRef(host)); err != nil {
			t.Fatalf("attach: %v", err)
		}
	})
	// The grant is read off the host's characteristics, as the Mace's
	// own test reads them before the sacrifice.
	if c := annCard(t, g, host); !game.HasKeyword(&c, "menace") {
		t.Fatal("setup: the Mace's grant")
	}
	if err := g.SacrificePermanent(opp.ID, aura); err != nil {
		t.Fatalf("sacrifice: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerOptionPick(t, g, opp.ID, 1) // don't pay
	passPriorityAroundTable(t, g)
	answerAuraHost(t, g, me.ID, cardRef(theirBear))
	a := adCard(t, g, aura)
	if a.Controller != me.ID || !a.IsAttachedTo(theirBear) {
		t.Fatalf("the Aura is mine, on the chosen Bear: controller=%v host=%+v", a.Controller, a.AttachedTo)
	}
}

// Ondu Spiritdancer: a token copy of an Aura chooses its host as it is
// created (CR 303.4f), and stays on the battlefield attached.
func TestOnduSpiritdancerCopiesAnAuraOntoAChosenHost(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ondu := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Ondu Spiritdancer", TypeLine: "Creature — Kor Cleric",
		OracleID: onduSpiritdancerOracle, Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID,
	})
	first := pushTribalCreature(g, opp.ID, "Bear A", "Creature — Bear", 2, 2)
	second := pushTribalCreature(g, opp.ID, "Bear B", "Creature — Bear", 2, 2)
	castFromHand(t, g, game.Card{Name: "Pacifism", TypeLine: auraTypeLine, OracleID: pacifismOracle},
		game.CastSpellParams{Targets: []game.TargetRef{cardRef(first)}})
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	c := auraHostChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the token copy asks what it enchants")
	}
	if !hasID(c.PickTargetCards, ondu) || !hasID(c.PickTargetCards, second) {
		t.Errorf("every creature is a legal host: %v", c.PickTargetCards)
	}
	answerAuraHost(t, g, me.ID, cardRef(second))
	passPriorityAroundTable(t, g)
	if n := cardsNamed(g, "Pacifism", me.ID); n != 2 {
		t.Fatalf("%d Pacifisms, want the card and its token copy", n)
	}
	onSecond := false
	for _, card := range g.Battlefield.Cards {
		if card.Name == "Pacifism" && card.IsAttachedTo(second) {
			onSecond = true
		}
	}
	if !onSecond {
		t.Error("the token copy enchants the chosen Bear")
	}
}

// Zur the Enchanter: a fetched Aura chooses its host.
func TestZurFetchesAnAuraThatChoosesItsHost(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	zur := b12Push(g, me.ID, "Zur the Enchanter", "Legendary Creature — Human Wizard", b26ZurTheEnchanterOracle, 1, 4)
	bear := pushTribalCreature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	lib := seedSearchLibrary(me, game.Card{Name: "Pacifism", TypeLine: auraTypeLine, OracleID: pacifismOracle, ManaCost: "{1}{W}", Colors: []string{"W"}})
	declareAttack(t, g, opp.ID, zur)
	passPriorityAroundTable(t, g)
	answerSearchByID(t, g, me.ID, lib[0])
	c := auraHostChoiceFor(g, me.ID)
	if c == nil || !hasID(c.PickTargetCards, zur) || !hasID(c.PickTargetCards, bear) {
		t.Fatalf("the fetched Aura asks between Zur and the Bear: %+v", c)
	}
	answerAuraHost(t, g, me.ID, cardRef(bear))
	if host := attachmentHostOf(t, g, lib[0]); host != cardRef(bear) {
		t.Fatalf("Pacifism enters on the Bear, got %+v", host)
	}
}

// Guardian Sunmare: the saddled attack's fetch, the same question.
func TestGuardianSunmareFetchesAnAuraThatChoosesItsHost(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mount := pushMountForTest(g, me.ID, "Guardian Sunmare", guardianSunmareOracle, 4)
	saddler := pushCrewerForTest(g, me.ID, "Big Saddler", 4)
	bear := pushTribalCreature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	aura := pushLibraryCard(me, game.Card{Name: "Pacifism", TypeLine: auraTypeLine, OracleID: pacifismOracle, ManaCost: "{1}{W}", Colors: []string{"W"}})
	saddleThenAttack(t, g, me.ID, opp.ID, mount, saddler)
	passPriorityAroundTable(t, g)
	answerSearchByID(t, g, me.ID, aura)
	answerAuraHost(t, g, me.ID, cardRef(bear))
	if host := attachmentHostOf(t, g, aura); host != cardRef(bear) {
		t.Fatalf("Pacifism enters on the Bear, got %+v", host)
	}
}

// Armored Skyhunter offers an Aura card among the six now, and the Aura
// chooses its host; it is not offered the Equipment attach.
func TestArmoredSkyhunterPutsAnAuraThatChoosesItsHost(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	sky := b12Push(g, me.ID, "Armored Skyhunter", "Creature — Cat Knight", plSkyhunterOracle, 3, 3)
	bear := pushTribalCreature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	aura := pushLibraryCard(me, game.Card{Name: "Pacifism", TypeLine: auraTypeLine, OracleID: pacifismOracle, ManaCost: "{1}{W}", Colors: []string{"W"}})
	declareAttack(t, g, opp.ID, sky)
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID, aura)
	answerAuraHost(t, g, me.ID, cardRef(bear))
	if host := attachmentHostOf(t, g, aura); host != cardRef(bear) {
		t.Fatalf("Pacifism enters on the Bear, got %+v", host)
	}
	if c := chooseCardsChoiceFor(g, me.ID); c != nil {
		t.Error("an Aura is not offered the Equipment's attach")
	}
}

// Light-Paws names the host, so nothing is asked, and an Aura that can't
// enchant Light-Paws stays in the library (CR 303.4h).
func TestLightPawsLeavesAnAuraThatCantEnchantItInTheLibrary(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	paws := b29Push(g, me.ID, "Light-Paws, Emperor's Voice", "Legendary Creature — Fox Advisor", b29LightPawsOracle, "{1}{W}", 2, 2, "W")
	bear := pushTribalCreature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	// A land it COULD enchant, were it choosing: it is not.
	pushCatalogPermanent(g, me.ID, "Forest", "Basic Land — Forest", "", false)
	ground := pushLibraryCardForTest(me, game.Card{Name: "Fertile Ground", TypeLine: auraTypeLine, OracleID: fertileGroundOracle, ManaCost: "{1}{G}", Colors: []string{"G"}})
	pushLibraryCardForTest(me, game.Card{Name: "Plains", TypeLine: "Basic Land — Plains"})
	advanceToMain(t, g)
	b29CastAs(t, g, me, "Ethereal Armor", auraTypeLine, "", "{W}{W}", game.CastSpellParams{
		Targets: []game.TargetRef{cardRef(bear)},
	})
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	answerSearchByID(t, g, me.ID, ground)
	passPriorityAroundTable(t, g)
	if auraHostChoiceFor(g, me.ID) != nil {
		t.Fatal("the effect names the host: nothing to ask")
	}
	if !me.Library.Contains(ground) || g.Battlefield.Contains(ground) || me.Graveyard.Contains(ground) {
		t.Fatalf("an enchant-land Aura can't enchant Light-Paws (%v) and stays in the library", paws)
	}
}

// A copy of an Aura spell becomes a token attached to the copy's target
// (CR 707.10, 608.3f). The copy NAMES its host, so it is not asked CR
// 303.4f's question even with several legal hosts on the board.
func TestACopyOfAnAuraSpellEntersAttachedToItsTargetWithoutAsking(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := pushTribalCreature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	pushTribalCreature(g, opp.ID, "Other Bear", "Creature — Bear", 2, 2)
	spell := castFromHand(t, g, game.Card{Name: "Pacifism", TypeLine: auraTypeLine, OracleID: pacifismOracle},
		game.CastSpellParams{Targets: []game.TargetRef{cardRef(bear)}})
	g.WithWriteLock(func() {
		if err := g.CopySpellForEffect(spell, me.ID, false, nil); err != nil {
			t.Fatalf("copy: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	if auraHostChoiceFor(g, me.ID) != nil {
		t.Fatal("the copy names its host: nothing to ask")
	}
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Pacifism" && c.IsAttachedTo(bear) {
			n++
		}
	}
	if n != 2 {
		t.Fatalf("the Aura and its token copy both enchant the target, got %d", n)
	}
}

// Gift of Immortality names its host too, and is not asked.
func TestGiftOfImmortalityReturnsAttachedWithoutAsking(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := auraBear(g, me.ID)
	pushTribalCreature(g, me.ID, "Other Creature", "Creature — Elf", 1, 1)
	gift := auraCast(t, g, "Gift of Immortality", giftOfImmortalityOracle, bear)
	killCreature(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if auraHostChoiceFor(g, me.ID) != nil {
		t.Fatal("Gift names its host: nothing to ask")
	}
	ids := battlefieldIDsNamed(g, "Bear")
	if len(ids) != 1 || attachmentHostOf(t, g, gift) != cardRef(ids[0]) {
		t.Fatalf("Gift returns attached to the creature it brought back: %+v", attachmentHostOf(t, g, gift))
	}
}
