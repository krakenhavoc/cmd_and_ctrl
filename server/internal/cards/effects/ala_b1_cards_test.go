package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ala_b1_cards_test.go — Shards of Alara, batch 1 (#5588). Each test plays
// the real catalog entry through the engine and asserts what its text says.

const (
	alaB1ObeliskGrixisOracle   = "e193ae18-91a0-4870-b8a1-f8dedf744708"
	alaB1ObeliskEsperOracle    = "9da963f6-734b-4d3c-a3be-dd5cca2c7b19"
	alaB1ObeliskNayaOracle     = "02e81bb8-ba80-4483-9b9f-973dcae0fc9d"
	alaB1ObeliskJundOracle     = "eb0be178-55f9-4743-8128-34bc86f134ec"
	alaB1ObeliskBantOracle     = "3f96dfd3-975a-4595-95c1-d2f04a86cb37"
	alaB1SquireOracle          = "2b8ef83b-1a1d-4299-ab69-e9e296bb224d"
	alaB1PaladinOracle         = "b68e99df-7899-493e-9453-59da14ff7de4"
	alaB1GuardiansOracle       = "7ef8df9f-3fab-44fd-b5a9-f57e54dbc38e"
	alaB1CourtArchersOracle    = "cc1cb63e-5811-44ab-af50-37a96f79fc78"
	alaB1RhoxChargerOracle     = "e1b7db10-58a1-46fc-ac6c-287393ed96c2"
	alaB1WaveskimmerOracle     = "113d6c3c-3492-47aa-81d1-359d852d8d0a"
	alaB1OutriderOracle        = "ccfe72b2-8a9a-4ad6-ab1d-676bb376b0ac"
	alaB1StrixOracle           = "19a39cc5-2134-4dee-932c-81389806c57b"
	alaB1FatestitcherOracle    = "55765934-07be-4f03-a2e8-8292c0b601ae"
	alaB1TitanicOracle         = "0cd28580-d718-4d0b-96e3-a836c2911c27"
	alaB1LethemancerOracle     = "d696aba9-ef3e-4f81-821a-c1df51431b8a"
	alaB1DrumhunterOracle      = "3aab84f2-731c-4079-9623-0b0c433442c3"
	alaB1AngelsongOracle       = "fe7bad80-f853-4af9-82e2-5ccf8038d93b"
	alaB1RangerOracle          = "b9c8b929-f3ee-457f-9cbc-9de18eb69dc5"
	alaB1EsperCharmOracle      = "d40ea3d9-8c48-4846-b1f4-99018d4fd789"
	alaB1ConnoisseurOracle     = "f39e2611-c837-4ad7-b3a8-1ee48b74777c"
	alaB1KnightCaptainOracle   = "611f1714-a8bb-4ce4-a810-4d26fe64e358"
	alaB1IguanarOracle         = "7c4fc524-4734-4fb8-a5ca-ac2617242c3f"
	alaB1NecrogenesisOracle    = "ad08952a-bcae-41d6-8093-c18535dad7d5"
	alaB1ThrinaxOracle         = "ea74e89e-0724-4744-a6c9-e35c415624fb"
	alaB1DispellersOracle      = "91fc862d-fdde-45ea-a05d-30d38e7a735c"
	alaB1CouriersOracle        = "4e694613-e2e1-4ae3-baa8-f372259e91cf"
	alaB1RafiqOracle           = "c6e17443-2379-419d-8c32-13ea9bf52993"
	alaB1BattlegraceOracle     = "40216d09-ae4d-4a41-b3f6-1a8035093045"
	alaB1BenedictionOracle     = "6b19377c-f178-49cb-95f8-fd3451acc7cf"
	alaB1JundCharmOracle       = "7bd2ac13-d598-4cf1-a483-2a6878aed93b"
	alaB1GrixisCharmOracle     = "387ae7be-1cb5-4ae8-bb63-edf4941945a9"
	alaB1BroodmateOracle       = "86529f12-6f0f-4b0d-96f1-785474f6763a"
	alaB1BlightningOracle      = "a6496440-dc0c-4d9b-bf37-f537b6f0187b"
	alaB1StingerOracle         = "1c329131-bef2-41c6-823a-136c110ff0ad"
	alaB1SphinxSovereignOracle = "41907508-9ffc-4e01-969a-4f9c35cdf548"
	alaB1HellsThunderOracle    = "5aad3485-47c1-4d87-b340-61d7003cd786"
	alaB1AncientsTreadOracle   = "fca2fcab-4f17-448d-bf6d-f6c913159df8"
	alaB1CallToHeelOracle      = "8ef3dd68-f6a1-43b9-8837-40b326019b1b"
	alaB1ViolentOracle         = "a04cfcc5-7bb4-4352-ad8d-9d625585e656"
	alaB1GargoyleOracle        = "6ee9802a-7a44-49c7-93d9-f75c4d42e71e"
	alaB1StoicAngelOracle      = "b50c865b-705f-489e-9ffc-ba8060ba262f"
)

func TestAlaB1EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		alaB1ObeliskGrixisOracle:   "Obelisk of Grixis",
		alaB1ObeliskEsperOracle:    "Obelisk of Esper",
		alaB1ObeliskNayaOracle:     "Obelisk of Naya",
		alaB1ObeliskJundOracle:     "Obelisk of Jund",
		alaB1ObeliskBantOracle:     "Obelisk of Bant",
		alaB1SquireOracle:          "Akrasan Squire",
		alaB1PaladinOracle:         "Sigiled Paladin",
		alaB1GuardiansOracle:       "Guardians of Akrasa",
		alaB1CourtArchersOracle:    "Court Archers",
		alaB1RhoxChargerOracle:     "Rhox Charger",
		alaB1WaveskimmerOracle:     "Waveskimmer Aven",
		alaB1OutriderOracle:        "Outrider of Jhess",
		alaB1StrixOracle:           "Tidehollow Strix",
		alaB1FatestitcherOracle:    "Fatestitcher",
		alaB1TitanicOracle:         "Titanic Ultimatum",
		alaB1LethemancerOracle:     "Cunning Lethemancer",
		alaB1DrumhunterOracle:      "Drumhunter",
		alaB1AngelsongOracle:       "Angelsong",
		alaB1RangerOracle:          "Ranger of Eos",
		alaB1EsperCharmOracle:      "Esper Charm",
		alaB1ConnoisseurOracle:     "Corpse Connoisseur",
		alaB1KnightCaptainOracle:   "Knight-Captain of Eos",
		alaB1IguanarOracle:         "Hissing Iguanar",
		alaB1NecrogenesisOracle:    "Necrogenesis",
		alaB1ThrinaxOracle:         "Sprouting Thrinax",
		alaB1DispellersOracle:      "Dispeller's Capsule",
		alaB1CouriersOracle:        "Courier's Capsule",
		alaB1RafiqOracle:           "Rafiq of the Many",
		alaB1BattlegraceOracle:     "Battlegrace Angel",
		alaB1BenedictionOracle:     "Angelic Benediction",
		alaB1JundCharmOracle:       "Jund Charm",
		alaB1GrixisCharmOracle:     "Grixis Charm",
		alaB1BroodmateOracle:       "Broodmate Dragon",
		alaB1BlightningOracle:      "Blightning",
		alaB1StingerOracle:         "Vithian Stinger",
		alaB1SphinxSovereignOracle: "Sphinx Sovereign",
		alaB1HellsThunderOracle:    "Hell's Thunder",
		alaB1AncientsTreadOracle:   "Where Ancients Tread",
		alaB1CallToHeelOracle:      "Call to Heel",
		alaB1ViolentOracle:         "Violent Ultimatum",
		alaB1GargoyleOracle:        "Sanctum Gargoyle",
		alaB1StoicAngelOracle:      "Stoic Angel",
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

// alaB1Settle passes priority until the stack is empty, putting a batch of
// simultaneous triggers on the stack in the order offered. It stops at any
// other prompt, which the test answers.
func alaB1Settle(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 48; i++ {
		var order *game.PendingChoice
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoiceTriggerOrder {
				order = c
			}
		}
		if order != nil {
			if err := g.ResolveTriggerOrder(order.ID, order.Chooser, append([]uuid.UUID(nil), order.TriggerOrderIDs...)); err != nil {
				t.Fatalf("ResolveTriggerOrder: %v", err)
			}
			continue
		}
		if len(g.PendingChoices) > 0 || stackFullyEmpty(g) {
			return
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
}

// alaB1CastCreature casts a vanilla creature card of the given size from
// the active player's hand and resolves it, stopping at any prompt its
// arrival raises.
func alaB1CastCreature(t *testing.T, g *game.Game, name string, power, toughness int) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: "Creature — Beast",
		Power: power, Toughness: toughness, Owner: active.ID, Controller: active.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	passPriorityAroundTable(t, g)
	return id
}

// alaB1Attack declares attackers against defender and settles the attack
// triggers.
func alaB1Attack(t *testing.T, g *game.Game, defender uuid.UUID, attackers ...uuid.UUID) {
	t.Helper()
	advanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range attackers {
		if err := g.DeclareAttacker(id, defender); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	lockInAttacks(t, g)
	alaB1Settle(t, g)
}

// --- printed keywords ----------------------------------------------------------------

func TestAlaB1CardsHaveTheirPrintedKeywords(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		keywords               []string
	}{
		{"Akrasan Squire", "Creature — Human Soldier", alaB1SquireOracle, []string{"exalted"}},
		{"Sigiled Paladin", "Creature — Human Knight", alaB1PaladinOracle, []string{"first strike", "exalted"}},
		{"Guardians of Akrasa", "Creature — Human Soldier", alaB1GuardiansOracle, []string{"defender", "exalted"}},
		{"Court Archers", "Creature — Human Archer", alaB1CourtArchersOracle, []string{"reach", "exalted"}},
		{"Rhox Charger", "Creature — Rhino Soldier", alaB1RhoxChargerOracle, []string{"trample", "exalted"}},
		{"Waveskimmer Aven", "Creature — Bird Soldier", alaB1WaveskimmerOracle, []string{"flying", "exalted"}},
		{"Outrider of Jhess", "Creature — Human Knight", alaB1OutriderOracle, []string{"exalted"}},
		{"Tidehollow Strix", "Artifact Creature — Bird", alaB1StrixOracle, []string{"flying", "deathtouch"}},
		{"Rafiq of the Many", "Legendary Creature — Human Knight", alaB1RafiqOracle, []string{"exalted"}},
		{"Battlegrace Angel", "Creature — Angel", alaB1BattlegraceOracle, []string{"flying", "exalted"}},
		{"Broodmate Dragon", "Creature — Dragon", alaB1BroodmateOracle, []string{"flying"}},
		{"Sphinx Sovereign", "Artifact Creature — Sphinx", alaB1SphinxSovereignOracle, []string{"flying"}},
		{"Hell's Thunder", "Creature — Elemental", alaB1HellsThunderOracle, []string{"flying", "haste"}},
		{"Sanctum Gargoyle", "Artifact Creature — Gargoyle", alaB1GargoyleOracle, []string{"flying"}},
		{"Stoic Angel", "Creature — Angel", alaB1StoicAngelOracle, []string{"flying", "vigilance"}},
	} {
		g := newCatalogGame(t)
		id := b12Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
		for _, kw := range tc.keywords {
			if !hasKeywordOnBattlefield(t, g, id, kw) {
				t.Errorf("%s lacks %s", tc.name, kw)
			}
		}
	}
}

// --- exalted ---------------------------------------------------------------------------

func TestAlaB1ExaltedPumpsALoneAttacker(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle string }{
		{"Akrasan Squire", "Creature — Human Soldier", alaB1SquireOracle},
		{"Sigiled Paladin", "Creature — Human Knight", alaB1PaladinOracle},
		{"Guardians of Akrasa", "Creature — Human Soldier", alaB1GuardiansOracle},
		{"Court Archers", "Creature — Human Archer", alaB1CourtArchersOracle},
		{"Rhox Charger", "Creature — Rhino Soldier", alaB1RhoxChargerOracle},
		{"Waveskimmer Aven", "Creature — Bird Soldier", alaB1WaveskimmerOracle},
		{"Outrider of Jhess", "Creature — Human Knight", alaB1OutriderOracle},
	} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		advanceToMain(t, g)
		b12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 0, 4)
		bear := seedBear(g, me.ID)
		alaB1Attack(t, g, opp.ID, bear)
		if got := effectivePower(t, g, bear); got != 3 {
			t.Errorf("%s: the lone attacker has power %d, want 3", tc.name, got)
		}
	}

	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Akrasan Squire", "Creature — Human Soldier", alaB1SquireOracle, 1, 1)
	a, b := seedBear(g, me.ID), seedBear(g, me.ID)
	alaB1Attack(t, g, opp.ID, a, b)
	if got := effectivePower(t, g, a); got != 2 {
		t.Errorf("one of two attackers has power %d, want 2: neither attacks alone", got)
	}
}

