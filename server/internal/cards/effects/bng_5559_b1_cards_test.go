package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// bng_5559_b1_cards_test.go — Born of the Gods, batch 1 (#5559): vanilla
// and keyword bodies, bestow pumps, heroic and inspired creatures, and
// spells whose text a catalogued card already prints. Each test plays the
// real catalog entry through the engine.

const (
	bngB1CyclopsOracle       = "78922912-d7b9-47cc-8908-f44d7cc3fc4b"
	bngB1GreatHartOracle     = "694afc0a-bb32-4f24-9347-66665117ac9f"
	bngB1SwordwiseOracle     = "5c2e0762-4994-4c34-b33f-909c2b3f0f43"
	bngB1BadgerOracle        = "04be1e69-2d5c-4931-8906-a6807b5b08ca"
	bngB1NyxEidolonOracle    = "bd7fd82f-154a-4001-bb8f-7fb50ac77b6b"
	bngB1ShieldmateOracle    = "a1012f33-7ebd-47c6-b670-1b0dc046fbce"
	bngB1TritonOracle        = "cdd37bbf-09db-4c1f-9947-312de48a0f4c"
	bngB1NyxWolfOracle       = "3fa11a6a-a760-45db-8341-f83a0f69c17f"
	bngB1GhostbladeOracle    = "78a1f291-22ed-41f1-87fd-0bdc97f9c065"
	bngB1AsphyxiateOracle    = "1c2ce0d0-8c3e-48e8-b1fd-54658bdb9af9"
	bngB1ExcoriateOracle     = "a9aedd96-1aee-480d-87fe-3ed94bd64599"
	bngB1GildOracle          = "a5530fac-0bae-451d-921d-0dd4865ec91d"
	bngB1BoltOracle          = "e07f50fb-98b2-40b7-9610-b8e9438be394"
	bngB1ConflagrationOracle = "ac0a110a-0721-4aa6-a891-95e3fcc847df"
	bngB1PinnacleOracle      = "6ca999d7-48c9-4517-8e63-82ddccdb430e"
	bngB1ScouringOracle      = "f7f32fec-b384-4021-a69f-2dc9220ebc64"
	bngB1DrownOracle         = "5c93ec85-f87f-4121-9fbf-61fd598d1e65"
	bngB1RetributionOracle   = "28cfafdb-5014-44e3-9133-4554129c1aee"
	bngB1ArdorOracle         = "c3401995-4fb4-4b43-b852-5ccad6dc3154"
	bngB1ResolveOracle       = "ef29f979-1a6d-47bb-83f5-4679389471da"
	bngB1RiseOracle          = "7f3095fb-ffc0-48ed-8792-4157897a9ce7"
	bngB1MischiefOracle      = "831273d4-ced3-4659-b1f6-d5a0aeed3421"
	bngB1ReapOracle          = "dd6cec11-5c66-4f0e-a12f-9b483a09a777"
	bngB1NullifyOracle       = "5ea295af-e117-49cc-836b-da16b5d5bbdf"
	bngB1InterventionOracle  = "6f85342b-b55b-461b-a3dc-17815e02d04d"
	bngB1WeightOracle        = "b960d17b-3679-4dc5-9a5d-c8e7fd81fa9a"
	bngB1OathswornOracle     = "78d5f882-5d17-4228-b9b5-d460ab2bef08"
	bngB1ChorusOracle        = "48925016-8a55-4efb-84c7-dc5001baa58e"
	bngB1AdeptOracle         = "80eca17e-5dfd-4051-9f94-0c8aaf3747bf"
	bngB1VanguardOracle      = "4a4f6e85-df16-49a4-8fc6-149a28c2cb42"
	bngB1DiscipleOracle      = "54634ef2-cb7a-4203-8e48-2733f8c3911b"
	bngB1SunGuideOracle      = "45d939d1-e9ca-4b4b-a0a9-59b936210a5b"
	bngB1TromperOracle       = "41b2876f-0217-4552-9fad-f12f1824d06f"
	bngB1ButcherOracle       = "3e6dad31-6192-4675-81df-09689906c12b"
	bngB1HypnotistOracle     = "e30c639a-d46c-4488-bfdd-d4282046153c"
	bngB1ServantOracle       = "1ea5a8c2-d8c4-4028-b1fe-9951514ea882"
	bngB1WarchanterOracle    = "9fb5ea12-5b87-462b-b9c0-904cd0f55a9c"
	bngB1RaidersOracle       = "2fb54c18-f69b-4712-af36-14d8a6d4df2b"
	bngB1GeneralOracle       = "e861ddd5-64f7-4c59-ab96-0c69ef74febe"
	bngB1NyxSmithOracle      = "3ad9d55d-02f7-412e-8756-6a026f8e68d7"
	bngB1DriftersOracle      = "4e94f685-f392-40a8-8ad4-683351877fd9"
	bngB1DreamfinderOracle   = "ca5d5cbf-fc4f-43dd-a2af-767f277dfef9"
	bngB1StormcallerOracle   = "16a66bbe-90f6-4c29-a69d-6ff254082956"
	bngB1SunchaserOracle     = "7f240b1f-fb7b-49a3-bcaf-6348956d2a38"
	bngB1RevelerOracle       = "17d3ceeb-c972-49bc-a530-cf67c49a7e19"
	bngB1GorgonsHeadOracle   = "190cdc23-fa05-475f-b853-471ca0f21b71"
	bngB1SuddenStormOracle   = "32e915fb-3836-4bed-93b1-61c9c3951ad1"
	bngB1CullingMarkOracle   = "0cb7f468-070a-463c-b00c-16db3421d16a"
	bngB1HoldAtBayOracle     = "a57b84e4-951c-4475-bb34-b68815858f87"

	bngB1Bear = "Creature — Bear"
)

