package testsa

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// akh_5535_b1_cards_test.go — the first batch of Amonkhet cards (#5535).
// Every test plays the real catalog entry through the engine.

const (
	akhB1ColossapedeOracle       = "b2cede1a-3a76-4658-bd12-e9df464d61e5"
	akhB1DuneBeetleOracle        = "3f79603a-3173-404e-8c69-b2254c1270dc"
	akhB1HyenaPackOracle         = "d38b027d-eb24-46ba-9d46-3be08eedbe2a"
	akhB1ThoseWhoServeOracle     = "0ee088dc-c7c7-4f7f-9d1f-be3d79550d95"
	akhB1CursedMinotaurOracle    = "88b50194-bc39-47a6-968b-983553d89cf9"
	akhB1PouncingCheetahOracle   = "1d5e32b1-1a2d-4b3d-80f2-f8c688b90587"
	akhB1ScaledBehemothOracle    = "abb39d0e-bf11-4982-b06a-ca7be16bce2c"
	akhB1SeraphOfTheSunsOracle   = "8354b70e-43fc-4581-bb53-b913335bf460"
	akhB1NimbleBladeKhenraOracle = "09c68fb5-2000-4bb8-b03b-94b093e50bca"
	akhB1WeaverOfCurrentsOracle  = "9aabb520-7e97-47b2-975c-11f457d0b870"
	akhB1AvenInitiateOracle      = "c872012d-b598-4076-8385-2e3771c0306f"
	akhB1HonoredHydraOracle      = "db63d999-97a6-4a62-83bd-e0677e6f8956"
	akhB1TahCropOracle           = "6fab0fa0-d082-414e-9bcf-edc0fc1f6211"
	akhB1UnwaveringOracle        = "e6088c23-a660-4d2c-b633-1c6cb1a59ebe"
	akhB1TrueheartDuelistOracle  = "a0185db5-8988-4f63-8f3c-a19b4807a431"
	akhB1OketrasAttendantOracle  = "c45d1df1-92a7-4b36-8e27-1f88c3737bf9"
	akhB1DesertCerodonOracle     = "6ad18589-ece4-4809-80f6-6764d2822372"
	akhB1ShimmerscaleOracle      = "f3bd5ca1-9854-4bc0-b108-b555808e1d5f"
	akhB1WastelandScorpionOracle = "0bc418c3-ebde-4e52-9284-65d7e3433f27"
	akhB1WingedShepherdOracle    = "e18c93bc-2073-4bb6-83b4-c5c70f2d8a82"
	akhB1GreaterSandwurmOracle   = "d1dfed64-6e88-44c2-b479-9a2d072e4be7"
	akhB1CensorOracle            = "3a42558c-7432-4c05-9599-95343eb09be1"
	akhB1DissentersOracle        = "1b73639b-2b07-4e82-b990-ca7721074e43"
	akhB1ForsakeOracle           = "f60fb81a-969a-476c-a227-5231bbed4ad4"
	akhB1HieroglyphicOracle      = "60e358ec-eb11-4a16-a0ae-b7c064a8ea6c"
	akhB1HazeOfPollenOracle      = "f0140f00-5414-493b-962d-33312d49f6ea"
	akhB1PursueGloryOracle       = "f85edb4f-1eec-40cc-9e21-0a6fb71e8499"
	akhB1ViolentImpactOracle     = "d48d9e93-72a2-48f2-8164-8032615ee217"
	akhB1WanderInDeathOracle     = "f25c5451-58fa-4916-8342-ae88c5e406cc"
	akhB1LimitsOracle            = "a2184c37-ee92-40f2-b7d1-97c39ce5f277"
	akhB1StingingShotOracle      = "7fd7d20d-d44c-443c-a2cf-320f39a3bec4"
	akhB1ElectrifyOracle         = "da952213-9a3c-4e13-bc5a-55d1e5dbaa5b"
	akhB1FinalRewardOracle       = "e654242f-c7c5-4713-bbd0-26d41de8e2e7"
	akhB1PainfulLessonOracle     = "6be78177-aef0-408c-9769-a1be0775f06d"
	akhB1SynchronizedOracle      = "59d9070c-dc60-420e-aa0e-32cfb97dfe88"
	akhB1DecisionParalysisOracle = "6e09ceaf-9005-4b06-ac5a-5e2250d4b663"
	akhB1BonePickerOracle        = "2edf106d-f504-430e-ad9c-50fc95b0ad9c"
	akhB1FanBearerOracle         = "d2e7198a-0eb5-4d30-973f-4b2d9617436f"
	akhB1FesteringMummyOracle    = "bdf78365-5887-497e-af25-ca0cb731f35a"
	akhB1GracefulCatOracle       = "69fffad0-6469-4623-9977-347143fa7289"
	akhB1KhenraCharioteerOracle  = "e75bc5df-f4a4-45af-a6fd-2711b98dd521"
	akhB1NagaOracleOracle        = "249e18a4-12c5-4836-a172-cb4e9b9cb7d7"
	akhB1CrypticSerpentOracle    = "0f26c04a-9bf7-4cbd-a95f-73ee4a3b1af1"
	akhB1WarfireOracle           = "33a7a1e9-23b4-4cde-baf1-b1c0165c6729"
	akhB1BalefulAmmitOracle      = "1cfa9f7d-ef1f-4907-aeb1-15c0df33fccc"
	akhB1CrocodileOracle         = "e5e31d27-3907-4c65-a988-a754f217e061"
	akhB1OrneryKuduOracle        = "8abef87a-4a5c-412e-8137-827cee50bb8b"
	akhB1HekmaSentinelsOracle    = "90ea758d-3783-47a5-bb18-9ff9377113fb"
	akhB1ShadowstormOracle       = "b3e5b559-660c-4e14-9209-29efc07d3563"
	akhB1HonedKhopeshOracle      = "f2142c41-ff66-4276-b426-fd945ed9b0f6"

	akhB1BearType = "Creature — Bear"
)