// --- Rafiq of the Many, Battlegrace Angel --------------------------------------------------

func TestAlaB1LoneAttackerGainsAKeyword(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle, keyword string }{
		{"Rafiq of the Many", "Legendary Creature — Human Knight", alaB1RafiqOracle, "double strike"},
		{"Battlegrace Angel", "Creature — Angel", alaB1BattlegraceOracle, "lifelink"},
	} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		advanceToMain(t, g)
		b12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 3, 3)
		bear := seedBear(g, me.ID)
		alaB1Attack(t, g, opp.ID, bear)
		if !hasEffectiveKeyword(t, g, bear, tc.keyword) {
			t.Errorf("%s: the lone attacker lacks %s", tc.name, tc.keyword)
		}
		if got := effectivePower(t, g, bear); got != 3 {
			t.Errorf("%s: the lone attacker has power %d, want 3 (exalted)", tc.name, got)
		}

		g = newCatalogGame(t)
		me, opp = g.Seats[0], g.Seats[1]
		advanceToMain(t, g)
		b12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 3, 3)
		a, b := seedBear(g, me.ID), seedBear(g, me.ID)
		alaB1Attack(t, g, opp.ID, a, b)
		if hasEffectiveKeyword(t, g, a, tc.keyword) {
			t.Errorf("%s: one of two attackers gained %s", tc.name, tc.keyword)
		}
	}
}