var bngB1Cards = map[string]string{
	bngB1CyclopsOracle: "Cyclops of One-Eyed Pass", bngB1GreatHartOracle: "Great Hart",
	bngB1SwordwiseOracle: "Swordwise Centaur", bngB1BadgerOracle: "Charging Badger",
	bngB1NyxEidolonOracle: "Nyxborn Eidolon", bngB1ShieldmateOracle: "Nyxborn Shieldmate",
	bngB1TritonOracle: "Nyxborn Triton", bngB1NyxWolfOracle: "Nyxborn Wolf",
	bngB1GhostbladeOracle: "Ghostblade Eidolon", bngB1AsphyxiateOracle: "Asphyxiate",
	bngB1ExcoriateOracle: "Excoriate", bngB1GildOracle: "Gild",
	bngB1BoltOracle: "Bolt of Keranos", bngB1ConflagrationOracle: "Fated Conflagration",
	bngB1PinnacleOracle: "Pinnacle of Rage", bngB1ScouringOracle: "Scouring Sands",
	bngB1DrownOracle: "Drown in Sorrow", bngB1RetributionOracle: "Fated Retribution",
	bngB1ArdorOracle: "Mortal's Ardor", bngB1ResolveOracle: "Mortal's Resolve",
	bngB1RiseOracle: "Rise to the Challenge", bngB1MischiefOracle: "Mischief and Mayhem",
	bngB1ReapOracle: "Reap What Is Sown", bngB1NullifyOracle: "Nullify",
	bngB1InterventionOracle: "Fated Intervention", bngB1WeightOracle: "Weight of the Underworld",
	bngB1OathswornOracle: "Setessan Oathsworn", bngB1ChorusOracle: "Chorus of the Tides",
	bngB1AdeptOracle: "Ashiok's Adept", bngB1VanguardOracle: "Vanguard of Brimaz",
	bngB1DiscipleOracle: "Sphinx's Disciple", bngB1SunGuideOracle: "Oreskos Sun Guide",
	bngB1TromperOracle: "Pheres-Band Tromper", bngB1ButcherOracle: "Kragma Butcher",
	bngB1HypnotistOracle: "Deepwater Hypnotist", bngB1ServantOracle: "Servant of Tymaret",
	bngB1WarchanterOracle: "Warchanter of Mogis", bngB1RaidersOracle: "Pheres-Band Raiders",
	bngB1GeneralOracle: "God-Favored General", bngB1NyxSmithOracle: "Satyr Nyx-Smith",
	bngB1DriftersOracle: "Forsaken Drifters", bngB1DreamfinderOracle: "Griffin Dreamfinder",
	bngB1StormcallerOracle: "Stormcaller of Keranos", bngB1SunchaserOracle: "Impetuous Sunchaser",
	bngB1RevelerOracle: "Reckless Reveler", bngB1GorgonsHeadOracle: "Gorgon's Head",
	bngB1SuddenStormOracle: "Sudden Storm", bngB1CullingMarkOracle: "Culling Mark",
	bngB1HoldAtBayOracle: "Hold at Bay",
}

// bngB1Library puts n filler cards on top of p's library, so a scry, a
// draw or a mill has cards to see.
func bngB1Library(p *game.Player, n int) {
	for i := 0; i < n; i++ {
		p.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery", Owner: p.ID, Controller: p.ID})
	}
}

// bngB1Settle resolves the stack, answering a scry prompt of p's if one
// opens on the way.
func bngB1Settle(t *testing.T, g *game.Game, p *game.Player) {
	t.Helper()
	passPriorityAroundTable(t, g)
	if scryChoiceFor(g, p.ID) != nil {
		answerScryKeepAll(t, g, p.ID)
		passPriorityAroundTable(t, g)
	}
}

