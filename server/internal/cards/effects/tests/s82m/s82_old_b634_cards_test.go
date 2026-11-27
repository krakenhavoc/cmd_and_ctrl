package testss82m

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b634_cards_test.go — Apocalypse, batch 634 (#5614). Each test plays
// the real catalog entry through the engine and asserts what its text says.

const (
	oldB634StrengthOfNight   = "0045cf16-86dd-4417-ae36-88ca63b30c26"
	oldB634Jilt              = "382bc554-d170-4c19-82dd-829f254e855e"
	oldB634DwarvenLandslide  = "5231a7d6-b6cf-4fa8-8da8-f0ecf023e054"
	oldB634DesolationAngel   = "87f82107-eaba-484a-800d-e61a0cc5e1f2"
	oldB634NecraSanctuary    = "2586a59d-8501-4c22-9d69-f4bf91de7024"
	oldB634GerrardCapashen   = "27521e70-3a1f-4819-a111-c38216ffe7c3"
	oldB634DegaSanctuary     = "75c626fb-9dfc-4a94-b59f-f0e45c7b2f56"
	oldB634MournfulZombie    = "f0ed2cf1-7bd4-4586-a682-fd588632cb3d"
	oldB634GraveDefiler      = "42c5c1aa-10a4-4cd5-b1f8-a393ed553f99"
	oldB634KavuHowler        = "635c0a7d-d832-4c10-a5f0-9cde19152cbb"
	oldB634TidalCourier      = "b0a0c17b-87c8-4610-adca-5e7f5ced49f6"
	oldB634EnlistmentOfficer = "e9517fc3-d1bc-42f7-b052-9f629b302735"
)

var oldB634Cards = map[string]string{
	oldB634StrengthOfNight: "Strength of Night", oldB634Jilt: "Jilt", oldB634DwarvenLandslide: "Dwarven Landslide",
	oldB634DesolationAngel: "Desolation Angel", oldB634NecraSanctuary: "Necra Sanctuary",
	oldB634GerrardCapashen: "Gerrard Capashen", oldB634DegaSanctuary: "Dega Sanctuary",
	oldB634MournfulZombie: "Mournful Zombie", oldB634GraveDefiler: "Grave Defiler", oldB634KavuHowler: "Kavu Howler",
	oldB634TidalCourier: "Tidal Courier", oldB634EnlistmentOfficer: "Enlistment Officer",
}

func TestOldB634EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range oldB634Cards {
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

func oldB634Main(t *testing.T) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	return g, g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

func oldB634Cast(t *testing.T, g *game.Game, me *game.Player, name, typeLine, cost, oracle, mana string, params game.CastSpellParams) (uuid.UUID, error) {
	t.Helper()
	effectstest.FloatMana(t, g, me, mana)
	params.Strict = true
	id, err := effectstest.CastWithTapParams(t, g, name, typeLine, cost, oracle, params)
	if err == nil {
		effectstest.PassPriorityAroundTable(t, g)
	}
	return id, err
}

func oldB634PT(t *testing.T, g *game.Game, id uuid.UUID) (int, int) {
	t.Helper()
	return effectstest.EffectivePTOf(t, g, id)
}

func TestOldB634StrengthOfNightPumpsAllAndZombiesWhenKicked(t *testing.T) {
	for _, kicked := range []bool{false, true} {
		g, me, opp := oldB634Main(t)
		bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		zombie := effectstest.B12Creature(g, me.ID, "Zed", "Creature — Zombie", 2, 2)
		theirs := effectstest.B12Creature(g, opp.ID, "Their Zombie", "Creature — Zombie", 2, 2)
		params, mana := game.CastSpellParams{}, "{C}{C}{G}"
		if kicked {
			params.OptionalCosts, mana = []int{0}, "{C}{C}{G}{B}"
		}
		if _, err := oldB634Cast(t, g, me, "Strength of Night", "Instant", "{2}{G}", oldB634StrengthOfNight, mana, params); err != nil {
			t.Fatalf("cast (kicked=%v): %v", kicked, err)
		}
		wantZ := 3
		if kicked {
			wantZ = 5
		}
		if p, th := oldB634PT(t, g, bear); p != 3 || th != 3 {
			t.Errorf("kicked=%v: the Bear is %d/%d, want 3/3", kicked, p, th)
		}
		if p, th := oldB634PT(t, g, zombie); p != wantZ || th != wantZ {
			t.Errorf("kicked=%v: the Zombie is %d/%d, want %d/%d", kicked, p, th, wantZ, wantZ)
		}
		if p, th := oldB634PT(t, g, theirs); p != 2 || th != 2 {
			t.Errorf("kicked=%v: the opponent's Zombie is %d/%d, want 2/2", kicked, p, th)
		}
	}
}

func TestOldB634JiltBouncesAndKickedDamagesAnotherCreature(t *testing.T) {
	g, me, opp := oldB634Main(t)
	bounced := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 4, 4)
	if _, err := oldB634Cast(t, g, me, "Jilt", "Instant", "{1}{U}", oldB634Jilt, "{C}{U}", game.CastSpellParams{Targets: effectstest.CardTarget(bounced)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if g.Battlefield.Contains(bounced) || !opp.Hand.Contains(bounced) {
		t.Error("the creature was not returned to its owner's hand")
	}

	g, me, opp = oldB634Main(t)
	bounced = effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 4, 4)
	hit := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if _, err := oldB634Cast(t, g, me, "Jilt", "Instant", "{1}{U}", oldB634Jilt, "{C}{U}{C}{R}", game.CastSpellParams{
		OptionalCosts: []int{0}, Targets: effectstest.CardRefs(bounced, bounced)}); err == nil {
		t.Error("the same creature was accepted as both Jilt targets")
	}
	if _, err := oldB634Cast(t, g, me, "Jilt", "Instant", "{1}{U}", oldB634Jilt, "{C}{U}{C}{R}", game.CastSpellParams{
		OptionalCosts: []int{0}, Targets: effectstest.CardRefs(bounced, hit)}); err != nil {
		t.Fatalf("kicked cast: %v", err)
	}
	if g.Battlefield.Contains(bounced) || !opp.Hand.Contains(bounced) {
		t.Error("kicked: the first target was not bounced")
	}
	if g.Battlefield.Contains(hit) {
		t.Error("kicked: the other creature survived 2 damage")
	}
}

func oldB634Land(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	return effectstest.PushLandFor(g, owner, name, "Basic Land — Mountain")
}

func TestOldB634DwarvenLandslideKickedSacrificesALandAndDestroysTwo(t *testing.T) {
	g, me, opp := oldB634Main(t)
	a, b := oldB634Land(g, opp.ID, "A"), oldB634Land(g, opp.ID, "B")
	if _, err := oldB634Cast(t, g, me, "Dwarven Landslide", "Sorcery", "{3}{R}", oldB634DwarvenLandslide, "{C}{C}{C}{R}", game.CastSpellParams{Targets: effectstest.CardTarget(a)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if g.Battlefield.Contains(a) || !g.Battlefield.Contains(b) {
		t.Error("unkicked: exactly the targeted land should be destroyed")
	}

	g, me, opp = oldB634Main(t)
	a, b = oldB634Land(g, opp.ID, "A"), oldB634Land(g, opp.ID, "B")
	mine := oldB634Land(g, me.ID, "Mine")
	mana := "{C}{C}{C}{R}{C}{C}{R}"
	if _, err := oldB634Cast(t, g, me, "Dwarven Landslide", "Sorcery", "{3}{R}", oldB634DwarvenLandslide, mana, game.CastSpellParams{
		OptionalCosts: []int{0}, Targets: effectstest.CardRefs(a, b)}); err == nil {
		t.Fatal("the kicker was accepted without a land to sacrifice")
	}
	if _, err := oldB634Cast(t, g, me, "Dwarven Landslide", "Sorcery", "{3}{R}", oldB634DwarvenLandslide, "", game.CastSpellParams{
		OptionalCosts: []int{0}, Targets: effectstest.CardRefs(a, b), SacrificeIDs: []uuid.UUID{mine}}); err != nil {
		t.Fatalf("kicked cast: %v", err)
	}
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Error("kicked: both target lands should be destroyed")
	}
	if g.Battlefield.Contains(mine) {
		t.Error("kicked: the sacrificed land is still on the battlefield")
	}
}

func TestOldB634DesolationAngelDestroysYourLandsAndAllWhenKicked(t *testing.T) {
	for _, kicked := range []bool{false, true} {
		g, me, opp := oldB634Main(t)
		mine, theirs := oldB634Land(g, me.ID, "Mine"), oldB634Land(g, opp.ID, "Theirs")
		params, mana := game.CastSpellParams{}, "{C}{C}{C}{B}{B}"
		if kicked {
			params.OptionalCosts, mana = []int{0}, "{C}{C}{C}{B}{B}{W}{W}"
		}
		if _, err := oldB634Cast(t, g, me, "Desolation Angel", "Creature — Angel", "{3}{B}{B}", oldB634DesolationAngel, mana, params); err != nil {
			t.Fatalf("cast (kicked=%v): %v", kicked, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(mine) {
			t.Errorf("kicked=%v: my land survived", kicked)
		}
		if g.Battlefield.Contains(theirs) != !kicked {
			t.Errorf("kicked=%v: the opponent's land on the battlefield = %v", kicked, g.Battlefield.Contains(theirs))
		}
	}
}

func oldB634UpkeepAfter(t *testing.T, g *game.Game) {
	t.Helper()
	effectstest.WalkToUpkeepOfSeat0(t, g)
}

func TestOldB634NecraSanctuaryDrainsOneOrThree(t *testing.T) {
	for _, tc := range []struct {
		colours []string
		loss    int
	}{{[]string{"G"}, 1}, {[]string{"W"}, 1}, {[]string{"G", "W"}, 3}, {[]string{"R"}, 0}} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.B12Push(g, me.ID, "Necra Sanctuary", "Enchantment", oldB634NecraSanctuary, 0, 0)
		for _, c := range tc.colours {
			effectstest.TmpB1Push(g, me.ID, "Rock "+c, "Artifact", "", 0, 0, c)
		}
		life := opp.Life
		oldB634UpkeepAfter(t, g)
		oldB634SettlePlayer(t, g, me.ID, opp.ID)
		if opp.Life != life-tc.loss {
			t.Errorf("%v: life %d -> %d, want -%d", tc.colours, life, opp.Life, tc.loss)
		}
	}
}

// oldB634SettlePlayer answers a target pick with a player and settles.
func oldB634SettlePlayer(t *testing.T, g *game.Game, chooser, player uuid.UUID) {
	t.Helper()
	for i := 0; i < 48; i++ {
		if p := effectstest.LatestPickTarget(g, chooser); p != nil {
			if err := g.ResolvePickTarget(p.ID, chooser, game.TargetRef{Kind: game.TargetPlayer, ID: player}); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
			continue
		}
		if len(g.PendingTriggers) > 0 {
			g.RunStateChecksForTest()
			if effectstest.LatestPickTarget(g, chooser) != nil {
				continue
			}
		}
		if len(g.PendingChoices) > 0 || effectstest.StackFullyEmpty(g) {
			return
		}
		if err := g.PassPriority(); err != nil {
			if errors.Is(err, game.ErrChoicePending) {
				continue
			}
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatal("the stack did not settle")
}

func TestOldB634DegaSanctuaryGainsTwoOrFour(t *testing.T) {
	for _, tc := range []struct {
		colours []string
		gain    int
	}{{[]string{"B"}, 2}, {[]string{"R"}, 2}, {[]string{"B", "R"}, 4}, {[]string{"G"}, 0}} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.B12Push(g, me.ID, "Dega Sanctuary", "Enchantment", oldB634DegaSanctuary, 0, 0)
		for _, c := range tc.colours {
			effectstest.TmpB1Push(g, me.ID, "Rock "+c, "Artifact", "", 0, 0, c)
		}
		life := me.Life
		oldB634UpkeepAfter(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		if me.Life != life+tc.gain {
			t.Errorf("%v: life %d -> %d, want +%d", tc.colours, life, me.Life, tc.gain)
		}
	}
}

func TestOldB634GerrardGainsForTheOpponentsHandAndTapsOnlyWhileAttacking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gerrard := effectstest.B12Push(g, me.ID, "Gerrard Capashen", "Legendary Creature — Human Soldier", oldB634GerrardCapashen, 3, 4)
	victim := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	for i := 0; i < 3; i++ {
		opp.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Card", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
	}
	life := me.Life
	oldB634UpkeepAfter(t, g)
	hand := opp.Hand.Size()
	oldB634SettlePlayer(t, g, me.ID, opp.ID)
	if me.Life != life+hand {
		t.Errorf("life %d -> %d, want +%d (one per card in the opponent's hand)", life, me.Life, hand)
	}
	if err := oldB378Activate(t, g, me, gerrard, 0, "{C}{C}{C}{W}", effectstest.CardTarget(victim)); err == nil {
		t.Fatal("the tap ability was activated while Gerrard was not attacking")
	}
	effectstest.DeclareAttack(t, g, opp.ID, gerrard)
	if err := oldB378Activate(t, g, me, gerrard, 0, "{C}{C}{C}{W}", effectstest.CardTarget(victim)); err != nil {
		t.Fatalf("activate while attacking: %v", err)
	}
	if !effectstest.TappedNow(g, victim) {
		t.Error("the target creature was not tapped")
	}
}

func TestOldB634MournfulZombieGainsOneForATargetPlayer(t *testing.T) {
	g, me, opp := oldB634Main(t)
	zombie := effectstest.PushCatalogPermanent(g, me.ID, "Mournful Zombie", "Creature — Zombie", oldB634MournfulZombie, false)
	life := opp.Life
	effectstest.FloatMana(t, g, me, "{W}")
	if err := g.ActivateCatalogAbility(me.ID, zombie, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life+1 {
		t.Errorf("life %d -> %d, want +1", life, opp.Life)
	}
	if !effectstest.TappedNow(g, zombie) {
		t.Error("the Zombie did not tap for its cost")
	}
}

func TestOldB634RevealFourCreaturesTakeTheirOwnType(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, sub string
	}{
		{"Grave Defiler", "Creature — Zombie", oldB634GraveDefiler, "Zombie"},
		{"Kavu Howler", "Creature — Kavu", oldB634KavuHowler, "Kavu"},
		{"Tidal Courier", "Creature — Merfolk", oldB634TidalCourier, "Merfolk"},
		{"Enlistment Officer", "Creature — Human Soldier", oldB634EnlistmentOfficer, "Soldier"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			ids := effectstest.SeedSearchLibrary(me,
				game.Card{Name: "Deep", TypeLine: "Land"},
				game.Card{Name: "Match D", TypeLine: "Creature — " + tc.sub},
				game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
				game.Card{Name: "Match B", TypeLine: "Creature — Elf " + tc.sub},
				game.Card{Name: "Island", TypeLine: "Basic Land — Island"})
			effectstest.EnterFromHand(t, g, me.ID, tc.name, tc.typeLine, tc.oracle)
			effectstest.PassPriorityAroundTable(t, g)
			pick := effectstest.PutInLibraryChoiceFor(g, me.ID)
			if pick == nil {
				t.Fatal("the controller was not asked to order the rest")
			}
			if err := g.ResolvePutInLibrary(pick.ID, me.ID, pick.ScryCards, nil); err != nil {
				t.Fatalf("order the rest: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if !me.Hand.Contains(ids[1]) || !me.Hand.Contains(ids[3]) {
				t.Errorf("the %s cards were not put into the hand", tc.sub)
			}
			if me.Hand.Contains(ids[0]) || me.Hand.Contains(ids[2]) || me.Hand.Contains(ids[4]) {
				t.Error("a card of another type went to the hand")
			}
			if !me.Library.Contains(ids[2]) || !me.Library.Contains(ids[4]) || !me.Library.Contains(ids[0]) {
				t.Error("the rest did not stay in the library")
			}
			if top := effectstest.LibraryTopIDs(me, 1); len(top) != 1 || top[0] != ids[4] {
				t.Errorf("top of library %v, want the unrevealed card", top)
			}
		})
	}
}

func TestOldB634GraveDefilerRegeneratesAndTidalCourierFlies(t *testing.T) {
	g, me, _ := oldB634Main(t)
	courier := effectstest.PushCatalogPermanent(g, me.ID, "Tidal Courier", "Creature — Merfolk", oldB634TidalCourier, false)
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{U}")
	if err := g.ActivateCatalogAbility(me.ID, courier, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, courier, "flying") {
		t.Error("Tidal Courier has no flying")
	}
	spec, ok := effects.Lookup(oldB634GraveDefiler)
	if !ok || len(spec.Activated) != 1 {
		t.Fatalf("Grave Defiler has %d activated abilities, want its regeneration", len(spec.Activated))
	}
}