// --- Angelic Benediction ---------------------------------------------------------------

func TestAlaB1AngelicBenedictionTapsATargetWhenOneAttacksAlone(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Angelic Benediction", "Enchantment", alaB1BenedictionOracle, 0, 0)
	bear := seedBear(g, me.ID)
	victim := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	alaB1Attack(t, g, opp.ID, bear)
	pickCard(t, g, me.ID, victim)
	alaB1Settle(t, g)
	if latestTriggerPrompt(g, me.ID) == nil {
		var kinds []game.PendingChoiceKind
		for _, c := range g.PendingChoices {
			kinds = append(kinds, c.Kind)
		}
		t.Fatalf("no \"you may\" prompt: pending %v, stack %d, step %s", kinds, len(g.StackMeta), g.Turn.Step)
	}
	answerLatestTriggerPrompt(t, g, me.ID, true)
	alaB1Settle(t, g)
	if !isTapped(g, victim) {
		t.Error("the target was not tapped")
	}
	if got := effectivePower(t, g, bear); got != 3 {
		t.Errorf("the lone attacker has power %d, want 3 (exalted)", got)
	}
}

// --- Obelisks ---------------------------------------------------------------------------

func TestAlaB1ObelisksTapForTheirShardsColours(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		colours      []string
		not          string
	}{
		{"Obelisk of Grixis", alaB1ObeliskGrixisOracle, []string{"U", "B", "R"}, "G"},
		{"Obelisk of Esper", alaB1ObeliskEsperOracle, []string{"W", "U", "B"}, "R"},
		{"Obelisk of Naya", alaB1ObeliskNayaOracle, []string{"R", "G", "W"}, "U"},
		{"Obelisk of Jund", alaB1ObeliskJundOracle, []string{"B", "R", "G"}, "W"},
		{"Obelisk of Bant", alaB1ObeliskBantOracle, []string{"G", "W", "U"}, "B"},
	} {
		g := newCatalogGame(t)
		advanceToMain(t, g)
		me := g.Seats[g.Turn.ActiveSeat]
		g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
		refused := b12Push(g, me.ID, tc.name, "Artifact", tc.oracle, 0, 0)
		if err := g.ActivateManaAbility(me.ID, refused, 0, game.ManaAbilityParams{Colors: []string{tc.not}}); err == nil {
			t.Errorf("%s made {%s}", tc.name, tc.not)
		}
		for _, colour := range tc.colours {
			rock := b12Push(g, me.ID, tc.name, "Artifact", tc.oracle, 0, 0)
			if err := g.ActivateManaAbility(me.ID, rock, 0, game.ManaAbilityParams{Colors: []string{colour}}); err != nil {
				t.Fatalf("%s: add {%s}: %v", tc.name, colour, err)
			}
			if got := me.ManaPool.CountColor(colour); got != 1 {
				t.Errorf("%s: pool holds %d {%s}, want 1", tc.name, got, colour)
			}
			if !isTapped(g, rock) {
				t.Errorf("%s did not tap", tc.name)
			}
		}
	}
}

