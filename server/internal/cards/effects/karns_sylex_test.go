package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// karns_sylex_test.go — #5660. Karn's Sylex on the real catalog card:
// the tapped entry, "Players can't pay life to cast spells or to
// activate abilities that aren't mana abilities" at every life cost a
// cast or an activation can carry, the payments it leaves alone, and
// the {X} sweep.

const (
	karnsSylexOracle  = "c5e817f9-a003-45b5-8e16-1746e6096166"
	birthingPodOracle = "f8b9dd54-0837-47f4-ad14-7a0322d46d5f"
	fetidHeathOracle  = "42bf259d-4bb9-49c3-b4ec-223dca62f4d6"
)

// sylexFor puts an untapped Karn's Sylex on the battlefield under
// `owner`.
func sylexFor(g *game.Game, owner uuid.UUID) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Karn's Sylex", OracleID: karnsSylexOracle,
		TypeLine: "Legendary Artifact", ManaCost: "{3}", Owner: owner, Controller: owner,
	})
}

// sylexTapped reports whether the battlefield permanent `id` is tapped,
// and whether it is there at all.
func sylexTapped(g *game.Game, id uuid.UUID) (tapped, found bool) {
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			return g.Battlefield.Cards[i].Tapped, true
		}
	}
	return false, false
}

// sylexRemove takes a permanent off the battlefield without ceremony.
func sylexRemove(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneExile}, id); err != nil {
		t.Fatalf("removing the permanent: %v", err)
	}
}

func TestKarnsSylexIsCompleteWithNoCaveat(t *testing.T) {
	spec, ok := Lookup(karnsSylexOracle)
	if !ok {
		t.Fatal("Karn's Sylex is not in the catalog")
	}
	if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("completeness %v, caveats %v; want Full with none", spec.Completeness, spec.Caveats)
	}
}

func TestKarnsSylexEntersTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	id := handCardFull(me, "Karn's Sylex", "Legendary Artifact", "{3}", karnsSylexOracle, nil)
	fillPool(me, 3)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("casting Karn's Sylex: %v", err)
	}
	passPriorityAroundTable(t, g)
	tapped, found := sylexTapped(g, id)
	if !found {
		t.Fatal("Karn's Sylex is not on the battlefield")
	}
	if !tapped {
		t.Error("Karn's Sylex entered untapped")
	}
	// Tapped, so its own ability cannot be paid this turn.
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("a tapped Sylex paid its {T}")
	}
}

// An opponent's Sylex stops K'rrik's life-for-mana on a spell, costs the
// caster nothing for trying, and leaves the mana payment alone. With the
// Sylex gone the same claim is accepted.
func TestKarnsSylexStopsLifeForManaOnASpell(t *testing.T) {
	g, me, _, _ := krrikTable(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	sylex := sylexFor(g, opp.ID)
	cheap := wpCreature(g, opp.ID, "Two Drop", "{1}{B}")
	spell := krrikHandSpell(me, "Bloodchief's Thirst", "Sorcery", bloodchiefsThirstOracle, "{B}", "B")
	life := me.Life

	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, Targets: twTarget(cheap), PhyrexianLife: 1}); err == nil {
		t.Fatal("a {B} was paid with 2 life under Karn's Sylex")
	}
	if me.Life != life || !me.Hand.Contains(spell) {
		t.Fatalf("the refused cast changed something: life %d -> %d, in hand %v", life, me.Life, me.Hand.Contains(spell))
	}
	v := protocol.ViewOfGameFor(g, me.ID.String())
	for _, s := range v.Seats {
		for _, c := range s.Hand.Cards {
			if c.InstanceID == spell.String() && c.PhyrexianSymbols != 0 {
				t.Errorf("the view offers %d symbol(s) payable with life under Karn's Sylex", c.PhyrexianSymbols)
			}
		}
	}

	sylexRemove(t, g, sylex)
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, Targets: twTarget(cheap), PhyrexianLife: 1}); err != nil {
		t.Fatalf("with the Sylex gone the {B} is payable with life again: %v", err)
	}
	if me.Life != life-2 {
		t.Errorf("life = %d, want %d", me.Life, life-2)
	}
}

