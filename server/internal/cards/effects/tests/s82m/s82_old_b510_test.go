package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b510_test.go — Shards of Alara, "Activated abilities" (S82 OLD batch
// 510, #5588): each card driven through a real activation.

const (
	oldB510PuppetOracle     = "06494c26-3822-43ee-bedf-611c9ec7cc6b"
	oldB510GodsireOracle    = "4252d7f4-c582-4a21-b3af-829c84cb0b09"
	oldB510AdeptOracle      = "44d30168-8203-4a8e-a11f-b4a068231e55"
	oldB510KathariOracle    = "5bcb63b4-6236-45c5-884d-188f8f5e8f02"
	oldB510BantOracle       = "92d8b956-87db-43e4-bbb5-436ba01b5714"
	oldB510DrakeOracle      = "95500513-5519-4131-bb89-42383289f2bd"
	oldB510GustriderOracle  = "a358fe8f-6ec2-4b42-8851-ef65308ccb1d"
	oldB510MosstodonOracle  = "a41b610e-125d-4921-81f9-6391f9214d12"
	oldB510ViashinoOracle   = "a85e893c-6027-42eb-8d05-14c745cc8c95"
	oldB510TaunterOracle    = "b59da333-e658-447e-a8bd-8c0124d8b905"
	oldB510SnapperOracle    = "c5e2ed91-edd2-47ec-809e-3e4a4245e2ce"
	oldB510SagesOracle      = "d6e78c6a-7555-4818-8054-1cdc0fca92e6"
	oldB510SilencersOracle  = "e132bba0-40c5-4095-99a3-02cad520ca99"
	oldB510OozeGardenOracle = "ed1e065d-93b8-4017-9e12-427443548fc5"
)

func TestS82OldB510AllAreFullWithNoCaveats(t *testing.T) {
	for _, id := range []string{oldB510PuppetOracle, oldB510GodsireOracle, oldB510AdeptOracle, oldB510KathariOracle,
		oldB510BantOracle, oldB510DrakeOracle, oldB510GustriderOracle, oldB510MosstodonOracle, oldB510ViashinoOracle,
		oldB510TaunterOracle, oldB510SnapperOracle, oldB510SagesOracle, oldB510SilencersOracle, oldB510OozeGardenOracle} {
		spec, ok := effects.Lookup(id)
		if !ok {
			t.Errorf("%s is not registered", id)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v, caveats %v", spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func oldB510Ready(g *game.Game, id uuid.UUID) {
	effectstest.UntapForTest(g, id)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
}

// oldB510Perm seeds a ready catalog permanent for the active seat.
func oldB510Perm(t *testing.T, g *game.Game, name, typeLine, oracle string) uuid.UUID {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	id := effectstest.SeedPermanentWithOracle(g, g.Seats[0].ID, name, typeLine, oracle)
	oldB510Ready(g, id)
	return id
}

func oldB510Effective(g *game.Game, id uuid.UUID) game.Card {
	c, _ := g.LookupCardForEffect(id)
	return c
}

func oldB510Activate(t *testing.T, g *game.Game, src uuid.UUID, ability int, mana string, params game.ActivateAbilityParams) error {
	t.Helper()
	me := g.Seats[0]
	if mana != "" {
		effectstest.FloatMana(t, g, me, mana)
	}
	params.Strict = true
	return g.ActivateCatalogAbility(me.ID, src, ability, params)
}

func TestS82OldB510PuppetConjurerMakesAHomunculusAndSacrificesOneEachUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	conj := oldB510Perm(t, g, "Puppet Conjurer", "Artifact Creature — Human Wizard", oldB510PuppetOracle)
	if err := oldB510Activate(t, g, conj, 0, "{U}", game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	var tok *game.Card
	for i := range g.Battlefield.Cards {
		c := g.Battlefield.Cards[i]
		if c.IsToken() && c.Name == "Homunculus" {
			tok = &g.Battlefield.Cards[i]
		}
	}
	if tok == nil {
		t.Fatal("no Homunculus token")
	}
	if tok.CurrentPower() != 0 || tok.CurrentToughness() != 1 || !tok.HasCardType("artifact") || !tok.HasColor("U") {
		t.Errorf("token = %s %d/%d", tok.TypeLine, tok.CurrentPower(), tok.CurrentToughness())
	}
	// Upkeep: sacrifice a Homunculus.
	spec, _ := effects.Lookup(oldB510PuppetOracle)
	if len(spec.Triggered) != 1 {
		t.Fatalf("triggers %d, want 1 upkeep trigger", len(spec.Triggered))
	}
	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	conj2 := oldB510Perm(t, g2, "Puppet Conjurer", "Artifact Creature — Human Wizard", oldB510PuppetOracle)
	homunculus := effectstest.B12Creature(g2, me2.ID, "Homunculus", "Token Creature — Homunculus", 0, 1)
	_ = me
	effectstest.AdvanceToUpkeepOfSeat(t, g2, 0)
	effectstest.PassPriorityAroundTable(t, g2)
	effectstest.AnswerChooseCards(t, g2, me2.ID, homunculus)
	if g2.Battlefield.Contains(homunculus) {
		t.Error("the Homunculus was not sacrificed at upkeep")
	}
	if !g2.Battlefield.Contains(conj2) {
		t.Error("the Conjurer itself was sacrificed (it is not a Homunculus)")
	}
}

func TestS82OldB510GodsireMakesAnEightEightBeast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	god := oldB510Perm(t, g, "Godsire", "Creature — Beast", oldB510GodsireOracle)
	if err := oldB510Activate(t, g, god, 0, "", game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	found := false
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() && c.Name == "Beast" && c.CurrentPower() == 8 && c.CurrentToughness() == 8 &&
			c.HasColor("R") && c.HasColor("G") && c.HasColor("W") {
			found = true
		}
	}
	if !found {
		t.Error("no 8/8 red, green and white Beast token")
	}
}

func TestS82OldB510CatharticAdeptMillsTheTargetPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	adept := oldB510Perm(t, g, "Cathartic Adept", "Creature — Human Wizard", oldB510AdeptOracle)
	lib, gy := opp.Library.Size(), opp.Graveyard.Size()
	if err := oldB510Activate(t, g, adept, 0, "", game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Library.Size() != lib-1 || opp.Graveyard.Size() != gy+1 {
		t.Errorf("library %d->%d graveyard %d->%d, want one card milled", lib, opp.Library.Size(), gy, opp.Graveyard.Size())
	}
}

func TestS82OldB510SkeletalKathariRegeneratesBySacrificingACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	kath := oldB510Perm(t, g, "Skeletal Kathari", "Creature — Bird Skeleton", oldB510KathariOracle)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	if err := oldB510Activate(t, g, kath, 0, "", game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}}); err == nil {
		t.Fatal("activated without paying {B}")
	}
	if err := oldB510Activate(t, g, kath, 0, "{B}", game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrificed creature stayed")
	}
	if !(g.RegenerationShieldsOn(kath) > 0) {
		t.Error("the Kathari has no regeneration shield")
	}
}