// --- Fatestitcher -------------------------------------------------------------------------

func TestAlaB1FatestitcherTapsOrUntapsAnotherPermanent(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, them := seatsForCombat(g)
	stitcher := b12Push(g, me.ID, "Fatestitcher", "Creature — Zombie Wizard", alaB1FatestitcherOracle, 1, 2)
	if err := ravB2Activate(t, g, me.ID, stitcher, 0, game.ActivateAbilityParams{Targets: cardTarget(stitcher)}); err == nil {
		t.Fatal("Fatestitcher targeted itself")
	}
	land := seedLand(g, them.ID, "Forest", "Basic Land — Forest", "")
	if err := ravB2Activate(t, g, me.ID, stitcher, 0, game.ActivateAbilityParams{Targets: cardTarget(land)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerOptionPick(t, g, me.ID, 0)
	if !b16Tapped(t, g, land) {
		t.Fatal("the land was not tapped")
	}
	findBattlefieldCardForTest(g, stitcher).Tapped = false
	ravB2MustActivate(t, g, me.ID, stitcher, 0, game.ActivateAbilityParams{Targets: cardTarget(land)})
	answerOptionPick(t, g, me.ID, 2)
	if !b16Tapped(t, g, land) {
		t.Error("leaving it as it is untapped the land")
	}
	findBattlefieldCardForTest(g, stitcher).Tapped = false
	ravB2MustActivate(t, g, me.ID, stitcher, 0, game.ActivateAbilityParams{Targets: cardTarget(land)})
	answerOptionPick(t, g, me.ID, 1)
	if b16Tapped(t, g, land) {
		t.Error("the land was not untapped")
	}
}

// --- unearth --------------------------------------------------------------------------

var alaB1Unearthers = []broB04Card{
	{"Fatestitcher", "Creature — Zombie Wizard", alaB1FatestitcherOracle, 1, 2, "{U}"},
	{"Corpse Connoisseur", "Creature — Zombie Wizard", alaB1ConnoisseurOracle, 3, 3, "{3}{B}"},
	{"Vithian Stinger", "Creature — Human Shaman", alaB1StingerOracle, 0, 1, "{1}{R}"},
	{"Hell's Thunder", "Creature — Elemental", alaB1HellsThunderOracle, 4, 4, "{4}{R}"},
}

func TestAlaB1UnearthAbilities(t *testing.T) {
	for _, tc := range alaB1Unearthers {
		var found *game.ActivatedAbilityShape
		for _, ab := range game.ActivatedAbilitiesForCard(game.Card{OracleID: tc.oracle}) {
			if game.AbilityFunctionsFromZone(ab, game.ZoneGraveyard) {
				ab := ab
				found = &ab
			}
		}
		if found == nil {
			t.Errorf("%s: no graveyard ability", tc.name)
			continue
		}
		if found.Cost.Mana != tc.unearth || !found.SorcerySpeed {
			t.Errorf("%s: unearth cost %q sorcery %v, want %q sorcery", tc.name, found.Cost.Mana, found.SorcerySpeed, tc.unearth)
		}
		if game.AbilityFunctionsFromZone(*found, game.ZoneBattlefield) {
			t.Errorf("%s: unearth is offered on the battlefield", tc.name)
		}
	}
}

func TestAlaB1VithianStingerUnearthsAndPings(t *testing.T) {
	g := newCatalogGame(t)
	id, me := broB04Unearth(t, g, alaB1Unearthers[2])
	if !g.Battlefield.Contains(id) {
		t.Fatal("Vithian Stinger did not return")
	}
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	life := opp.Life
	ravB2MustActivate(t, g, me.ID, id, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{playerTarget(opp.ID)}})
	if opp.Life != life-1 {
		t.Errorf("life %d → %d, want 1 damage from an unearthed (hasty) Stinger", life, opp.Life)
	}
}

// --- Titanic Ultimatum -------------------------------------------------------------------

func TestAlaB1TitanicUltimatumPumpsYourCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seatsForCombat(g)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Titanic Ultimatum", "Sorcery", alaB1TitanicOracle, nil)
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, mine); p != 7 || tough != 7 {
		t.Errorf("your creature is %d/%d, want 7/7", p, tough)
	}
	for _, kw := range []string{"first strike", "trample", "lifelink"} {
		if !hasEffectiveKeyword(t, g, mine, kw) {
			t.Errorf("your creature lacks %s", kw)
		}
		if hasEffectiveKeyword(t, g, theirs, kw) {
			t.Errorf("an opponent's creature gained %s", kw)
		}
	}
	if p, tough := effectivePTOf(t, g, theirs); p != 2 || tough != 2 {
		t.Errorf("an opponent's creature is %d/%d, want 2/2", p, tough)
	}
	dcCleanup(g)
	if p, _ := effectivePTOf(t, g, mine); p != 2 {
		t.Errorf("the pump outlived the cleanup step (power %d)", p)
	}
}