// akhB1GraveyardCard puts a plain card of the given type line into p's
// graveyard.
func akhB1GraveyardCard(p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: p.ID, Controller: p.ID})
	return id
}

// akhB1Tap taps a battlefield permanent.
func akhB1Tap(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, id).Tapped = true })
}

// akhB1Token returns the token copy named `name` that p controls, other
// than the card `not`.
func akhB1Token(g *game.Game, p *game.Player, name string, not uuid.UUID) *game.Card {
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Name == name && c.Controller == p.ID && c.InstanceID != not {
			return c
		}
	}
	return nil
}

// --- every card is in the catalog ----------------------------------------

func TestAkh5535B1EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{
		akhB1ColossapedeOracle, akhB1DuneBeetleOracle, akhB1HyenaPackOracle, akhB1ThoseWhoServeOracle,
		akhB1CursedMinotaurOracle, akhB1PouncingCheetahOracle, akhB1ScaledBehemothOracle, akhB1SeraphOfTheSunsOracle,
		akhB1NimbleBladeKhenraOracle, akhB1WeaverOfCurrentsOracle, akhB1AvenInitiateOracle, akhB1HonoredHydraOracle,
		akhB1TahCropOracle, akhB1UnwaveringOracle, akhB1TrueheartDuelistOracle, akhB1OketrasAttendantOracle,
		akhB1DesertCerodonOracle, akhB1ShimmerscaleOracle, akhB1WastelandScorpionOracle, akhB1WingedShepherdOracle,
		akhB1GreaterSandwurmOracle, akhB1CensorOracle, akhB1DissentersOracle, akhB1ForsakeOracle,
		akhB1HieroglyphicOracle, akhB1HazeOfPollenOracle, akhB1PursueGloryOracle, akhB1ViolentImpactOracle,
		akhB1WanderInDeathOracle, akhB1LimitsOracle, akhB1StingingShotOracle, akhB1ElectrifyOracle,
		akhB1FinalRewardOracle, akhB1PainfulLessonOracle, akhB1SynchronizedOracle, akhB1DecisionParalysisOracle,
		akhB1BonePickerOracle, akhB1FanBearerOracle, akhB1FesteringMummyOracle, akhB1GracefulCatOracle,
		akhB1KhenraCharioteerOracle, akhB1NagaOracleOracle, akhB1CrypticSerpentOracle, akhB1WarfireOracle,
		akhB1BalefulAmmitOracle, akhB1CrocodileOracle, akhB1OrneryKuduOracle, akhB1HekmaSentinelsOracle,
		akhB1ShadowstormOracle, akhB1HonedKhopeshOracle,
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", oracle)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v with %d caveats, want Full with none", spec.Name, spec.Completeness, len(spec.Caveats))
		}
	}
}