func TestS82OldB510ViashinoSkeletonRegeneratesByDiscarding(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	skel := oldB510Perm(t, g, "Viashino Skeleton", "Creature — Lizard Skeleton", oldB510ViashinoOracle)
	card := effectstest.HandSpell(me, "Spare", "Sorcery", "")
	hand := me.Hand.Size()
	effectstest.FloatMana(t, g, me, "{C}{B}")
	if err := g.ActivateCatalogAbility(me.ID, skel, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{card}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand-1 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand-1)
	}
	if !(g.RegenerationShieldsOn(skel) > 0) {
		t.Error("the Skeleton has no regeneration shield")
	}
}

func TestS82OldB510BantBattlemageGrantsTrampleOrFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	mage := oldB510Perm(t, g, "Bant Battlemage", "Creature — Human Wizard", oldB510BantOracle)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if err := oldB510Activate(t, g, mage, 0, "{G}", game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}); err != nil {
		t.Fatalf("trample: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.P1g2HasKeyword(g, bear, "trample") {
		t.Error("no trample")
	}
	oldB510Ready(g, mage)
	if err := oldB510Activate(t, g, mage, 1, "{U}", game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}); err != nil {
		t.Fatalf("flying: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.P1g2HasKeyword(g, bear, "flying") {
		t.Error("no flying")
	}
}