// bngB1Inspire taps the permanent and untaps it through the untap verb
// (EventUntapCard), then lets the inspired trigger reach the stack.
func bngB1Inspire(t *testing.T, g *game.Game, p *game.Player, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() { battlefieldCardPtr(t, g, id).Tapped = true })
	if err := g.UntapAll(p.ID); err != nil {
		t.Fatalf("UntapAll: %v", err)
	}
	passPriorityAroundTable(t, g)
}

// bngB1Named lists the battlefield permanents controller controls with
// the given name.
func bngB1Named(g *game.Game, controller uuid.UUID, name string) []game.Card {
	var out []game.Card
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Controller == controller && c.Name == name {
				out = append(out, c)
			}
		}
	})
	return out
}

func TestBngB1EveryCardIsRegisteredFull(t *testing.T) {
	if len(bngB1Cards) != 49 {
		t.Fatalf("the batch lists %d cards, want 49", len(bngB1Cards))
	}
	for oracle, name := range bngB1Cards {
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

func TestBngB1VanillaBodiesDeclareNoAbilities(t *testing.T) {
	for _, oracle := range []string{bngB1CyclopsOracle, bngB1GreatHartOracle, bngB1SwordwiseOracle} {
		spec, _ := Lookup(oracle)
		if spec.OnResolve != nil || len(spec.Triggered) != 0 || len(spec.Static) != 0 ||
			len(spec.Activated) != 0 || len(spec.PrintedKeywords) != 0 {
			t.Errorf("%s is a vanilla body and declares no abilities", spec.Name)
		}
	}
	g := newCatalogGame(t)
	hart := b12Push(g, g.Seats[0].ID, "Great Hart", "Creature — Elk", bngB1GreatHartOracle, 2, 4)
	if p, tough := ptOf(t, g, hart); p != 2 || tough != 4 {
		t.Errorf("Great Hart is %d/%d, want 2/4", p, tough)
	}
}

func TestBngB1PrintedKeywords(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		keywords               []string
	}{
		{"Charging Badger", "Creature — Badger", bngB1BadgerOracle, []string{"trample"}},
		{"Ghostblade Eidolon", "Enchantment Creature — Spirit", bngB1GhostbladeOracle, []string{"double strike"}},
		{"Chorus of the Tides", "Creature — Siren", bngB1ChorusOracle, []string{"flying"}},
		{"Vanguard of Brimaz", "Creature — Cat Soldier", bngB1VanguardOracle, []string{"vigilance"}},
		{"Sphinx's Disciple", "Creature — Human Wizard", bngB1DiscipleOracle, []string{"flying"}},
		{"Satyr Nyx-Smith", "Creature — Satyr Shaman", bngB1NyxSmithOracle, []string{"haste"}},
		{"Griffin Dreamfinder", "Creature — Griffin", bngB1DreamfinderOracle, []string{"flying"}},
		{"Stormcaller of Keranos", "Creature — Human Shaman", bngB1StormcallerOracle, []string{"haste"}},
		{"Impetuous Sunchaser", "Creature — Human Soldier", bngB1SunchaserOracle, []string{"flying", "haste"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			id := b12Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, 1, 1)
			for _, kw := range tc.keywords {
				if !hasEffectiveKeyword(t, g, id, kw) {
					t.Errorf("%s has no %s", tc.name, kw)
				}
			}
		})
	}
}

// --- bestow -------------------------------------------------------------------

func TestBngB1BestowedEidolonsPumpTheirHost(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		power, tough           int
		wantP, wantT           int
	}{
		{"Nyxborn Eidolon", "Enchantment Creature — Spirit", bngB1NyxEidolonOracle, 2, 1, 4, 3},
		{"Nyxborn Shieldmate", "Enchantment Creature — Human Soldier", bngB1ShieldmateOracle, 1, 2, 3, 4},
		{"Nyxborn Triton", "Enchantment Creature — Merfolk", bngB1TritonOracle, 2, 3, 4, 5},
		{"Nyxborn Wolf", "Enchantment Creature — Wolf", bngB1NyxWolfOracle, 3, 1, 5, 3},
		{"Ghostblade Eidolon", "Enchantment Creature — Spirit", bngB1GhostbladeOracle, 1, 1, 3, 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			card := bestowCard(t, g, tc.name, tc.typeLine, tc.oracle, tc.power, tc.tough)
			bear := b12Creature(g, me.ID, "Bear", bngB1Bear, 2, 2)
			castBestowed(t, g, card, bear)
			passPriorityAroundTable(t, g)
			if s := readBestowState(g, card); !s.onBattlefield || s.attachedTo != bear {
				t.Fatalf("bestowed %s is not attached to the bear: %+v", tc.name, s)
			}
			if p, tough := ptOf(t, g, bear); p != tc.wantP || tough != tc.wantT {
				t.Errorf("the bear is %d/%d, want %d/%d", p, tough, tc.wantP, tc.wantT)
			}
			if tc.oracle == bngB1GhostbladeOracle && !hasEffectiveKeyword(t, g, bear, "double strike") {
				t.Error("Ghostblade Eidolon's host has no double strike")
			}
		})
	}
}