func TestKarnsSylexLeavesTheManaPaymentOfAPhyrexianCostAlone(t *testing.T) {
	g, me, _, _ := krrikTable(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	sylexFor(g, me.ID)
	cheap := wpCreature(g, opp.ID, "Two Drop", "{1}{B}")
	spell := krrikHandSpell(me, "Bloodchief's Thirst", "Sorcery", bloodchiefsThirstOracle, "{B}", "B")
	fillPoolColored(me, "B", 1)
	life := me.Life
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, Targets: twTarget(cheap)}); err != nil {
		t.Fatalf("{B} paid with mana under Karn's Sylex: %v", err)
	}
	if me.Life != life {
		t.Errorf("a mana payment cost %d life", life-me.Life)
	}
}

// An alternative cost's life (Snuff Out's "pay 4 life") is a cost to cast
// the spell: neither offered nor accepted.
func TestKarnsSylexStopsAnAlternativeCostsLife(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	toMainForCost(t, g)
	snuff := visibleHandCard(me, "Snuff Out", "Instant", "{3}{B}", snuffOutOracle, []string{"B"})
	freeSpellPermanent(g, me.ID, "Swamp", "Basic Land — Swamp")
	bear := freeSpellPermanent(g, them.ID, "Grizzly Bears", "Creature — Bear")
	target := []game.TargetRef{{Kind: game.TargetCard, ID: bear}}

	if keys := offeredKeys(t, g, me, snuff); !hasKey(keys, "pay_life") {
		t.Fatalf("without the Sylex the pay_life offer is missing: %v", keys)
	}
	sylexFor(g, them.ID)
	if keys := offeredKeys(t, g, me, snuff); hasKey(keys, "pay_life") {
		t.Errorf("under Karn's Sylex the pay_life offer is still shown: %v", keys)
	}
	life := me.Life
	if err := g.CastSpell(me.ID, snuff, game.CastSpellParams{
		Strict: true, AlternativeCost: "pay_life", Targets: target,
	}); err == nil {
		t.Error("Snuff Out was cast for 4 life under Karn's Sylex")
	}
	if me.Life != life {
		t.Errorf("the refused cast cost %d life", life-me.Life)
	}
}

// "Pay X life" (Toxic Deluge): X = 0 is no payment and stays legal; any
// X above it is refused.
func TestKarnsSylexHoldsPayXLifeToZero(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	sylexFor(g, me.ID)
	deluge := handCardFull(me, "Toxic Deluge", "Sorcery", "{2}{B}", toxicDelugeOracle, []string{"B"})
	fillPoolColored(me, "B", 3)
	life := me.Life
	if err := g.CastSpell(me.ID, deluge, game.CastSpellParams{Strict: true, XValue: 2}); err == nil {
		t.Fatal("Toxic Deluge paid 2 life under Karn's Sylex")
	}
	if me.Life != life {
		t.Fatalf("the refused cast cost %d life", life-me.Life)
	}
	if err := g.CastSpell(me.ID, deluge, game.CastSpellParams{Strict: true, XValue: 0}); err != nil {
		t.Fatalf("Toxic Deluge with X = 0 pays no life and is castable: %v", err)
	}
	if me.Life != life {
		t.Errorf("X = 0 cost %d life", life-me.Life)
	}
}