// --- printed keywords ------------------------------------------------------

func TestAkh5535B1KeywordCreatures(t *testing.T) {
	cases := []struct {
		name, typeLine, oracle string
		keywords               []string
	}{
		{"Cursed Minotaur", "Creature — Zombie Minotaur", akhB1CursedMinotaurOracle, []string{"menace"}},
		{"Pouncing Cheetah", "Creature — Cat", akhB1PouncingCheetahOracle, []string{"flash"}},
		{"Scaled Behemoth", "Creature — Crocodile", akhB1ScaledBehemothOracle, []string{"hexproof"}},
		{"Seraph of the Suns", "Creature — Angel", akhB1SeraphOfTheSunsOracle, []string{"flying", "indestructible"}},
		{"Nimble-Blade Khenra", "Creature — Jackal Warrior", akhB1NimbleBladeKhenraOracle, []string{"prowess"}},
		{"Aven Initiate", "Creature — Bird Warrior", akhB1AvenInitiateOracle, []string{"flying"}},
		{"Honored Hydra", "Creature — Snake Hydra", akhB1HonoredHydraOracle, []string{"trample"}},
		{"Unwavering Initiate", "Creature — Human Warrior", akhB1UnwaveringOracle, []string{"vigilance"}},
		{"Oketra's Attendant", "Creature — Bird Soldier", akhB1OketrasAttendantOracle, []string{"flying"}},
		{"Shimmerscale Drake", "Creature — Drake", akhB1ShimmerscaleOracle, []string{"flying"}},
		{"Wasteland Scorpion", "Creature — Scorpion", akhB1WastelandScorpionOracle, []string{"deathtouch"}},
		{"Winged Shepherd", "Creature — Angel", akhB1WingedShepherdOracle, []string{"flying", "vigilance"}},
		{"Bone Picker", "Creature — Bird", akhB1BonePickerOracle, []string{"flying", "deathtouch"}},
		{"Khenra Charioteer", "Creature — Jackal Warrior", akhB1KhenraCharioteerOracle, []string{"trample"}},
		{"Baleful Ammit", "Creature — Crocodile Demon", akhB1BalefulAmmitOracle, []string{"lifelink"}},
		{"Crocodile of the Crossing", "Creature — Crocodile", akhB1CrocodileOracle, []string{"haste"}},
		{"Shadowstorm Vizier", "Creature — Human Cleric", akhB1ShadowstormOracle, []string{"flying"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			id := effectstest.B12Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
			abilities := effectstest.EffectiveAbilities(t, g, id)
			for _, kw := range tc.keywords {
				if !effectstest.HasAbility(abilities, kw) {
					t.Errorf("%s has no %s", tc.name, kw)
				}
			}
		})
	}
}

// Flash: Pouncing Cheetah is cast outside a main phase.
func TestAkh5535B1PouncingCheetahHasFlash(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.B13HandCard(me, "Pouncing Cheetah", "Creature — Cat", akhB1PouncingCheetahOracle)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("Pouncing Cheetah at the beginning of combat: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Error("Pouncing Cheetah did not resolve")
	}
}

// Weaver of Currents taps for two colourless mana.
func TestAkh5535B1WeaverOfCurrentsAddsTwoColorless(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	weaver := effectstest.B12Push(g, me.ID, "Weaver of Currents", "Creature — Snake Druid", akhB1WeaverOfCurrentsOracle, 2, 2)
	effectstest.B40TapForMana(t, g, me.ID, weaver, 0, "")
	if n := me.ManaPool.Total(); n != 2 {
		t.Errorf("%d mana in the pool, want 2", n)
	}
	if !effectstest.HmlTapped(t, g, weaver) {
		t.Error("the Weaver did not tap")
	}
}

// --- embalm ------------------------------------------------------------------