// --- Cunning Lethemancer -------------------------------------------------------------------

func TestAlaB1CunningLethemancerMakesEachPlayerDiscard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	b12Push(g, me.ID, "Cunning Lethemancer", "Creature — Human Wizard", alaB1LethemancerOracle, 2, 2)
	for _, p := range g.Seats {
		d2dSetHand(p, 2)
	}
	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		if discardChoiceFor(g, p.ID) == nil {
			t.Errorf("seat %s was not asked to discard", p.ID)
			continue
		}
		answerDiscard(t, g, p.ID, p.Hand.Cards[0].InstanceID)
		if p.Hand.Size() != 1 {
			t.Errorf("hand %d after the discard, want 1", p.Hand.Size())
		}
	}
}

// --- Drumhunter ----------------------------------------------------------------------------

func TestAlaB1DrumhunterDrawsOnlyWithAPowerFiveCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Drumhunter", "Creature — Human Druid Warrior", alaB1DrumhunterOracle, 2, 2)
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if latestTriggerPrompt(g, me.ID) != nil {
		t.Fatal("Drumhunter asked with no creature of power 5 or greater")
	}

	g = newCatalogGame(t)
	me = g.Seats[0]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Drumhunter", "Creature — Human Druid Warrior", alaB1DrumhunterOracle, 2, 2)
	b12Creature(g, me.ID, "Big Beast", "Creature — Beast", 5, 5)
	hand := me.Hand.Size()
	advanceToEndStepOf(t, g, 0)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want one draw", hand, me.Hand.Size())
	}
}

func TestAlaB1DrumhunterTapsForColourless(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	id := b12Push(g, me.ID, "Drumhunter", "Creature — Human Druid Warrior", alaB1DrumhunterOracle, 2, 2)
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("add {C}: %v", err)
	}
	if got := me.ManaPool.CountColor("C"); got != 1 {
		t.Errorf("pool holds %d {C}, want 1", got)
	}
}

// --- Angelsong ------------------------------------------------------------------------------

func TestAlaB1AngelsongPreventsCombatDamageAndCycles(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	castCatalogSpell(t, g, "Angelsong", "Instant", alaB1AngelsongOracle, nil)
	passPriorityAroundTable(t, g)
	life := opp.Life
	attackWith(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	if opp.Life != life {
		t.Errorf("life %d → %d, want the combat damage prevented", life, opp.Life)
	}
	usgB4Cycles(t, "Angelsong", "Instant", alaB1AngelsongOracle)
}

// --- Ranger of Eos ----------------------------------------------------------------------------

func TestAlaB1RangerOfEosFindsTwoSmallCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushLibraryCardForTest(me, game.Card{Name: "Elf One", TypeLine: "Creature — Elf", ManaCost: "{G}"})
	pushLibraryCardForTest(me, game.Card{Name: "Elf Two", TypeLine: "Creature — Elf", ManaCost: "{G}"})
	pushLibraryCardForTest(me, game.Card{Name: "Big Beast", TypeLine: "Creature — Beast", ManaCost: "{4}{G}"})
	castAndResolveCreature(t, g, "Ranger of Eos", "Creature — Human Soldier Ranger", alaB1RangerOracle)
	passPriorityAroundTable(t, g)
	answerSearchNamed(t, g, me.ID, "Elf One", "Elf Two")
	passPriorityAroundTable(t, g)
	got := map[string]bool{}
	for _, c := range me.Hand.Cards {
		got[c.Name] = true
	}
	if !got["Elf One"] || !got["Elf Two"] {
		t.Errorf("the two one-drops did not reach the hand: %v", got)
	}
	if got["Big Beast"] {
		t.Error("a five-mana creature was found")
	}
}

// --- Esper Charm -----------------------------------------------------------------------------

func TestAlaB1EsperCharmModes(t *testing.T) {
	g := newCatalogGame(t)
	_, opp := seatsForCombat(g)
	ench := b12Permanent(g, opp.ID, "Their Enchantment", "Enchantment")
	castCatalogSpellWithModesTargets(t, g, alaB1EsperCharmOracle, []int{0}, cardTarget(ench))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ench) {
		t.Error("the enchantment survived")
	}

	g = newCatalogGame(t)
	me, _ := seatsForCombat(g)
	hand := me.Hand.Size()
	castCatalogSpellWithModes(t, g, "Esper Charm", "Instant", alaB1EsperCharmOracle, []int{1})
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 {
		t.Errorf("hand %d → %d, want two draws", hand, me.Hand.Size())
	}

	g = newCatalogGame(t)
	_, opp = seatsForCombat(g)
	d2dSetHand(opp, 3)
	castCatalogSpellWithModesTargets(t, g, alaB1EsperCharmOracle, []int{2}, []game.TargetRef{playerTarget(opp.ID)})
	passPriorityAroundTable(t, g)
	c := discardChoiceFor(g, opp.ID)
	if c == nil {
		t.Fatal("the target player was not asked to discard")
	}
	answerDiscard(t, g, opp.ID, opp.Hand.Cards[0].InstanceID, opp.Hand.Cards[1].InstanceID)
	if opp.Hand.Size() != 1 {
		t.Errorf("the target player's hand is %d, want 1", opp.Hand.Size())
	}
}

// --- Corpse Connoisseur -------------------------------------------------------------------