// A targeting tax paid in life (Terror of the Peaks) is an additional
// cost to cast: the spell can't be aimed at the Dragon, and can be aimed
// elsewhere.
func TestKarnsSylexStopsATargetingTaxPaidInLife(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	terror := pushCatalogPermanent(g, opp.ID, "Terror of the Peaks", "Creature — Dragon", b04TerrorOfThePeaksOracle, false)
	other := pushCatalogPermanent(g, opp.ID, "Bystander", "Creature — Bear", "", false)
	sylexFor(g, opp.ID)
	advanceToMain(t, g)
	life := me.Life
	if _, err := castBoltAtFrom(t, g, me, cardRef(terror)); err == nil {
		t.Fatal("a Bolt at Terror of the Peaks paid its 3 life under Karn's Sylex")
	}
	if me.Life != life {
		t.Fatalf("the refused cast cost %d life", life-me.Life)
	}
	if _, err := castBoltAtFrom(t, g, me, cardRef(other)); err != nil {
		t.Fatalf("a Bolt at an untaxed creature: %v", err)
	}
}

// A printed life cost on an ability that isn't a mana ability (War
// Room's draw) is refused, with nothing tapped or spent.
func TestKarnsSylexStopsAnActivatedAbilitysLifeCost(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	clrCommander(me, "Mono", "G")
	room := clrWarRoom(t, g, me)
	sylex := sylexFor(g, g.Seats[1].ID)
	life, hand := me.Life, me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, room, warRoomDraw, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("War Room's draw paid life under Karn's Sylex")
	}
	if tapped, _ := sylexTapped(g, room); me.Life != life || me.Hand.Size() != hand || tapped {
		t.Fatalf("the refused activation changed something: life %d -> %d, hand %d -> %d", life, me.Life, hand, me.Hand.Size())
	}
	sylexRemove(t, g, sylex)
	clrWarRoomCosts(t, g, me, room, 1)
}

// The life half of a Phyrexian symbol in an activation cost (Birthing
// Pod's {G/P}) is refused, and the view offers no such payment.
func TestKarnsSylexStopsPhyrexianLifeOnAnActivation(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	pod := b12Push(g, me.ID, "Birthing Pod", "Artifact", birthingPodOracle, 0, 0)
	fodder := b12Creature(g, me.ID, "Fodder", "Creature — Goblin", 1, 1)
	sylex := sylexFor(g, me.ID)
	fillPool(me, 1)
	life := me.Life
	defer func() {
		// The control: with the Sylex gone the same announcement is paid.
		sylexRemove(t, g, sylex)
		if err := g.ActivateCatalogAbility(me.ID, pod, 0, game.ActivateAbilityParams{
			Strict: true, PhyrexianLife: 1, SacrificeIDs: []uuid.UUID{fodder},
		}); err != nil {
			t.Errorf("with the Sylex gone Birthing Pod's {G/P} is payable with life: %v", err)
		} else if me.Life != life-2 {
			t.Errorf("life = %d, want %d", me.Life, life-2)
		}
	}()
	if err := g.ActivateCatalogAbility(me.ID, pod, 0, game.ActivateAbilityParams{
		Strict: true, PhyrexianLife: 1, SacrificeIDs: []uuid.UUID{fodder},
	}); err == nil {
		t.Fatal("Birthing Pod's {G/P} was paid with 2 life under Karn's Sylex")
	}
	if me.Life != life || !g.Battlefield.Contains(fodder) {
		t.Fatalf("the refused activation changed something: life %d -> %d", life, me.Life)
	}
	v := protocol.ViewOfGameFor(g, me.ID.String())
	for _, c := range v.Battlefield.Cards {
		if c.InstanceID != pod.String() {
			continue
		}
		for _, a := range c.ActivatedAbilities {
			if a.PhyrexianSymbols != 0 {
				t.Errorf("the view offers %d symbol(s) payable with life on Birthing Pod", a.PhyrexianSymbols)
			}
		}
	}
}