// Each embalm cost exiles the card from the graveyard and makes a white
// Zombie token copy of it.
func TestAkh5535B1EmbalmMakesAWhiteZombieCopy(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, mana string
		index                        int
	}{
		{"Aven Initiate", "Creature — Bird Warrior", akhB1AvenInitiateOracle, "{C}{C}{C}{C}{C}{C}{U}", 0},
		{"Honored Hydra", "Creature — Snake Hydra", akhB1HonoredHydraOracle, "{C}{C}{C}{G}", 0},
		{"Tah-Crop Skirmisher", "Creature — Snake Warrior", akhB1TahCropOracle, "{C}{C}{C}{U}", 0},
		{"Unwavering Initiate", "Creature — Human Warrior", akhB1UnwaveringOracle, "{C}{C}{C}{C}{W}", 0},
		{"Trueheart Duelist", "Creature — Human Warrior", akhB1TrueheartDuelistOracle, "{C}{C}{W}", 0},
		{"Oketra's Attendant", "Creature — Bird Soldier", akhB1OketrasAttendantOracle, "{C}{C}{C}{W}{W}", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			effectstest.AdvanceToMain(t, g)
			me := effectstest.Top100ActiveSeat(g)
			id := effectstest.PushCatalogGraveyardCard(me, tc.name, tc.typeLine, tc.oracle, 2, 2)
			if err := g.AddManaForEffect(me.ID, uuid.Nil, tc.mana); err != nil {
				t.Fatalf("AddManaForEffect: %v", err)
			}
			if err := g.ActivateCatalogAbility(me.ID, id, tc.index, game.ActivateAbilityParams{}); err != nil {
				t.Fatalf("embalm: %v", err)
			}
			if !g.Exile.Contains(id) {
				t.Fatal("the embalmed card is not in exile")
			}
			effectstest.PassPriorityAroundTable(t, g)
			token := akhB1Token(g, me, tc.name, id)
			if token == nil {
				t.Fatal("no embalm token was created")
			}
			if !strings.Contains(token.TypeLine, "Zombie") {
				t.Errorf("token type line %q is not a Zombie", token.TypeLine)
			}
			if len(token.Colors) != 1 || token.Colors[0] != "W" {
				t.Errorf("token colours %v, want [W]", token.Colors)
			}
		})
	}
}

// --- cycling ------------------------------------------------------------------

// Every cycling card in the batch discards itself from hand for its
// cost and draws a card.
func TestAkh5535B1CyclingDrawsACard(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle, mana string }{
		{"Oketra's Attendant", "Creature — Bird Soldier", akhB1OketrasAttendantOracle, "{C}{C}"},
		{"Desert Cerodon", "Creature — Beast", akhB1DesertCerodonOracle, "{R}"},
		{"Shimmerscale Drake", "Creature — Drake", akhB1ShimmerscaleOracle, "{C}{C}"},
		{"Wasteland Scorpion", "Creature — Scorpion", akhB1WastelandScorpionOracle, "{C}{C}"},
		{"Winged Shepherd", "Creature — Angel", akhB1WingedShepherdOracle, "{W}"},
		{"Greater Sandwurm", "Creature — Wurm", akhB1GreaterSandwurmOracle, "{C}{C}"},
		{"Censor", "Instant", akhB1CensorOracle, "{U}"},
		{"Dissenter's Deliverance", "Instant", akhB1DissentersOracle, "{G}"},
		{"Forsake the Worldly", "Instant", akhB1ForsakeOracle, "{C}{C}"},
		{"Hieroglyphic Illumination", "Instant", akhB1HieroglyphicOracle, "{U}"},
		{"Haze of Pollen", "Instant", akhB1HazeOfPollenOracle, "{C}{C}{C}"},
		{"Pursue Glory", "Instant", akhB1PursueGloryOracle, "{C}{C}"},
		{"Violent Impact", "Sorcery", akhB1ViolentImpactOracle, "{C}{C}"},
		{"Wander in Death", "Sorcery", akhB1WanderInDeathOracle, "{C}{C}"},
		{"Limits of Solidarity", "Sorcery", akhB1LimitsOracle, "{C}{C}"},
		{"Stinging Shot", "Instant", akhB1StingingShotOracle, "{C}{C}"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			id, me := effectstest.CycleFromHand(t, g, tc.name, tc.typeLine, tc.oracle, tc.mana)
			if !me.Graveyard.Contains(id) {
				t.Fatal("the cycled card is not in the graveyard")
			}
			before := len(me.Hand.Cards)
			effectstest.PassPriorityAroundTable(t, g)
			if got := len(me.Hand.Cards); got != before+1 {
				t.Errorf("hand %d -> %d, want the cycling draw", before, got)
			}
		})
	}
}

// --- statics ------------------------------------------------------------------