// --- removal ------------------------------------------------------------------

func TestBngB1AsphyxiateTakesOnlyAnUntappedCreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	tapped := b12Creature(g, opp.ID, "Tapped Bear", bngB1Bear, 2, 2)
	untapped := b12Creature(g, opp.ID, "Untapped Bear", bngB1Bear, 2, 2)
	g.WithWriteLock(func() { battlefieldCardPtr(t, g, tapped).Tapped = true })
	if err := castCatalogSpellErr(t, g, "Asphyxiate", "Sorcery", bngB1AsphyxiateOracle, cardRefs(tapped)); err == nil {
		t.Fatal("Asphyxiate targeted a tapped creature")
	}
	castCatalogSpell(t, g, "Asphyxiate", "Sorcery", bngB1AsphyxiateOracle, cardRefs(untapped))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(untapped) {
		t.Error("the untapped creature survived Asphyxiate")
	}
}

func TestBngB1ExcoriateExilesOnlyATappedCreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	tapped := b12Creature(g, opp.ID, "Tapped Bear", bngB1Bear, 2, 2)
	untapped := b12Creature(g, opp.ID, "Untapped Bear", bngB1Bear, 2, 2)
	g.WithWriteLock(func() { battlefieldCardPtr(t, g, tapped).Tapped = true })
	if err := castCatalogSpellErr(t, g, "Excoriate", "Sorcery", bngB1ExcoriateOracle, cardRefs(untapped)); err == nil {
		t.Fatal("Excoriate targeted an untapped creature")
	}
	castCatalogSpell(t, g, "Excoriate", "Sorcery", bngB1ExcoriateOracle, cardRefs(tapped))
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(tapped) {
		t.Error("the tapped creature is not in exile")
	}
}

func TestBngB1GildExilesAndMakesAGoldToken(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := b12Creature(g, opp.ID, "Bear", bngB1Bear, 2, 2)
	castCatalogSpell(t, g, "Gild", "Sorcery", bngB1GildOracle, cardRefs(bear))
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(bear) {
		t.Error("the creature is not in exile")
	}
	if got := len(bngB1Named(g, me.ID, "Gold")); got != 1 {
		t.Errorf("%d Gold tokens, want 1", got)
	}
}

// --- damage and sweepers ---------------------------------------------------------

func TestBngB1BoltOfKeranosHitsAPlayerAndScrys(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	bngB1Library(me, 3)
	life := lifeOf(g, opp.ID)
	castCatalogSpell(t, g, "Bolt of Keranos", "Sorcery", bngB1BoltOracle, playerRefs(opp.ID))
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, opp.ID); got != life-3 {
		t.Errorf("opponent life %d → %d, want -3", life, got)
	}
	if scryChoiceFor(g, me.ID) == nil {
		t.Error("Bolt of Keranos did not scry")
	}
}

func TestBngB1FatedConflagrationKillsAndScrysOnYourTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	bngB1Library(me, 3)
	big := b12Creature(g, opp.ID, "Big", bngB1Bear, 5, 5)
	castCatalogSpell(t, g, "Fated Conflagration", "Instant", bngB1ConflagrationOracle, cardRefs(big))
	passPriorityAroundTable(t, g)
	if scryChoiceFor(g, me.ID) == nil {
		t.Error("no scry on the caster's own turn")
	}
	bngB1Settle(t, g, me)
	if g.Battlefield.Contains(big) {
		t.Error("the 5/5 survived 5 damage")
	}
}

func TestBngB1PinnacleOfRageHitsTwoTargets(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	bear := b12Creature(g, opp.ID, "Bear", bngB1Bear, 3, 3)
	life := lifeOf(g, opp.ID)
	castCatalogSpell(t, g, "Pinnacle of Rage", "Sorcery", bngB1PinnacleOracle, append(playerRefs(opp.ID), cardRefs(bear)...))
	bngB1Settle(t, g, me)
	if got := lifeOf(g, opp.ID); got != life-3 {
		t.Errorf("opponent life %d → %d, want -3", life, got)
	}
	if g.Battlefield.Contains(bear) {
		t.Error("the 3/3 survived 3 damage")
	}
}

func TestBngB1ScouringSandsHitsOnlyOpponentsCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	mine := b12Creature(g, me.ID, "Mine", bngB1Bear, 1, 1)
	theirs := b12Creature(g, opp.ID, "Theirs", bngB1Bear, 1, 1)
	castCatalogSpell(t, g, "Scouring Sands", "Sorcery", bngB1ScouringOracle, nil)
	bngB1Settle(t, g, me)
	if !g.Battlefield.Contains(mine) {
		t.Error("my creature was dealt damage")
	}
	if g.Battlefield.Contains(theirs) {
		t.Error("the opponent's 1/1 survived")
	}
}

