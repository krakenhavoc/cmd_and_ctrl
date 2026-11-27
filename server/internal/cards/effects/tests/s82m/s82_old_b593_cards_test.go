package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	o593Maga     = "1fb1afc0-a3bf-4368-9ca0-47dd8c4c3c60"
	o593Descend  = "d5bfb740-8ab2-4160-a8f0-5b70b63c2f3a"
	o593Naginata = "5641c831-5254-4ef2-ba48-328d86007479"
	o593Manriki  = "cc2eb3e4-084c-471b-9fe0-72aa9ff2019a"
	o593Ashes    = "6437867b-a658-4a71-a8a5-0c5a782d1f84"
	o593Kirin    = "a5f7967e-d357-4ab6-96ac-77262864f280"
	o593Tomb     = "f002be6a-e459-49c4-b765-062e30107439"
)

func TestO593CardsAreFullWithNoCaveats(t *testing.T) {
	for oracle, name := range map[string]string{
		o593Maga: "Maga, Traitor to Mortals", o593Descend: "Descendant of Masumaro", o593Naginata: "O-Naginata",
		o593Manriki: "Manriki-Gusari", o593Ashes: "Ashes of the Fallen", o593Kirin: "Celestial Kirin", o593Tomb: "Tomb of Urami",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered %v as %q (%v, caveats %v)", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestO593MagaDrainsForItsCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	before := opp.Life
	id := effectstest.AssistCardCastX(t, g, "Maga, Traitor to Mortals", "Legendary Creature — Human Wizard", o593Maga, "{X}{B}{B}{B}",
		[]string{"B", "B", "B", "B", "B"}, 2, nil)
	pick := game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}
	effectstest.AssistCardSettle(t, g, me.ID, &pick, true)
	if got := effectstest.CounterCount(g, id, game.CounterPlusOne); got != 2 {
		t.Fatalf("Maga has %d +1/+1 counters, want 2", got)
	}
	if got := before - opp.Life; got != 2 {
		t.Errorf("the target lost %d life, want 2", got)
	}
}

func TestO593ONaginataAttachesOnlyToPowerThreeOrMore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	nag := effectstest.PushCatalogPermanent(g, me.ID, "O-Naginata", "Artifact — Equipment", o593Naginata, false)
	small := effectstest.B12Creature(g, me.ID, "Small", "Creature — Bear", 2, 2)
	big := effectstest.B12Creature(g, me.ID, "Big", "Creature — Giant", 3, 3)
	effectstest.AdvanceToMain(t, g)
	if err := effectstest.StrictEquip(t, g, me.ID, nag, small, "{C}{C}"); err == nil {
		effectstest.PassPriorityAroundTable(t, g)
	}
	if effectstest.AttachedTo(t, g, nag, small) {
		t.Fatal("O-Naginata attached to a creature with power 2")
	}
	if err := effectstest.StrictEquip(t, g, me.ID, nag, big, "{C}{C}"); err != nil {
		t.Fatalf("equip to power 3: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.AttachedTo(t, g, nag, big) {
		t.Fatal("O-Naginata did not attach to a creature with power 3")
	}
	if p := effectstest.EffectivePower(t, g, big); p != 6 {
		t.Errorf("the equipped creature has power %d, want 6", p)
	}
	effectstest.AssertKeywords(t, g, big, "trample")
}