func TestAkh5535B1TrueheartDuelistBlocksTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	duelist := effectstest.B12Push(g, g.Seats[1].ID, "Trueheart Duelist", "Creature — Human Warrior", akhB1TrueheartDuelistOracle, 2, 2)
	if got := effectstest.MbCapacity(t, g, duelist); got != 2 {
		t.Errorf("capacity = %d, want 2", got)
	}
}

func TestAkh5535B1GreaterSandwurmEvadesSmallBlockers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wurm := effectstest.B12Push(g, me.ID, "Greater Sandwurm", "Creature — Wurm", akhB1GreaterSandwurmOracle, 7, 7)
	small := effectstest.B12Creature(g, opp.ID, "Small", akhB1BearType, 2, 2)
	big := effectstest.B12Creature(g, opp.ID, "Big", akhB1BearType, 3, 3)
	effectstest.BrAttack(t, g, wurm)
	if err := g.DeclareBlocker(small, wurm); err == nil {
		t.Error("a creature with power 2 blocked the Sandwurm")
	}
	if err := g.DeclareBlocker(big, wurm); err != nil {
		t.Errorf("a creature with power 3 can't block the Sandwurm: %v", err)
	}
}

func TestAkh5535B1KhenraCharioteerGivesYourOtherCreaturesTrample(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Khenra Charioteer", "Creature — Jackal Warrior", akhB1KhenraCharioteerOracle, 3, 3)
	mine := effectstest.B12Creature(g, me.ID, "Mine", akhB1BearType, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", akhB1BearType, 2, 2)
	if !effectstest.HasEffectiveKeyword(t, g, mine, "trample") {
		t.Error("your other creature has no trample")
	}
	if effectstest.HasEffectiveKeyword(t, g, theirs, "trample") {
		t.Error("an opponent's creature has trample")
	}
}

func TestAkh5535B1HonedKhopeshPumpsTheEquippedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	khopesh := effectstest.PushCatalogPermanent(g, me.ID, "Honed Khopesh", "Artifact — Equipment", akhB1HonedKhopeshOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", akhB1BearType, 2, 2)
	effectstest.EquipTo(t, g, me.ID, khopesh, bear)
	if p, tough := effectstest.PtOf(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("the equipped Bear is %d/%d, want 3/3", p, tough)
	}
}

// --- cost reductions ----------------------------------------------------------

func TestAkh5535B1BonePickerCostsThreeLessAfterACreatureDied(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	if mv := effectstest.SelfPricedMV(t, g, me, akhB1BonePickerOracle, "Creature — Bird", "{3}{B}"); mv != 4 {
		t.Fatalf("with no creature dead this turn Bone Picker costs %d, want 4", mv)
	}
	victim := effectstest.B12Creature(g, opp.ID, "Their Bear", akhB1BearType, 2, 2)
	effectstest.B15Destroy(t, g, victim)
	if mv := effectstest.SelfPricedMV(t, g, me, akhB1BonePickerOracle, "Creature — Bird", "{3}{B}"); mv != 1 {
		t.Errorf("after a creature died Bone Picker costs %d, want 1 ({B})", mv)
	}
}

func TestAkh5535B1CrypticSerpentCostsLessPerInstantAndSorcery(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	akhB1GraveyardCard(me, "Old Instant", "Instant")
	akhB1GraveyardCard(me, "Old Sorcery", "Sorcery")
	akhB1GraveyardCard(me, "Old Bear", akhB1BearType)
	if mv := effectstest.SelfPricedMV(t, g, me, akhB1CrypticSerpentOracle, "Creature — Serpent", "{5}{U}{U}"); mv != 5 {
		t.Errorf("with an instant and a sorcery in the graveyard the Serpent costs %d, want 5", mv)
	}
}

// --- triggered abilities -------------------------------------------------------

func TestAkh5535B1NagaOracleSurveilsThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.CastHoldCreature(t, g, "Naga Oracle", "Creature — Snake Cleric", akhB1NagaOracleOracle, 2, 4)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SurveilChoiceFor(g, me.ID) == nil {
		t.Error("no surveil prompt")
	}
}

func TestAkh5535B1WarfireJavelineerDealsXToAnOpponentsCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	akhB1GraveyardCard(me, "Old Instant", "Instant")
	akhB1GraveyardCard(me, "Old Sorcery", "Sorcery")
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	effectstest.CastHoldCreature(t, g, "Warfire Javelineer", "Creature — Minotaur Warrior", akhB1WarfireOracle, 2, 3)
	effectstest.AnswerPickTargetIfOpen(t, g, ogre)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.FindCardAnywhere(t, g, ogre).DamageMarked; got != 2 {
		t.Errorf("the Ogre has %d damage, want 2", got)
	}
}