func TestBngB1DrownInSorrowShrinksEveryCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	small := b12Creature(g, me.ID, "Small", bngB1Bear, 2, 2)
	big := b12Creature(g, opp.ID, "Big", bngB1Bear, 3, 3)
	castCatalogSpell(t, g, "Drown in Sorrow", "Sorcery", bngB1DrownOracle, nil)
	bngB1Settle(t, g, me)
	if g.Battlefield.Contains(small) {
		t.Error("the 2/2 survived -2/-2")
	}
	if p, tough := ptOf(t, g, big); p != 1 || tough != 1 {
		t.Errorf("the 3/3 is %d/%d, want 1/1", p, tough)
	}
}

func TestBngB1FatedRetributionDestroysEveryCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	mine := b12Creature(g, me.ID, "Mine", bngB1Bear, 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", bngB1Bear, 2, 2)
	castCatalogSpell(t, g, "Fated Retribution", "Instant", bngB1RetributionOracle, nil)
	bngB1Settle(t, g, me)
	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(theirs) {
		t.Error("a creature survived Fated Retribution")
	}
}

func TestBngB1FatedInterventionMakesTwoCentaurEnchantments(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	castCatalogSpell(t, g, "Fated Intervention", "Instant", bngB1InterventionOracle, nil)
	bngB1Settle(t, g, me)
	toks := bngB1Named(g, me.ID, "Centaur")
	if len(toks) != 2 {
		t.Fatalf("%d Centaur tokens, want 2", len(toks))
	}
	for _, c := range toks {
		if !c.IsEnchantment() || !c.IsCreature() || c.CurrentPower() != 3 || c.CurrentToughness() != 3 {
			t.Errorf("token %s is not a 3/3 enchantment creature", c.TypeLine)
		}
	}
}

// --- pumps and counters -------------------------------------------------------------

func TestBngB1PumpAndKeywordInstants(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		power, tough int
		keyword      string
	}{
		{"Mortal's Ardor", bngB1ArdorOracle, 3, 3, "lifelink"},
		{"Mortal's Resolve", bngB1ResolveOracle, 3, 3, "indestructible"},
		{"Rise to the Challenge", bngB1RiseOracle, 4, 2, "first strike"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			bear := b12Creature(g, me.ID, "Bear", bngB1Bear, 2, 2)
			castCatalogSpell(t, g, tc.name, "Instant", tc.oracle, cardRefs(bear))
			passPriorityAroundTable(t, g)
			if p, tough := ptOf(t, g, bear); p != tc.power || tough != tc.tough {
				t.Errorf("the bear is %d/%d, want %d/%d", p, tough, tc.power, tc.tough)
			}
			if !hasEffectiveKeyword(t, g, bear, tc.keyword) {
				t.Errorf("the bear has no %s", tc.keyword)
			}
		})
	}
}

func TestBngB1MischiefAndMayhemPumpsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := b12Creature(g, me.ID, "A", bngB1Bear, 2, 2)
	b := b12Creature(g, me.ID, "B", bngB1Bear, 2, 2)
	castCatalogSpell(t, g, "Mischief and Mayhem", "Sorcery", bngB1MischiefOracle, cardRefs(a, b))
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if p, tough := ptOf(t, g, id); p != 6 || tough != 6 {
			t.Errorf("a target is %d/%d, want 6/6", p, tough)
		}
	}
}

func TestBngB1ReapWhatIsSownPutsACounterOnEachOfThree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := []uuid.UUID{
		b12Creature(g, me.ID, "A", bngB1Bear, 2, 2),
		b12Creature(g, me.ID, "B", bngB1Bear, 2, 2),
		b12Creature(g, me.ID, "C", bngB1Bear, 2, 2),
	}
	castCatalogSpell(t, g, "Reap What Is Sown", "Instant", bngB1ReapOracle, cardRefs(ids...))
	passPriorityAroundTable(t, g)
	for _, id := range ids {
		if got := countersOn(g, id, game.CounterPlusOne); got != 1 {
			t.Errorf("a target has %d +1/+1 counters, want 1", got)
		}
	}
}

func TestBngB1WeightOfTheUnderworldShrinksItsHost(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	beast := b12Creature(g, opp.ID, "Beast", bngB1Bear, 4, 4)
	castCatalogSpell(t, g, "Weight of the Underworld", "Enchantment — Aura", bngB1WeightOracle, cardRefs(beast))
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, beast); p != 1 || tough != 2 {
		t.Errorf("the enchanted creature is %d/%d, want 1/2", p, tough)
	}
}