// Mana abilities are exempt as printed: Mana Confluence's "Pay 1 life"
// and a filter land's {B} paid with life under K'rrik both still work.
func TestKarnsSylexLeavesManaAbilitiesAlone(t *testing.T) {
	g, me, _, _ := krrikTable(t)
	sylexFor(g, me.ID)
	land := seedPermanentWithOracle(g, me.ID, "Mana Confluence", "Land", manaConfluenceOracle)
	life := me.Life
	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("Mana Confluence under Karn's Sylex: %v", err)
	}
	if me.Life != life-1 {
		t.Errorf("Mana Confluence: life %d -> %d, want %d", life, me.Life, life-1)
	}
	heath := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Fetid Heath", OracleID: fetidHeathOracle,
		TypeLine: "Land", Owner: me.ID, Controller: me.ID,
	})
	life = me.Life
	if err := g.ActivateManaAbility(me.ID, heath, 1, game.ManaAbilityParams{
		Colors: []string{"B", "B"}, PhyrexianLife: 1,
	}); err != nil {
		t.Fatalf("Fetid Heath's filter paid with 2 life under Karn's Sylex: %v", err)
	}
	if me.Life != life-game.PhyrexianLifePerSymbol {
		t.Errorf("Fetid Heath: life %d -> %d, want %d", life, me.Life, life-game.PhyrexianLifePerSymbol)
	}
}

// The rule is read off the battlefield, so it is live on a restored
// game with nothing stored for it.
func TestKarnsSylexRuleSurvivesARestore(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	clrCommander(me, "Mono", "G")
	room := clrWarRoom(t, g, me)
	sylexFor(g, g.Seats[1].ID)
	restored := restoreThroughJSON(t, g)
	if err := restored.ActivateCatalogAbility(me.ID, room, warRoomDraw, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("War Room's draw paid life on the restored game")
	}
	var live bool
	restored.WithWriteLock(func() { live = restored.CantPayLifeToCastOrActivateForEffect() })
	if !live {
		t.Error("the restored game lost the Sylex's rule")
	}
}

// {X}, {T}, Exile Karn's Sylex: every nonland permanent with mana value
// X or less is destroyed, tokens included; lands and anything dearer
// stay; the Sylex is exiled as the cost.
func TestKarnsSylexSweepsNonlandPermanentsUpToX(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	sylex := sylexFor(g, me.ID)
	two := wpCreature(g, opp.ID, "Two Drop", "{1}{B}")
	mine := wpCreature(g, me.ID, "My Two Drop", "{2}")
	three := wpCreature(g, opp.ID, "Three Drop", "{2}{B}")
	token := b12Creature(g, opp.ID, "Soldier", "Token Creature — Soldier", 1, 1)
	land := freeSpellPermanent(g, opp.ID, "Swamp", "Basic Land — Swamp")
	fillPool(me, 2)

	if err := g.ActivateCatalogAbility(me.ID, sylex, 0, game.ActivateAbilityParams{Strict: true, XValue: 2}); err != nil {
		t.Fatalf("activating Karn's Sylex for X = 2: %v", err)
	}
	if g.Battlefield.Contains(sylex) || !g.Exile.Contains(sylex) {
		t.Fatal("Karn's Sylex was not exiled as a cost")
	}
	if !g.Battlefield.Contains(two) {
		t.Fatal("the sweep happened before the ability resolved")
	}
	passPriorityAroundTable(t, g)
	for name, id := range map[string]uuid.UUID{"an opponent's two-drop": two, "the controller's two-drop": mine, "a token": token} {
		if g.Battlefield.Contains(id) {
			t.Errorf("%s survived X = 2", name)
		}
	}
	for name, id := range map[string]uuid.UUID{"a three-drop": three, "a land": land} {
		if !g.Battlefield.Contains(id) {
			t.Errorf("%s was destroyed at X = 2", name)
		}
	}
}

// "Activate only as a sorcery": refused on another player's turn.
func TestKarnsSylexActivatesOnlyAsASorcery(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	sylex := sylexFor(g, opp.ID)
	fillPool(opp, 2)
	if err := g.ActivateCatalogAbility(opp.ID, sylex, 0, game.ActivateAbilityParams{Strict: true, XValue: 2}); err == nil {
		t.Error("Karn's Sylex was activated on another player's turn")
	}
	if !g.Battlefield.Contains(sylex) {
		t.Error("the refused activation exiled the Sylex")
	}
}