// "Put a -1/-1 counter on target creature you control" lands on the
// chosen creature.
func TestAkh5535B1EntersPutsACounterOnYourCreature(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		power, toughness       int
	}{
		{"Baleful Ammit", "Creature — Crocodile Demon", akhB1BalefulAmmitOracle, 4, 3},
		{"Crocodile of the Crossing", "Creature — Crocodile", akhB1CrocodileOracle, 5, 4},
		{"Ornery Kudu", "Creature — Antelope", akhB1OrneryKuduOracle, 3, 4},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := effectstest.Top100ActiveSeat(g)
			bear := effectstest.B12Creature(g, me.ID, "Bear", akhB1BearType, 3, 3)
			effectstest.CastHoldCreature(t, g, tc.name, tc.typeLine, tc.oracle, tc.power, tc.toughness)
			effectstest.AnswerPickTargetIfOpen(t, g, bear)
			effectstest.PassPriorityAroundTable(t, g)
			if got := effectstest.CounterCount(g, bear, game.CounterMinusOne); got != 1 {
				t.Errorf("-1/-1 counters on the Bear = %d, want 1", got)
			}
		})
	}
}

func TestAkh5535B1FesteringMummyMayShrinkACreatureAsItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	mummy := effectstest.B12Push(g, me.ID, "Festering Mummy", "Creature — Zombie", akhB1FesteringMummyOracle, 1, 1)
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	effectstest.KillCreature(t, g, me.ID, mummy)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickCard(t, g, me.ID, ogre)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, ogre, game.CounterMinusOne); got != 1 {
		t.Errorf("-1/-1 counters on the target = %d, want 1", got)
	}
}

func TestAkh5535B1GracefulCatGrowsWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	cat := effectstest.B12Push(g, me.ID, "Graceful Cat", "Creature — Cat", akhB1GracefulCatOracle, 2, 2)
	effectstest.DeclareAttack(t, g, g.Seats[1].ID, cat)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, cat); p != 3 || tough != 3 {
		t.Errorf("the attacking Cat is %d/%d, want 3/3", p, tough)
	}
}

// Cycling a card discards it, which is what "cycle or discard" watches.
func TestAkh5535B1CycleOrDiscardPumps(t *testing.T) {
	for _, tc := range []struct{ name, oracle string }{
		{"Hekma Sentinels", akhB1HekmaSentinelsOracle},
		{"Shadowstorm Vizier", akhB1ShadowstormOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			id := effectstest.B12Push(g, me.ID, tc.name, "Creature — Human Cleric", tc.oracle, 2, 3)
			effectstest.CycleFromHand(t, g, "Desert Cerodon", "Creature — Beast", akhB1DesertCerodonOracle, "{R}")
			effectstest.PassPriorityAroundTable(t, g)
			if p, tough := effectstest.PtOf(t, g, id); p != 3 || tough != 4 {
				t.Errorf("%s is %d/%d after a cycle, want 3/4", tc.name, p, tough)
			}
		})
	}
}

// --- activated abilities --------------------------------------------------------

func TestAkh5535B1FanBearerTapsACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	bearer := effectstest.B12Push(g, me.ID, "Fan Bearer", "Creature — Zombie", akhB1FanBearerOracle, 1, 2)
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	effectstest.ApaMana(me, "C", "C")
	effectstest.ApaActivate(t, g, me, bearer, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(ogre)})
	if !effectstest.HmlTapped(t, g, ogre) {
		t.Error("the target is untapped")
	}
	if !effectstest.HmlTapped(t, g, bearer) {
		t.Error("Fan Bearer did not tap")
	}
}

// --- the spells -----------------------------------------------------------------

func TestAkh5535B1CensorCountersUnlessPaid(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bolt := effectstest.B13OpponentCasts(t, g, opp, "Their Bolt", "Instant", "", "{R}", nil)
	id := effectstest.B13HandCard(me, "Censor", "Instant", akhB1CensorOracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: effectstest.B16TargetCard(bolt)}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, opp.ID) {
		effectstest.AnswerPayUnless(t, g, opp.ID, false)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if !opp.Graveyard.Contains(bolt) {
		t.Error("the unpaid spell was not countered")
	}
}