// Nullify counters a creature spell and refuses a sorcery spell.
func TestBngB1NullifyCountersACreatureSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	hart := castCatalogSpell(t, g, "Great Hart", "Creature — Elk", bngB1GreatHartOracle, nil)
	nullify := pushCatalogHandCard(opp, "Nullify", "Instant", bngB1NullifyOracle)
	if err := g.CastSpell(opp.ID, nullify, game.CastSpellParams{Targets: cardTarget(hart)}); err != nil {
		t.Fatalf("cast Nullify at a creature spell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(hart) || !me.Graveyard.Contains(hart) {
		t.Error("the countered creature spell is not in its owner's graveyard")
	}

	sands := castCatalogSpell(t, g, "Scouring Sands", "Sorcery", bngB1ScouringOracle, nil)
	again := pushCatalogHandCard(opp, "Nullify", "Instant", bngB1NullifyOracle)
	if err := g.CastSpell(opp.ID, again, game.CastSpellParams{Targets: cardTarget(sands)}); err == nil {
		t.Error("Nullify targeted a sorcery spell")
	}
}

// --- heroic -----------------------------------------------------------------------

// bngB1Target casts Mortal's Ardor at the creature: a spell that targets
// it, which is what heroic watches.
func bngB1Target(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	castCatalogSpell(t, g, "Mortal's Ardor", "Instant", bngB1ArdorOracle, cardRefs(id))
}

func TestBngB1SetessanOathswornGetsTwoCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	oath := b12Push(g, me.ID, "Setessan Oathsworn", "Creature — Satyr Warrior", bngB1OathswornOracle, 1, 1)
	bngB1Target(t, g, oath)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, oath, game.CounterPlusOne); got != 2 {
		t.Errorf("%d +1/+1 counters, want 2", got)
	}
}

func TestBngB1ChorusOfTheTidesScrysWhenTargeted(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bngB1Library(me, 3)
	chorus := b12Push(g, me.ID, "Chorus of the Tides", "Creature — Siren", bngB1ChorusOracle, 3, 2)
	bngB1Target(t, g, chorus)
	passPriorityAroundTable(t, g)
	if scryChoiceFor(g, me.ID) == nil {
		t.Error("heroic did not scry")
	}
}

func TestBngB1AshioksAdeptMakesOpponentsDiscard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	opp.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Their Card", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
	adept := b12Push(g, me.ID, "Ashiok's Adept", "Creature — Human Wizard", bngB1AdeptOracle, 1, 3)
	bngB1Target(t, g, adept)
	passPriorityAroundTable(t, g)
	before := opp.Hand.Size()
	answerDiscard(t, g, opp.ID, opp.Hand.Cards[0].InstanceID)
	if got := opp.Hand.Size(); got != before-1 {
		t.Errorf("opponent hand %d → %d, want one discarded", before, got)
	}
}

func TestBngB1VanguardOfBrimazMakesACatSoldier(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	vanguard := b12Push(g, me.ID, "Vanguard of Brimaz", "Creature — Cat Soldier", bngB1VanguardOracle, 2, 2)
	bngB1Target(t, g, vanguard)
	passPriorityAroundTable(t, g)
	toks := bngB1Named(g, me.ID, "Cat Soldier")
	if len(toks) != 1 {
		t.Fatalf("%d Cat Soldier tokens, want 1", len(toks))
	}
	if !hasEffectiveKeyword(t, g, toks[0].InstanceID, "vigilance") {
		t.Error("the Cat Soldier has no vigilance")
	}
}

// --- inspired -----------------------------------------------------------------------

func TestBngB1SphinxsDiscipleDrawsWhenItUntaps(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	bngB1Library(me, 3)
	disciple := b12Push(g, me.ID, "Sphinx's Disciple", "Creature — Human Wizard", bngB1DiscipleOracle, 2, 2)
	before := me.Hand.Size()
	bngB1Inspire(t, g, me, disciple)
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand %d → %d, want one drawn", before, got)
	}
}

func TestBngB1OreskosSunGuideGainsTwoWhenItUntaps(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	guide := b12Push(g, me.ID, "Oreskos Sun Guide", "Creature — Cat Monk", bngB1SunGuideOracle, 2, 2)
	life := me.Life
	bngB1Inspire(t, g, me, guide)
	if me.Life != life+2 {
		t.Errorf("life %d → %d, want +2", life, me.Life)
	}
}

func TestBngB1PheresBandTromperGetsACounterWhenItUntaps(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	tromper := b12Push(g, me.ID, "Pheres-Band Tromper", "Creature — Centaur Warrior", bngB1TromperOracle, 3, 3)
	bngB1Inspire(t, g, me, tromper)
	if got := countersOn(g, tromper, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters, want 1", got)
	}
}