func TestAlaB1CorpseConnoisseurPutsACreatureIntoYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushLibraryCardForTest(me, game.Card{Name: "Big Body", TypeLine: "Creature — Giant", ManaCost: "{7}"})
	pushLibraryCardForTest(me, game.Card{Name: "A Spell", TypeLine: "Sorcery", ManaCost: "{1}"})
	castAndResolveCreature(t, g, "Corpse Connoisseur", "Creature — Zombie Wizard", alaB1ConnoisseurOracle)
	passPriorityAroundTable(t, g)
	answerSearchNamed(t, g, me.ID, "Big Body")
	passPriorityAroundTable(t, g)
	found := false
	for _, c := range me.Graveyard.Cards {
		if c.Name == "Big Body" {
			found = true
		}
	}
	if !found {
		t.Error("the creature card did not reach the graveyard")
	}
}

// --- Knight-Captain of Eos -----------------------------------------------------------------

func TestAlaB1KnightCaptainMakesSoldiersAndSacrificesOneToFog(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	captain := castAndResolveCreature(t, g, "Knight-Captain of Eos", "Creature — Human Knight", alaB1KnightCaptainOracle)
	passPriorityAroundTable(t, g)
	if n := tokensNamedControlledBy(g, "Soldier", me.ID); n != 2 {
		t.Fatalf("%d Soldier tokens, want 2", n)
	}
	var soldier uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Soldier" && c.Controller == me.ID {
			soldier = c.InstanceID
		}
	}
	bear := seedBear(g, me.ID)
	if err := ravB2Activate(t, g, me.ID, captain, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{bear}}, "W"); err == nil {
		t.Error("a non-Soldier paid the sacrifice")
	}
	ravB2MustActivate(t, g, me.ID, captain, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{soldier}}, "W")
	if g.Battlefield.Contains(soldier) {
		t.Error("the Soldier was not sacrificed")
	}
	if n := tokensNamedControlledBy(g, "Soldier", me.ID); n != 1 {
		t.Errorf("%d Soldier tokens after the sacrifice, want 1", n)
	}
}

// --- Hissing Iguanar ----------------------------------------------------------------------

func TestAlaB1HissingIguanarPingsWhenAnotherCreatureDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seatsForCombat(g)
	b12Push(g, me.ID, "Hissing Iguanar", "Creature — Lizard", alaB1IguanarOracle, 3, 1)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	life := opp.Life
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(theirs); err != nil {
			t.Fatal(err)
		}
	})
	b04WaitForPick(t, g, me.ID)
	pickPlayer(t, g, me.ID, opp.ID)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("life %d → %d, want 1 damage", life, opp.Life)
	}
}

// --- Necrogenesis ------------------------------------------------------------------------

func TestAlaB1NecrogenesisExilesACreatureCardForASaproling(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := seatsForCombat(g)
	ench := b12Push(g, me.ID, "Necrogenesis", "Enchantment", alaB1NecrogenesisOracle, 0, 0)
	spell := pushGraveyardCardTyped(opp, "A Sorcery", "Sorcery")
	body := pushGraveyardCardTyped(opp, "A Body", "Creature — Bear")
	if err := ravB2Activate(t, g, me.ID, ench, 0, game.ActivateAbilityParams{Targets: cardTarget(spell)}, "C", "C"); err == nil {
		t.Error("Necrogenesis targeted a noncreature card")
	}
	ravB2MustActivate(t, g, me.ID, ench, 0, game.ActivateAbilityParams{Targets: cardTarget(body)}, "C", "C")
	if b12ZoneOf(g, body) != game.ZoneExile {
		t.Errorf("the creature card is in %q, want exile", b12ZoneOf(g, body))
	}
	if n := tokensNamedControlledBy(g, "Saproling", me.ID); n != 1 {
		t.Errorf("%d Saprolings, want 1", n)
	}
}

// --- Sprouting Thrinax ---------------------------------------------------------------------

func TestAlaB1SproutingThrinaxLeavesThreeSaprolings(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := b12Push(g, me.ID, "Sprouting Thrinax", "Creature — Lizard", alaB1ThrinaxOracle, 3, 3)
	killCreature(t, g, me.ID, id)
	passPriorityAroundTable(t, g)
	if n := tokensNamedControlledBy(g, "Saproling", me.ID); n != 3 {
		t.Errorf("%d Saprolings, want 3", n)
	}
}

// --- Dispeller's Capsule, Courier's Capsule ---------------------------------------------------

func TestAlaB1DispellersCapsuleDestroysAnArtifactOrEnchantment(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := seatsForCombat(g)
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	capsule := b12Push(g, me.ID, "Dispeller's Capsule", "Artifact", alaB1DispellersOracle, 0, 0)
	if err := ravB2Activate(t, g, me.ID, capsule, 0, game.ActivateAbilityParams{Targets: cardTarget(bear)}, "C", "C", "W"); err == nil {
		t.Fatal("the Capsule targeted a creature")
	}
	me.ManaPool.EmptyPool()
	ench := b12Permanent(g, opp.ID, "Their Enchantment", "Enchantment")
	ravB2MustActivate(t, g, me.ID, capsule, 0, game.ActivateAbilityParams{Targets: cardTarget(ench)}, "C", "C", "W")
	if g.Battlefield.Contains(ench) {
		t.Error("the enchantment survived")
	}
	if !me.Graveyard.Contains(capsule) {
		t.Error("the Capsule was not sacrificed")
	}
}