func TestAkh5535B1DestroysTheRightPermanent(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		good, bad              string
	}{
		{"Dissenter's Deliverance", "Instant", akhB1DissentersOracle, "Artifact", "Enchantment"},
		{"Violent Impact", "Sorcery", akhB1ViolentImpactOracle, "Land", "Enchantment"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			effectstest.AdvanceToMain(t, g)
			opp := effectstest.Top100Opponent(g)
			good := effectstest.B12Permanent(g, opp.ID, "Good", tc.good)
			bad := effectstest.B12Permanent(g, opp.ID, "Bad", tc.bad)
			if err := effectstest.CastCatalogSpellErr(t, g, tc.name, tc.typeLine, tc.oracle, effectstest.CardTarget(bad)); err == nil {
				t.Fatalf("%s targeted a %s", tc.name, tc.bad)
			}
			effectstest.CastCatalogSpell(t, g, tc.name, tc.typeLine, tc.oracle, effectstest.CardTarget(good))
			effectstest.PassPriorityAroundTable(t, g)
			if g.Battlefield.Contains(good) {
				t.Errorf("the %s survived", tc.good)
			}
		})
	}
}

func TestAkh5535B1ForsakeTheWorldlyExilesAnEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	opp := effectstest.Top100Opponent(g)
	charm := effectstest.B12Permanent(g, opp.ID, "Charm", "Enchantment")
	land := effectstest.B12Permanent(g, opp.ID, "Land", "Land")
	if err := effectstest.CastCatalogSpellErr(t, g, "Forsake the Worldly", "Instant", akhB1ForsakeOracle, effectstest.CardTarget(land)); err == nil {
		t.Fatal("Forsake the Worldly targeted a land")
	}
	effectstest.CastCatalogSpell(t, g, "Forsake the Worldly", "Instant", akhB1ForsakeOracle, effectstest.CardTarget(charm))
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(charm) {
		t.Error("the enchantment was not exiled")
	}
}

func TestAkh5535B1FinalRewardExilesACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	opp := effectstest.Top100Opponent(g)
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	effectstest.CastCatalogSpell(t, g, "Final Reward", "Instant", akhB1FinalRewardOracle, effectstest.CardTarget(ogre))
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(ogre) {
		t.Error("the creature was not exiled")
	}
}

func TestAkh5535B1ElectrifyDealsFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	opp := effectstest.Top100Opponent(g)
	giant := effectstest.B12Creature(g, opp.ID, "Giant", "Creature — Giant", 5, 5)
	effectstest.CastCatalogSpell(t, g, "Electrify", "Instant", akhB1ElectrifyOracle, effectstest.CardTarget(giant))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.FindCardAnywhere(t, g, giant).DamageMarked; got != 4 {
		t.Errorf("the Giant has %d damage, want 4", got)
	}
}

func TestAkh5535B1HieroglyphicIlluminationDrawsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	effectstest.CastCatalogSpell(t, g, "Hieroglyphic Illumination", "Instant", akhB1HieroglyphicOracle, nil)
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d, want 2", got)
	}
}

func TestAkh5535B1PainfulLessonDrawsTwoAndLosesTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	opp := effectstest.Top100Opponent(g)
	hand, life := opp.Hand.Size(), opp.Life
	effectstest.CastCatalogSpell(t, g, "Painful Lesson", "Sorcery", akhB1PainfulLessonOracle, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.Hand.Size() - hand; got != 2 {
		t.Errorf("the target drew %d, want 2", got)
	}
	if opp.Life != life-2 {
		t.Errorf("the target is at %d, want %d", opp.Life, life-2)
	}
}

func TestAkh5535B1HazeOfPollenPreventsCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.CastCatalogSpell(t, g, "Haze of Pollen", "Instant", akhB1HazeOfPollenOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	defender := effectstest.B12Creature(g, me.ID, "Defender", akhB1BearType, 2, 2)
	attacker := effectstest.B12Creature(g, opp.ID, "Attacker", akhB1BearType, 2, 2)
	if err := g.MarkCombatDamage(attacker, defender, 2); err != nil {
		t.Fatalf("MarkCombatDamage: %v", err)
	}
	if got := effectstest.FindCardAnywhere(t, g, defender).DamageMarked; got != 0 {
		t.Errorf("the defender has %d combat damage, want 0", got)
	}
}