func TestS82OldB510SelfGrantRows(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle, mana, keyword string }{
		{"Cloudheath Drake", "Artifact Creature — Drake", oldB510DrakeOracle, "{C}{W}", "vigilance"},
		{"Shore Snapper", "Creature — Beast", oldB510SnapperOracle, "{U}", "islandwalk"},
		{"Vectis Silencers", "Artifact Creature — Human Rogue", oldB510SilencersOracle, "{C}{C}{B}", "deathtouch"},
	} {
		g := effectstest.NewCatalogGame(t)
		id := oldB510Perm(t, g, tc.name, tc.typeLine, tc.oracle)
		if effectstest.P1g2HasKeyword(g, id, tc.keyword) {
			t.Fatalf("%s starts with %s", tc.name, tc.keyword)
		}
		if err := oldB510Activate(t, g, id, 0, "", game.ActivateAbilityParams{}); err == nil {
			t.Errorf("%s: activated without mana", tc.name)
		}
		if err := oldB510Activate(t, g, id, 0, tc.mana, game.ActivateAbilityParams{}); err != nil {
			t.Fatalf("%s: activate: %v", tc.name, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if !effectstest.P1g2HasKeyword(g, id, tc.keyword) {
			t.Errorf("%s did not gain %s", tc.name, tc.keyword)
		}
	}
}

func TestS82OldB510PowerFiveTargetsOnly(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, mana, keyword string
		tap                                   bool
	}{
		{"Mosstodon", "Creature — Plant Elephant", oldB510MosstodonOracle, "{C}", "trample", false},
		{"Bloodthorn Taunter", "Creature — Human Scout", oldB510TaunterOracle, "", "haste", true},
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		src := oldB510Perm(t, g, tc.name, tc.typeLine, tc.oracle)
		small := effectstest.B12Creature(g, me.ID, "Small", "Creature — Bear", 4, 4)
		big := effectstest.B12Creature(g, me.ID, "Big", "Creature — Wurm", 5, 5)
		if err := oldB510Activate(t, g, src, 0, tc.mana, game.ActivateAbilityParams{
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: small}}}); err == nil {
			t.Errorf("%s: a 4-power creature was a legal target", tc.name)
		}
		if err := oldB510Activate(t, g, src, 0, tc.mana, game.ActivateAbilityParams{
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: big}}}); err != nil {
			t.Fatalf("%s: activate: %v", tc.name, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if !effectstest.P1g2HasKeyword(g, big, tc.keyword) {
			t.Errorf("%s: the 5-power creature did not gain %s", tc.name, tc.keyword)
		}
	}
}

func TestS82OldB510FiligreeSagesUntapsAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	sages := oldB510Perm(t, g, "Filigree Sages", "Artifact Creature — Vedalken Wizard", oldB510SagesOracle)
	rock := effectstest.SeedPermanent(g, me.ID, "Rock", "Artifact")
	effectstest.TapForTest(t, g, rock)
	if err := oldB510Activate(t, g, sages, 0, "{C}{C}{U}", game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: rock}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if oldB510Effective(g, rock).Tapped {
		t.Error("the artifact is still tapped")
	}
}

func TestS82OldB510GustriderGivesBigCreaturesFlyingWhenSacrificed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	gust := oldB510Perm(t, g, "Gustrider Exuberant", "Creature — Human Wizard", oldB510GustriderOracle)
	big := effectstest.B12Creature(g, me.ID, "Big", "Creature — Wurm", 5, 5)
	small := effectstest.B12Creature(g, me.ID, "Small", "Creature — Bear", 4, 4)
	if err := oldB510Activate(t, g, gust, 0, "", game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(gust) {
		t.Error("the Exuberant was not sacrificed")
	}
	if !effectstest.P1g2HasKeyword(g, big, "flying") {
		t.Error("the 5-power creature did not gain flying")
	}
	if effectstest.P1g2HasKeyword(g, small, "flying") {
		t.Error("the 4-power creature gained flying")
	}
}

func TestS82OldB510OozeGardenMakesAnOozeAsBigAsTheSacrificedPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	garden := effectstest.SeedPermanentWithOracle(g, me.ID, "Ooze Garden", "Enchantment", oldB510OozeGardenOracle)
	ooze := effectstest.B12Creature(g, me.ID, "Old Ooze", "Creature — Ooze", 3, 3)
	wurm := effectstest.B12Creature(g, me.ID, "Wurm", "Creature — Wurm", 4, 4)
	if err := oldB510Activate(t, g, garden, 0, "{C}{G}", game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{ooze}}); err == nil {
		t.Fatal("an Ooze was accepted as the non-Ooze creature")
	}
	if err := oldB510Activate(t, g, garden, 0, "{C}{G}", game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{wurm}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	found := false
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() && c.Name == "Ooze" && c.CurrentPower() == 4 && c.CurrentToughness() == 4 && c.HasColor("G") {
			found = true
		}
	}
	if !found {
		t.Error("no 4/4 green Ooze token")
	}
	if g.Battlefield.Contains(wurm) {
		t.Error("the Wurm stayed")
	}
}