func TestAlaB1CouriersCapsuleDrawsTwo(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	capsule := b12Push(g, me.ID, "Courier's Capsule", "Artifact", alaB1CouriersOracle, 0, 0)
	hand := me.Hand.Size()
	ravB2MustActivate(t, g, me.ID, capsule, 0, game.ActivateAbilityParams{}, "C", "U")
	if me.Hand.Size() != hand+2 {
		t.Errorf("hand %d → %d, want two draws", hand, me.Hand.Size())
	}
	if !me.Graveyard.Contains(capsule) {
		t.Error("the Capsule was not sacrificed")
	}
}

// --- Jund Charm, Grixis Charm --------------------------------------------------------------

func TestAlaB1JundCharmModes(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seatsForCombat(g)
	small := b12Creature(g, opp.ID, "Small", "Creature — Bear", 2, 2)
	big := b12Creature(g, me.ID, "Big", "Creature — Beast", 3, 3)
	castCatalogSpellWithModes(t, g, "Jund Charm", "Instant", alaB1JundCharmOracle, []int{1})
	passPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(small) {
		t.Error("a 2/2 survived 2 damage")
	}
	if !g.Battlefield.Contains(big) {
		t.Error("a 3/3 died to 2 damage")
	}

	g = newCatalogGame(t)
	me, _ = seatsForCombat(g)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	castCatalogSpellWithModesTargets(t, g, alaB1JundCharmOracle, []int{2}, cardTarget(bear))
	passPriorityAroundTable(t, g)
	if n := b12Counter(t, g, bear, game.CounterPlusOne); n != 2 {
		t.Errorf("%d +1/+1 counters, want 2", n)
	}

	g = newCatalogGame(t)
	_, opp = seatsForCombat(g)
	gone := pushGraveyardCardTyped(opp, "A Body", "Creature — Bear")
	castCatalogSpellWithModesTargets(t, g, alaB1JundCharmOracle, []int{0}, []game.TargetRef{playerTarget(opp.ID)})
	passPriorityAroundTable(t, g)
	if b12ZoneOf(g, gone) != game.ZoneExile {
		t.Errorf("the graveyard card is in %q, want exile", b12ZoneOf(g, gone))
	}
}

func TestAlaB1GrixisCharmModes(t *testing.T) {
	g := newCatalogGame(t)
	_, opp := seatsForCombat(g)
	art := b12Permanent(g, opp.ID, "Their Artifact", "Artifact")
	castCatalogSpellWithModesTargets(t, g, alaB1GrixisCharmOracle, []int{0}, cardTarget(art))
	passPriorityAroundTable(t, g)
	if !opp.Hand.Contains(art) {
		t.Error("the permanent did not return to its owner's hand")
	}

	g = newCatalogGame(t)
	_, opp = seatsForCombat(g)
	beast := b12Creature(g, opp.ID, "Beast", "Creature — Beast", 5, 5)
	castCatalogSpellWithModesTargets(t, g, alaB1GrixisCharmOracle, []int{1}, cardTarget(beast))
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, beast); p != 1 || tough != 1 {
		t.Errorf("the target is %d/%d, want 1/1", p, tough)
	}

	g = newCatalogGame(t)
	me, opp := seatsForCombat(g)
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	castCatalogSpellWithModes(t, g, "Grixis Charm", "Instant", alaB1GrixisCharmOracle, []int{2})
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, mine); p != 4 || tough != 2 {
		t.Errorf("your creature is %d/%d, want 4/2", p, tough)
	}
	if p, _ := effectivePTOf(t, g, theirs); p != 2 {
		t.Errorf("an opponent's creature has power %d, want 2", p)
	}
}

// --- Broodmate Dragon -------------------------------------------------------------------------

func TestAlaB1BroodmateDragonMakesADragon(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	castAndResolveCreature(t, g, "Broodmate Dragon", "Creature — Dragon", alaB1BroodmateOracle)
	passPriorityAroundTable(t, g)
	if n := tokensNamedControlledBy(g, "Dragon", me.ID); n != 1 {
		t.Errorf("%d Dragon tokens, want 1", n)
	}
}

// --- Blightning ---------------------------------------------------------------------------

func TestAlaB1BlightningBurnsAndMakesThePlayerDiscardTwo(t *testing.T) {
	g := newCatalogGame(t)
	_, opp := seatsForCombat(g)
	d2dSetHand(opp, 3)
	life := opp.Life
	castCatalogSpell(t, g, "Blightning", "Sorcery", alaB1BlightningOracle, []game.TargetRef{playerTarget(opp.ID)})
	passPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("life %d → %d, want 3 damage", life, opp.Life)
	}
	if discardChoiceFor(g, opp.ID) == nil {
		t.Fatal("the player was not asked to discard")
	}
	answerDiscard(t, g, opp.ID, opp.Hand.Cards[0].InstanceID, opp.Hand.Cards[1].InstanceID)
	if opp.Hand.Size() != 1 {
		t.Errorf("hand %d, want 1", opp.Hand.Size())
	}
}

// --- Sphinx Sovereign ------------------------------------------------------------------------

func TestAlaB1SphinxSovereignGainsOrDrainsByItsTappedStatus(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Sphinx Sovereign", "Artifact Creature — Sphinx", alaB1SphinxSovereignOracle, 6, 6)
	life := me.Life
	opps := lifeOfOpponents(g)
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("an untapped Sphinx: your life %d → %d, want +3", life, me.Life)
	}
	for i, l := range lifeOfOpponents(g) {
		if l != opps[i] {
			t.Errorf("an untapped Sphinx drained an opponent: %d → %d", opps[i], l)
		}
	}

	g = newCatalogGame(t)
	me = g.Seats[0]
	advanceToMain(t, g)
	sphinx := b12Push(g, me.ID, "Sphinx Sovereign", "Artifact Creature — Sphinx", alaB1SphinxSovereignOracle, 6, 6)
	findBattlefieldCardForTest(g, sphinx).Tapped = true
	life = me.Life
	opps = lifeOfOpponents(g)
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("a tapped Sphinx: your life %d → %d, want unchanged", life, me.Life)
	}
	for i, l := range lifeOfOpponents(g) {
		if l != opps[i]-3 {
			t.Errorf("a tapped Sphinx: an opponent's life %d → %d, want -3", opps[i], l)
		}
	}
}