func TestAkh5535B1PursueGloryPumpsAttackers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	raider := effectstest.B12Creature(g, me.ID, "Raider", akhB1BearType, 2, 2)
	idle := effectstest.B12Creature(g, me.ID, "Idle", akhB1BearType, 2, 2)
	effectstest.DeclareAttack(t, g, g.Seats[1].ID, raider)
	id := effectstest.B13HandCard(me, "Pursue Glory", "Instant", akhB1PursueGloryOracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.PtOf(t, g, raider); p != 4 {
		t.Errorf("the attacker has power %d, want 4", p)
	}
	if p, _ := effectstest.PtOf(t, g, idle); p != 2 {
		t.Errorf("a creature that isn't attacking has power %d, want 2", p)
	}
}

func TestAkh5535B1WanderInDeathReturnsTwoCreatureCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	one := akhB1GraveyardCard(me, "One", akhB1BearType)
	two := akhB1GraveyardCard(me, "Two", akhB1BearType)
	rock := akhB1GraveyardCard(me, "Rock", "Artifact")
	if err := effectstest.CastCatalogSpellErr(t, g, "Wander in Death", "Sorcery", akhB1WanderInDeathOracle, effectstest.CardTarget(rock)); err == nil {
		t.Fatal("Wander in Death targeted an artifact card")
	}
	effectstest.CastCatalogSpell(t, g, "Wander in Death", "Sorcery", akhB1WanderInDeathOracle, effectstest.CardTargets(one, two))
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(one) || !me.Hand.Contains(two) {
		t.Error("the creature cards did not return to hand")
	}
}

func TestAkh5535B1LimitsOfSolidarityThreatens(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	akhB1Tap(g, ogre)
	effectstest.CastCatalogSpell(t, g, "Limits of Solidarity", "Sorcery", akhB1LimitsOracle, effectstest.CardTarget(ogre))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.ApaLive(g, ogre).Controller; got != me.ID {
		t.Error("control of the creature did not change")
	}
	if effectstest.HmlTapped(t, g, ogre) {
		t.Error("the creature was not untapped")
	}
	if !effectstest.HasEffectiveKeyword(t, g, ogre, "haste") {
		t.Error("the creature has no haste")
	}
}

func TestAkh5535B1StingingShotHitsOnlyAFlyer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	opp := effectstest.Top100Opponent(g)
	walker := effectstest.B12Creature(g, opp.ID, "Walker", akhB1BearType, 5, 5)
	flyer := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Flyer", TypeLine: "Creature — Bird", Power: 5, Toughness: 5, Keywords: []string{"flying"}})
	if err := effectstest.CastCatalogSpellErr(t, g, "Stinging Shot", "Instant", akhB1StingingShotOracle, effectstest.CardTarget(walker)); err == nil {
		t.Fatal("Stinging Shot targeted a creature without flying")
	}
	effectstest.CastCatalogSpell(t, g, "Stinging Shot", "Instant", akhB1StingingShotOracle, effectstest.CardTarget(flyer))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, flyer, game.CounterMinusOne); got != 3 {
		t.Errorf("-1/-1 counters on the flyer = %d, want 3", got)
	}
}

func TestAkh5535B1SynchronizedStrikeUntapsAndPumpsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	a := effectstest.B12Creature(g, me.ID, "A", akhB1BearType, 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", akhB1BearType, 2, 2)
	akhB1Tap(g, a)
	akhB1Tap(g, b)
	effectstest.CastCatalogSpell(t, g, "Synchronized Strike", "Instant", akhB1SynchronizedOracle, effectstest.CardTargets(a, b))
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if effectstest.HmlTapped(t, g, id) {
			t.Error("a target is still tapped")
		}
		if p, tough := effectstest.PtOf(t, g, id); p != 4 || tough != 4 {
			t.Errorf("a target is %d/%d, want 4/4", p, tough)
		}
	}
}

func TestAkh5535B1DecisionParalysisTapsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	opp := effectstest.Top100Opponent(g)
	a := effectstest.B12Creature(g, opp.ID, "A", akhB1BearType, 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "B", akhB1BearType, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Decision Paralysis", "Instant", akhB1DecisionParalysisOracle, effectstest.CardTargets(a, b))
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if !effectstest.HmlTapped(t, g, id) {
			t.Error("a target is untapped")
		}
	}
}