func TestBngB1KragmaButcherGetsPlusTwoWhenItUntaps(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	butcher := b12Push(g, me.ID, "Kragma Butcher", "Creature — Minotaur Warrior", bngB1ButcherOracle, 2, 3)
	bngB1Inspire(t, g, me, butcher)
	if p, tough := ptOf(t, g, butcher); p != 4 || tough != 3 {
		t.Errorf("Kragma Butcher is %d/%d, want 4/3", p, tough)
	}
}

func TestBngB1DeepwaterHypnotistShrinksAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	advanceToMain(t, g)
	hypnotist := b12Push(g, me.ID, "Deepwater Hypnotist", "Creature — Merfolk Wizard", bngB1HypnotistOracle, 2, 1)
	theirs := b12Creature(g, opp.ID, "Theirs", bngB1Bear, 4, 4)
	bngB1Inspire(t, g, me, hypnotist)
	b04WaitForPickOrSkip(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, theirs); p != 1 || tough != 4 {
		t.Errorf("the opponent's creature is %d/%d, want 1/4", p, tough)
	}
}

func TestBngB1WarchanterOfMogisGrantsIntimidate(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	warchanter := b12Push(g, me.ID, "Warchanter of Mogis", "Creature — Minotaur Shaman", bngB1WarchanterOracle, 3, 3)
	bear := b12Creature(g, me.ID, "Bear", bngB1Bear, 2, 2)
	bngB1Inspire(t, g, me, warchanter)
	b04WaitForPickOrSkip(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, bear, "intimidate") && !hasEffectiveKeyword(t, g, warchanter, "intimidate") {
		t.Error("no creature gained intimidate")
	}
}

func TestBngB1ServantOfTymaretDrainsEachOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	servant := b12Push(g, me.ID, "Servant of Tymaret", "Creature — Zombie", bngB1ServantOracle, 1, 3)
	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lives[p.ID] = p.Life
	}
	bngB1Inspire(t, g, me, servant)
	opponents := 0
	for _, p := range g.Seats {
		if p.ID == me.ID {
			continue
		}
		opponents++
		if p.Life != lives[p.ID]-1 {
			t.Errorf("an opponent's life %d → %d, want -1", lives[p.ID], p.Life)
		}
	}
	if me.Life != lives[me.ID]+opponents {
		t.Errorf("my life %d → %d, want +%d", lives[me.ID], me.Life, opponents)
	}
	spec, _ := Lookup(bngB1ServantOracle)
	if len(spec.Activated) != 1 || spec.Activated[0].Label != "{2}{B}: Regenerate this creature." {
		t.Error("Servant of Tymaret has no regeneration ability")
	}
}

func TestBngB1InspiredPaymentsMakeEnchantmentTokens(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		color                  string
		token                  string
		n, power, tough        int
	}{
		{"Pheres-Band Raiders", "Creature — Centaur Warrior", bngB1RaidersOracle, "G", "Centaur", 1, 3, 3},
		{"God-Favored General", "Creature — Human Soldier", bngB1GeneralOracle, "W", "Soldier", 2, 1, 1},
		{"Satyr Nyx-Smith", "Creature — Satyr Shaman", bngB1NyxSmithOracle, "R", "Elemental", 1, 3, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			advanceToMain(t, g)
			id := b12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
			bngB1Inspire(t, g, me, id)
			if !hasPayUnlessFor(g, me.ID) {
				t.Fatal("no pay prompt for the inspired trigger")
			}
			b09AddColorless(me, 2)
			me.ManaPool.AddMana(game.ManaToken{Color: tc.color})
			answerPayUnless(t, g, me.ID, true)
			passPriorityAroundTable(t, g)
			toks := bngB1Named(g, me.ID, tc.token)
			if len(toks) != tc.n {
				t.Fatalf("%d %s tokens, want %d", len(toks), tc.token, tc.n)
			}
			for _, c := range toks {
				if !c.IsEnchantment() || c.CurrentPower() != tc.power || c.CurrentToughness() != tc.tough {
					t.Errorf("token %s %d/%d, want a %d/%d enchantment creature", c.TypeLine, c.CurrentPower(), c.CurrentToughness(), tc.power, tc.tough)
				}
			}
		})
	}
}

// --- other creatures -----------------------------------------------------------------

func TestBngB1ForsakenDriftersMillsFourWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	bngB1Library(me, 6)
	drifters := b12Push(g, me.ID, "Forsaken Drifters", "Creature — Zombie", bngB1DriftersOracle, 4, 2)
	source := b12Creature(g, opp.ID, "Source", bngB1Bear, 2, 2)
	library := me.Library.Size()
	b27Damage(g, source, drifters, 2)
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(drifters) {
		t.Fatal("Forsaken Drifters did not die")
	}
	if got := me.Library.Size(); got != library-4 {
		t.Errorf("library %d → %d, want four milled", library, got)
	}
}