// --- Hell's Thunder ------------------------------------------------------------------------

func TestAlaB1HellsThunderIsSacrificedAtTheEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	id := b12Push(g, me.ID, "Hell's Thunder", "Creature — Elemental", alaB1HellsThunderOracle, 4, 4)
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(id) || !me.Graveyard.Contains(id) {
		t.Error("Hell's Thunder was not sacrificed at another player's end step")
	}
}

// --- Where Ancients Tread ----------------------------------------------------------------------

func TestAlaB1WhereAncientsTreadDealsFiveForABigCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seatsForCombat(g)
	b12Push(g, me.ID, "Where Ancients Tread", "Enchantment", alaB1AncientsTreadOracle, 0, 0)
	alaB1CastCreature(t, g, "Small Bear", 4, 4)
	if latestPickTarget(g, me.ID) != nil || latestTriggerPrompt(g, me.ID) != nil {
		t.Fatal("a 4/4 entering made it trigger")
	}
	life := opp.Life
	alaB1CastCreature(t, g, "Big Beast", 5, 5)
	b04WaitForPick(t, g, me.ID)
	pickPlayer(t, g, me.ID, opp.ID)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if opp.Life != life-5 {
		t.Errorf("life %d → %d, want 5 damage", life, opp.Life)
	}
}

// --- Call to Heel ---------------------------------------------------------------------------

func TestAlaB1CallToHeelBouncesAndItsControllerDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seatsForCombat(g)
	stolen := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Stolen Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: opp.ID,
	})
	theirs := opp.Hand.Size()
	castCatalogSpell(t, g, "Call to Heel", "Instant", alaB1CallToHeelOracle, cardTarget(stolen))
	mine := me.Hand.Size()
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(stolen) {
		t.Error("the creature did not return to its owner's hand")
	}
	if opp.Hand.Size() != theirs+1 {
		t.Errorf("its controller's hand %d → %d, want one draw", theirs, opp.Hand.Size())
	}
	if me.Hand.Size() != mine+1 {
		t.Errorf("the owner's hand %d → %d, want only the returned card", mine, me.Hand.Size())
	}
}

// --- Violent Ultimatum ----------------------------------------------------------------------

func TestAlaB1ViolentUltimatumDestroysThreeTargets(t *testing.T) {
	g := newCatalogGame(t)
	_, opp := seatsForCombat(g)
	a := b12Permanent(g, opp.ID, "First", "Artifact")
	b := b12Permanent(g, opp.ID, "Second", "Enchantment")
	c := b12Creature(g, opp.ID, "Third", "Creature — Bear", 2, 2)
	spare := b12Permanent(g, opp.ID, "Spare", "Artifact")
	if err := castCatalogSpellErr(t, g, "Violent Ultimatum", "Sorcery", alaB1ViolentOracle, cardRefs(a, b)); err == nil {
		t.Error("Violent Ultimatum was cast with two targets")
	}
	castCatalogSpell(t, g, "Violent Ultimatum", "Sorcery", alaB1ViolentOracle, cardRefs(a, b, c))
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b, c} {
		if g.Battlefield.Contains(id) {
			t.Errorf("%s survived", b12Card(t, g, id).Name)
		}
	}
	if !g.Battlefield.Contains(spare) {
		t.Error("an untargeted permanent was destroyed")
	}
}

// --- Sanctum Gargoyle -------------------------------------------------------------------------

func TestAlaB1SanctumGargoyleReturnsAnArtifactCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rock := pushGraveyardCardTyped(me, "Old Rock", "Artifact")
	castAndResolveCreature(t, g, "Sanctum Gargoyle", "Artifact Creature — Gargoyle", alaB1GargoyleOracle)
	pickCard(t, g, me.ID, rock)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(rock) {
		t.Error("the artifact card did not return to hand")
	}
}

// --- Stoic Angel ------------------------------------------------------------------------------

func TestAlaB1StoicAngelLetsEachPlayerUntapOneCreature(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[1]
	a := pushTappedForTest(g, owner.ID, "Bear A", "", "Creature — Bear")
	b := pushTappedForTest(g, owner.ID, "Bear B", "", "Creature — Bear")
	land := pushTappedForTest(g, owner.ID, "Forest", "", "Basic Land — Forest")
	pushPermanentForTest(g, g.Seats[0].ID, "Stoic Angel", alaB1StoicAngelOracle, "Creature — Angel")
	choice := advanceToUntapChoiceOf(t, g, 1)
	if choice.ChooseMax != 1 || len(choice.ChooseCards) != 2 {
		t.Fatalf("Stoic Angel offers %d creatures, at most %d; want the two creatures, at most one", len(choice.ChooseCards), choice.ChooseMax)
	}
	if err := g.ResolveUntapChoice(choice.ID, owner.ID, []uuid.UUID{a}); err != nil {
		t.Fatalf("answer the untap choice: %v", err)
	}
	if tappedForTest(t, g, a) || !tappedForTest(t, g, b) {
		t.Error("the chosen creature alone should untap")
	}
	if tappedForTest(t, g, land) {
		t.Error("a land is not a creature and untaps as normal")
	}
}