func TestO593ManrikiGusariGrantsADestroyEquipmentAbility(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	man := effectstest.PushCatalogPermanent(g, me.ID, "Manriki-Gusari", "Artifact — Equipment", o593Manriki, false)
	other := effectstest.PushCatalogPermanent(g, opp.ID, "O-Naginata", "Artifact — Equipment", o593Naginata, false)
	host := effectstest.B12Creature(g, me.ID, "Host", "Creature — Monk", 1, 1)
	effectstest.AdvanceToMain(t, g)
	if err := effectstest.StrictEquip(t, g, me.ID, man, host, "{C}"); err != nil {
		t.Fatalf("equip: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, host); p != 2 {
		t.Errorf("power %d, want 2", p)
	}
	idx, ref := effectstest.BeGrantedRow(t, g, host, "{T}: Destroy target Equipment")
	if idx < 0 {
		t.Fatal("the equipped creature has no granted destroy-Equipment row")
	}
	if err := g.ActivateCatalogAbility(me.ID, host, idx, game.ActivateAbilityParams{Ref: ref, Strict: true,
		Targets: effectstest.CardRefs(host)}); err == nil {
		t.Fatal("the row took a non-Equipment target")
	}
	if err := g.ActivateCatalogAbility(me.ID, host, idx, game.ActivateAbilityParams{Ref: ref, Strict: true,
		Targets: effectstest.CardRefs(other)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(other) {
		t.Error("the targeted Equipment survived")
	}
}

func TestO593AshesGivesGraveyardCreaturesTheChosenType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.PushGraveyardPermanent(me, "Dead Bear", "Creature — Bear", "{1}{G}")
	theirs := effectstest.PushGraveyardPermanent(opp, "Their Bear", "Creature — Bear", "{1}{G}")
	spell := effectstest.PushGraveyardPermanent(me, "Old Spell", "Sorcery", "{1}")
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Ashes of the Fallen", "Artifact", o593Ashes, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerCreatureType(t, g, me.ID, "Ogre")
	g.RunStateChecksForTest()
	has := func(p *game.Player, id uuid.UUID, sub string) bool {
		for _, c := range p.Graveyard.Cards {
			if c.InstanceID == id {
				return c.HasSubtype(sub)
			}
		}
		t.Fatalf("card %s is not in %s's graveyard", id, p.ID)
		return false
	}
	if !has(me, mine, "Ogre") || !has(me, mine, "Bear") {
		t.Error("the creature card in your graveyard is not a Bear and an Ogre")
	}
	if has(opp, theirs, "Ogre") {
		t.Error("an opponent's graveyard creature got the type")
	}
	if has(me, spell, "Ogre") {
		t.Error("a non-creature card in your graveyard got the type")
	}
}

func TestO593KirinDestroysEveryPermanentOfTheSpellsManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	kirin := effectstest.B12Push(g, me.ID, "Celestial Kirin", "Legendary Creature — Kirin Spirit", o593Kirin, 3, 3)
	two := effectstest.ApcB1Push(g, opp.ID, "Two Drop", "Creature — Bear", "", "{1}{G}", 2, 2)
	three := effectstest.ApcB1Push(g, opp.ID, "Three Drop", "Creature — Bear", "", "{2}{G}", 2, 2)
	effectstest.AdvanceToMain(t, g)
	// A Spirit spell with mana value 2.
	effectstest.P1g2Cast(t, g, "Small Spirit", "Creature — Spirit", "", "{1}{W}", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(two) {
		t.Error("the mana value 2 permanent survived")
	}
	if !g.Battlefield.Contains(three) {
		t.Error("the mana value 3 permanent was destroyed")
	}
	if !g.Battlefield.Contains(kirin) {
		t.Error("the Kirin (mana value 4) was destroyed by a mana value 2 spell")
	}
}

func TestO593KirinIgnoresASpellThatIsNeitherSpiritNorArcane(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Celestial Kirin", "Legendary Creature — Kirin Spirit", o593Kirin, 3, 3)
	two := effectstest.ApcB1Push(g, opp.ID, "Two Drop", "Creature — Bear", "", "{1}{G}", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.P1g2Cast(t, g, "Plain Bear", "Creature — Bear", "", "{1}{G}", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(two) {
		t.Error("a Bear spell set the Kirin off")
	}
}

func TestO593DescendantNetsItsHandAgainTheirs(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	d := effectstest.B12Push(g, me.ID, "Descendant of Masumaro", "Creature — Human Monk", o593Descend, 1, 1)
	for i := 0; i < 3; i++ {
		effectstest.HandSpell(me, "Mine", "Sorcery", "{1}")
	}
	for i := 0; i < 1; i++ {
		effectstest.HandSpell(opp, "Theirs", "Sorcery", "{1}")
	}
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	pick := game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}
	effectstest.AssistCardSettle(t, g, me.ID, &pick, true)
	hand, oppHand := me.Hand.Size(), opp.Hand.Size()
	want := hand - oppHand
	if got := effectstest.CounterCount(g, d, game.CounterPlusOne); got != want {
		t.Errorf("the creature has %d +1/+1 counters, want %d (hands %d and %d)", got, want, hand, oppHand)
	}
}

func TestO593TombOfUramiHurtsYouUnlessYouControlAnOgre(t *testing.T) {
	for _, ogre := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		tomb := effectstest.PushCatalogPermanent(g, me.ID, "Tomb of Urami", "Legendary Land", o593Tomb, false)
		if ogre {
			effectstest.B12Creature(g, me.ID, "Ogre Friend", "Creature — Ogre", 2, 2)
		}
		effectstest.AdvanceToMain(t, g)
		life := me.Life
		effectstest.TapForMana(t, g, me.ID, tomb)
		lost := life - me.Life
		if ogre && lost != 0 || !ogre && lost != 1 {
			t.Errorf("ogre=%v: lost %d life", ogre, lost)
		}
		if got := effectstest.SortedPool(me); len(got) != 1 || got[0] != "B" {
			t.Errorf("ogre=%v: pool %v, want one B", ogre, got)
		}
	}
}

func TestO593TombOfUramiSacrificesEveryLandForUrami(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	tomb := effectstest.PushCatalogPermanent(g, me.ID, "Tomb of Urami", "Legendary Land", o593Tomb, false)
	other := effectstest.PushCatalogPermanent(g, me.ID, "Swamp", "Basic Land — Swamp", "", false)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "B", "B", "B", "B")
	effectstest.B16Activate(t, g, me.ID, tomb, 0, game.ActivateAbilityParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(tomb) || g.Battlefield.Contains(other) {
		t.Error("a land survived the sacrifice")
	}
	var urami *game.Card
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Name == "Urami" {
			c := c
			urami = &c
		}
	}
	if urami == nil {
		t.Fatal("no Urami token")
	}
	if urami.Power != 5 || urami.Toughness != 5 || !urami.HasSubtype("Demon") || !urami.HasSubtype("Spirit") || !urami.IsLegendary() {
		t.Errorf("Urami is %q %d/%d", urami.TypeLine, urami.Power, urami.Toughness)
	}
	effectstest.AssertKeywords(t, g, urami.InstanceID, "flying")
}