func TestBngB1GriffinDreamfinderReturnsAnEnchantmentCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	aura := pushGraveyardCardTyped(me, "Old Aura", "Enchantment — Aura")
	castHoldCreature(t, g, "Griffin Dreamfinder", "Creature — Griffin", bngB1DreamfinderOracle, 1, 4)
	b04WaitForPickOrSkip(t, g, me.ID, aura)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(aura) {
		t.Error("the enchantment card did not return to hand")
	}
}

func TestBngB1StormcallerOfKeranosScrysForMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	bngB1Library(me, 3)
	caller := b12Push(g, me.ID, "Stormcaller of Keranos", "Creature — Human Shaman", bngB1StormcallerOracle, 2, 2)
	b09AddColorless(me, 1)
	me.ManaPool.AddMana(game.ManaToken{Color: "U"})
	if err := g.ActivateCatalogAbility(me.ID, caller, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if scryChoiceFor(g, me.ID) == nil {
		t.Error("the ability did not scry")
	}
}

func TestBngB1ImpetuousSunchaserMustAttack(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := b12Push(g, me.ID, "Impetuous Sunchaser", "Creature — Human Soldier", bngB1SunchaserOracle, 1, 1)
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	c := findBattlefieldCardByID(g, id)
	if c == nil {
		t.Fatal("Impetuous Sunchaser is not on the battlefield")
	}
	if len(c.Effective().AttackRequirements) == 0 {
		t.Error("Impetuous Sunchaser has no attack requirement")
	}
}

func TestBngB1RecklessRevelerDestroysAnArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	advanceToMain(t, g)
	reveler := b12Push(g, me.ID, "Reckless Reveler", "Creature — Satyr", bngB1RevelerOracle, 2, 1)
	relic := b12Permanent(g, opp.ID, "Relic", "Artifact")
	me.ManaPool.AddMana(game.ManaToken{Color: "R"})
	if err := g.ActivateCatalogAbility(me.ID, reveler, 0, game.ActivateAbilityParams{Targets: cardRefs(relic)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(reveler) {
		t.Error("Reckless Reveler was not sacrificed")
	}
	if g.Battlefield.Contains(relic) {
		t.Error("the artifact survived")
	}
}

func TestBngB1GorgonsHeadGrantsDeathtouch(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	head := b12Push(g, me.ID, "Gorgon's Head", "Artifact — Equipment", bngB1GorgonsHeadOracle, 0, 0)
	bear := b12Creature(g, me.ID, "Bear", bngB1Bear, 2, 2)
	b09AddColorless(me, 2)
	if err := g.ActivateCatalogAbility(me.ID, head, 0, game.ActivateAbilityParams{Targets: cardRefs(bear)}); err != nil {
		t.Fatalf("equip: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, bear, "deathtouch") {
		t.Error("the equipped creature has no deathtouch")
	}
}

// --- other spells --------------------------------------------------------------------

func TestBngB1SuddenStormTapsTwoCreatures(t *testing.T) {
	g := newCatalogGame(t)
	_, opp, _ := blbB1Seats(g)
	a := b12Creature(g, opp.ID, "A", bngB1Bear, 2, 2)
	b := b12Creature(g, opp.ID, "B", bngB1Bear, 2, 2)
	castCatalogSpell(t, g, "Sudden Storm", "Instant", bngB1SuddenStormOracle, cardRefs(a, b))
	passPriorityAroundTable(t, g)
	if !b27Tapped(t, g, a) || !b27Tapped(t, g, b) {
		t.Error("both targets should be tapped")
	}
}

func TestBngB1CullingMarkMakesItBlock(t *testing.T) {
	g := newCatalogGame(t)
	_, opp, _ := blbB1Seats(g)
	bear := b12Creature(g, opp.ID, "Bear", bngB1Bear, 2, 2)
	castCatalogSpell(t, g, "Culling Mark", "Sorcery", bngB1CullingMarkOracle, cardRefs(bear))
	passPriorityAroundTable(t, g)
	if !mkmB1MustBlock(t, g, bear) {
		t.Error("the target has no blocks-if-able requirement")
	}
}

func TestBngB1HoldAtBayPreventsSevenDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	source := b12Creature(g, opp.ID, "Source", bngB1Bear, 2, 2)
	castCatalogSpell(t, g, "Hold at Bay", "Instant", bngB1HoldAtBayOracle, playerRefs(me.ID))
	passPriorityAroundTable(t, g)
	life := me.Life
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(source, me.ID, 7) })
	if me.Life != life {
		t.Errorf("life %d → %d, want the 7 prevented", life, me.Life)
	}
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(source, me.ID, 1) })
	if me.Life != life-1 {
		t.Errorf("life %d → %d, want the eighth point dealt", life, me.Life)
	}
}
