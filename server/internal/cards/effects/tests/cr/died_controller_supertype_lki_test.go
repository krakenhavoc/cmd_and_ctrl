package testscr

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardsap "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/ap"
	cardsd "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/d"
	cardss "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/s"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// died_controller_supertype_lki_test.go — #1682, CR 603.10a: a dies /
// leaves-the-battlefield condition is judged on the permanent as it
// last existed on the battlefield, and two more of its facts now ride
// the EventLTB: its SUPERTYPES (a Clone copying a legend was legendary,
// and is a plain Clone in the graveyard) and its CONTROLLER ("you
// control" asks about the permanent; a card in a graveyard has no
// controller, CR 108.4). Readers: leftAsSupertype, leftUnderControlOf.

// --- the stolen creature, end to end -------------------------------

// The issue's case: a stolen creature sacrificed by its thief died
// under the THIEF's control. The thief's Zulaport Cutthroat ("whenever
// ~ or another creature you control dies") drains the table, owner
// included; the owner's Zulaport does not trigger.
func TestStolenCreatureSacrificedByItsThiefTriggersTheThiefsYouControlDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner, thief := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, owner.ID, "Zulaport Cutthroat", "Creature — Human Rogue Ally", effectstest.ZulaportCutthroatOracle, false)
	effectstest.PushCatalogPermanent(g, thief.ID, "Zulaport Cutthroat", "Creature — Human Rogue Ally", effectstest.ZulaportCutthroatOracle, false)
	bear := effectstest.PushCatalogPermanent(g, owner.ID, "Owner's Bear", "Creature — Bear", "", false)
	g.ReadSnapshot(func() {})
	effectstest.StealForTest(t, g, bear, thief.ID)
	before := effectstest.LifeSnapshot(g)

	g.WithWriteLock(func() {
		if err := g.SacrificePermanentForEffect(bear); err != nil {
			t.Fatalf("SacrificePermanentForEffect: %v", err)
		}
	})
	if who, _ := lastLTBOf(t, g, bear).LeftUnderControlOf(); who != thief.ID {
		t.Fatalf("setup: the bear left under %s, want the thief", who)
	}
	effectstest.PassPriorityAroundTable(t, g)

	// Only the thief's Zulaport: the thief gains 1, each of the thief's
	// opponents — the owner among them — loses 1.
	want := []int{before[0] - 1, before[1] + 1, before[2] - 1, before[3] - 1}
	for i, p := range g.Seats {
		if p.Life != want[i] {
			t.Errorf("seat %d life %d → %d, want %d", i, before[i], p.Life, want[i])
		}
	}
}

// The other half: with only the OWNER's "you control" dies trigger on
// the table, the thief sacrificing the stolen creature triggers
// nothing — the owner owned it, but did not control it.
func TestStolenCreatureSacrificedByItsThiefDoesNotTriggerTheOwnersYouControlDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner, thief := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, owner.ID, "Zulaport Cutthroat", "Creature — Human Rogue Ally", effectstest.ZulaportCutthroatOracle, false)
	bear := effectstest.PushCatalogPermanent(g, owner.ID, "Owner's Bear", "Creature — Bear", "", false)
	g.ReadSnapshot(func() {})
	effectstest.StealForTest(t, g, bear, thief.ID)
	before := effectstest.LifeSnapshot(g)

	g.WithWriteLock(func() {
		if err := g.SacrificePermanentForEffect(bear); err != nil {
			t.Fatalf("SacrificePermanentForEffect: %v", err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		if p.Life != before[i] {
			t.Errorf("seat %d life %d → %d; the owner's Zulaport must not see a creature the thief controlled", i, before[i], p.Life)
		}
	}
}

// --- the legend by effect, end to end -------------------------------

// A Clone copying an opponent's legend is a legendary creature I
// control. When it dies, Rakdos Joins Up triggers — though the card in
// my graveyard is a plain, non-legendary Clone.
func TestRakdosJoinsUpSeesACloneOfALegendAsLegendary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.CastCatalogSpell(t, g, "Rakdos Joins Up", "Legendary Enchantment", effectstest.B33RakdosJoinsUpOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	legend := effectstest.SeedCopyableCreature(g, opp.ID, "Isamaru, Hound of Konda", "Legendary Creature — Dog", 2, 2)

	clone := effectstest.CastCatalogSpell(t, g, "Clone", "Creature — Shapeshifter", effectstest.OracleClone, nil)
	effectstest.ResolveWithCopyChoice(t, g, legend)
	if !effectstest.HasCopySupertype(effectstest.CopyBattlefieldCard(t, g, clone), "Legendary") {
		t.Fatal("setup: the Clone copying a legend is not legendary")
	}
	life := opp.Life

	effectstest.B25Destroy(g, clone)
	if gy, _ := g.LookupCardForEffect(clone); gy.HasSupertype("Legendary") {
		t.Fatal("setup: the Clone in the graveyard should not be legendary — the point of the test")
	}
	effectstest.B04WaitForPick(t, g, me.ID)
	if effectstest.LatestPickTarget(g, me.ID) == nil {
		t.Fatal("Rakdos Joins Up did not trigger on a Clone that died a legend")
	}
	effectstest.B16PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("opponent life %d → %d, want -2 (the copied legend's power)", life, opp.Life)
	}
}

// The other direction, read straight off the condition: a printed
// legend that left NOT legendary on the battlefield (the stamp says
// so) is not "a legendary creature" to Rakdos Joins Up, though its
// graveyard card reads as one.
func TestRakdosJoinsUpConditionReadsTheStampedSupertypesNotTheCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	legend := effectstest.B16Creature(g, me.ID, "My Legend", "Legendary Creature — Human", 2, 2, "B")
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(legend) })
	ev := lastLTBOf(t, g, legend)
	if was, known := ev.WasSupertype("Legendary"); !was || !known {
		t.Fatalf("setup: the stamp says %v, %v for a printed legend", was, known)
	}
	source := &game.Card{InstanceID: uuid.New(), Controller: me.ID, Owner: me.ID}
	if !effects.B33LegendaryCreatureYouControlDied(ev, source, g) {
		t.Error("a printed legend that died a legend did not count")
	}
	ev.LastKnownSupertypes = nil
	if effects.B33LegendaryCreatureYouControlDied(ev, source, g) {
		t.Error("the condition read the graveyard card's supertype over a stamp that says it was not legendary")
	}
}

// The dying permanents the table needs.
const (
	subjectSacrificed = "sacrificed" // everything the conditions ask of a creature, sacrificed
	subjectAttacking  = "attacking"  // the same, attacking as it died
	subjectBlocking   = "blocking"   // the same, blocking as it died
	subjectBounced    = "bounced"    // the same, returned to its owner's hand
	subjectLand       = "land"       // a land that went to the graveyard
	subjectMunitions  = "munitions"  // a Munitions token that died
)

// controllerDiesCheck is one switched "you control" / "an opponent
// controls" condition, the permanent it needs to see leave, and which
// of the two it asks.
type controllerDiesCheck struct {
	name    string
	subject string
	// opponents is true for "a creature an OPPONENT controls" (the
	// condition holds for the owner's source, whose opponent the thief
	// is), false for "a creature YOU control" (it holds for the thief's).
	opponents bool
	check     func(ev game.Event, source *game.Card, g *game.Game) bool
}

// ltbTriggerOf returns the first triggered ability of a registered
// card that watches EventLTB.
func ltbTriggerOf(t *testing.T, oracle string) game.TriggeredAbility {
	t.Helper()
	spec, ok := effects.Lookup(oracle)
	if !ok {
		t.Fatalf("%s is not registered", oracle)
	}
	for _, tr := range spec.Triggered {
		if slices.Contains(tr.Watches, game.EventLTB) {
			return tr
		}
	}
	t.Fatalf("%s has no EventLTB trigger", spec.Name)
	return game.TriggeredAbility{}
}

// controllerDiesChecks is every dies / leaves-the-battlefield
// condition #1682 switched to leftUnderControlOf. A new one belongs
// here.
func controllerDiesChecks(t *testing.T) []controllerDiesCheck {
	var none game.Characteristic
	when := func(w effects.When) func(game.Event, *game.Card, *game.Game) bool {
		return func(ev game.Event, s *game.Card, g *game.Game) bool { return w(ev, s, none, g) }
	}
	inline := func(oracle string) func(game.Event, *game.Card, *game.Game) bool {
		applies := ltbTriggerOf(t, oracle).AppliesTo
		return func(ev game.Event, s *game.Card, g *game.Game) bool { return applies(ev, s, none, g) }
	}
	massacreBuild := ltbTriggerOf(t, "93cf50cf-0ecc-4d3e-abea-778c1ebacec4").Build
	return []controllerDiesCheck{
		{"ACreatureYouControlDied (Zulaport Cutthroat)", subjectSacrificed, false, when(effects.ACreatureYouControlDied)},
		{"anEggYouControlDied (Atla Palani)", subjectSacrificed, false, when(cardsap.AnEggYouControlDied)},
		{"b16SelfOrAnotherCreatureYouControlDied (Vengeful Bloodwitch)", subjectSacrificed, false, effects.B16SelfOrAnotherCreatureYouControlDied},
		{"b16AnotherCreatureYouControlEnteredOrDied (Daxos)", subjectSacrificed, false, effects.B16AnotherCreatureYouControlEnteredOrDied},
		{"b17SelfOrZombieYouControlDied (Undead Augur)", subjectSacrificed, false, effects.B17SelfOrZombieYouControlDied},
		{"b18OpponentsCreatureDied (Sangromancer)", subjectSacrificed, true, effects.B18OpponentsCreatureDied},
		{"b19AnotherNontokenCreatureYouControlDied (Liesa)", subjectSacrificed, false, effects.B19AnotherNontokenCreatureYouControlDied},
		{"b22CreatureYouControlDied (Cauldron of Essence)", subjectSacrificed, false, effects.B22CreatureYouControlDied},
		{"b22SlimedCreatureYouDontControlDied (Toxrill)", subjectSacrificed, true, effects.B22SlimedCreatureYouDontControlDied},
		{"b25AnotherGoblinYouControlDied (Pashalik Mons)", subjectSacrificed, false, effects.B25AnotherGoblinYouControlDied},
		{"b27SelfOrNontokenZombieYouControlDied (Headless Rider)", subjectSacrificed, false, effects.B27SelfOrNontokenZombieYouControlDied},
		{"b33LegendaryCreatureYouControlDied (Rakdos Joins Up)", subjectSacrificed, false, effects.B33LegendaryCreatureYouControlDied},
		{"b33OpponentsCreatureDied (Patron of the Vein)", subjectSacrificed, true, effects.B33OpponentsCreatureDied},
		{"b33OpponentsNontokenCreatureDied (Overseer of the Damned)", subjectSacrificed, true, effects.B33OpponentsNontokenCreatureDied},
		{"b34VampireYouControlDied (Crossway Troublemakers)", subjectSacrificed, false, effects.B34VampireYouControlDied},
		{"anotherZombieYouControlDied (Diregraf Captain)", subjectSacrificed, false, effects.AnotherZombieYouControlDied},
		{"anotherCreatureYouControlDied (Pitiless Plunderer)", subjectSacrificed, false, effects.AnotherCreatureYouControlDied},
		{"b36AngelYouControlDied (Bishop of Wings)", subjectSacrificed, false, effects.B36AngelYouControlDied},
		{"b36AnotherFaerieYouControlDied (Tegwyll)", subjectSacrificed, false, effects.B36AnotherFaerieYouControlDied},
		{"b39AnotherCreatureYouControlDied", subjectSacrificed, false, effects.B39AnotherCreatureYouControlDied},
		{"edeaCreatureYouControlButDontOwnDied (Edea)", subjectSacrificed, false, when(effects.EdeaCreatureYouControlButDontOwnDied)},
		{"b21ArtifactOrCreatureYouControlDied (Agent of the Iron Throne)", subjectSacrificed, false, func(ev game.Event, s *game.Card, g *game.Game) bool {
			_, ok := effects.B21ArtifactOrCreatureYouControlDied(ev, s, g)
			return ok
		}},
		{"b14PermanentYouControlLeft (Resourceful Defense)", subjectSacrificed, false, effects.B14PermanentYouControlLeft},
		{"aCreatureYouControlLeft (Outpost Siege)", subjectSacrificed, false, when(effects.ACreatureYouControlLeft)},
		{"scrapTrawlerArtifactHitTheYard (Scrap Trawler)", subjectSacrificed, false, when(cardss.ScrapTrawlerArtifactHitTheYard)},
		{"Massacre Wurm's condition", subjectSacrificed, true, inline("93cf50cf-0ecc-4d3e-abea-778c1ebacec4")},
		// "That player" is the dead creature's controller: an opponent
		// of the Wurm's controller exactly when the condition holds.
		{"Massacre Wurm's \"that player\"", subjectSacrificed, true, func(ev game.Event, s *game.Card, g *game.Game) bool {
			item := massacreBuild(ev, s, none, g)
			return item != nil && item.Params.Player != s.Controller
		}},
		{"Midnight Reaper", subjectSacrificed, false, inline("e8c7566d-7cc0-48af-a986-83223ec7e06c")},
		{"Omnath, Locus of Rage", subjectSacrificed, false, inline("1816eede-c5bd-49df-958f-a3af64cb2932")},
		{"Open the Graves", subjectSacrificed, false, inline("28778958-a1f9-4fea-b551-c193d1257f18")},
		{"Pawn of Ulamog", subjectSacrificed, false, inline("9bcaf141-1f1f-491f-aced-13dc093b9e2c")},
		{"Teysa, Orzhov Scion", subjectSacrificed, false, inline("8191342b-b25e-4c4d-8f69-aee662148ff4")},
		{"Vraan, Executioner Thane", subjectSacrificed, false, inline("b2f2645f-5f74-456a-bd02-83169d8b8a7e")},
		{"Yahenni, Undying Partisan", subjectSacrificed, true, inline("fdba89eb-1cf5-46e6-9d09-1adb9bc40fcd")},
		{"Cruel Celebrant", subjectSacrificed, false, inline("3ee78cfc-0e9e-4737-a7e2-b42f94228040")},
		{"The Ozolith", subjectSacrificed, false, inline("1946ded1-5f53-409f-b0a6-5433bb0357d2")},
		{"Ares, God of War", subjectAttacking, false, inline("8a09d4a2-4796-4a15-ab75-a3a0b717f62d")},
		{"deathTyrantCombatDeath — an attacker you control", subjectAttacking, false, when(cardsd.DeathTyrantCombatDeath)},
		{"deathTyrantCombatDeath — a blocker an opponent controls", subjectBlocking, true, when(cardsd.DeathTyrantCombatDeath)},
		{"creatureYouControlLeftWithoutDying (Aang, Dour Port-Mage)", subjectBounced, false, effects.CreatureYouControlLeftWithoutDying},
		{"b10LandYouControlDied (Titania)", subjectLand, false, effects.B10LandYouControlDied},
		{"Nadier's Nightblade", subjectMunitions, false, inline("391978f6-0bbc-41e8-9246-f7d0e21c7900")},
	}
}

// seedControllerSubject puts the permanent a subject names on the
// battlefield under owner's control and returns it.
func seedControllerSubject(t *testing.T, g *game.Game, owner uuid.UUID, subject string) uuid.UUID {
	t.Helper()
	c := game.Card{
		InstanceID: uuid.New(), Owner: owner, Controller: owner,
		Name: "Owner's Legend", Colors: []string{"B"}, Power: 2, Toughness: 2,
		// Every tribe, type and supertype the switched conditions ask
		// about, so the only thing left to vary is the controller.
		TypeLine: "Legendary Artifact Creature — Zombie Goblin Vampire Angel Faerie Egg Elemental",
	}
	switch subject {
	case subjectLand:
		c.Name, c.TypeLine, c.Colors, c.Power, c.Toughness = "Owner's Land", "Land", nil, 0, 0
	case subjectMunitions:
		c.Name, c.TypeLine, c.Colors, c.Power, c.Toughness = "Munitions", "Token Artifact", nil, 0, 0
	}
	id := effectstest.PushBattlefieldCardWithTimestamp(g, c)
	if subject != subjectLand && subject != subjectMunitions {
		// "If it had counters on it" (Resourceful Defense, The
		// Ozolith) and Toxrill's slime counter.
		g.WithWriteLock(func() {
			for _, name := range []string{"+1/+1", "slime"} {
				if err := g.AddCounterForEffect(id, name, 1); err != nil {
					t.Fatalf("setup: AddCounterForEffect %s: %v", name, err)
				}
			}
		})
	}
	return id
}

// Every switched condition, against a permanent the OWNER owns that
// the THIEF controlled as it left. At harvest time the probe rewrites
// the moved card's Controller to its owner — what the card reads once
// "a card in a graveyard has no controller" (CR 108.4) is honoured by
// the move, or once a reanimation has written a new controller onto
// it — so a condition that still reads the card, not the event, gives
// the owner's answer and fails here. The stamp says the thief:
//
//   - "you control": true for the thief's source, false for the owner's;
//   - "an opponent controls": the reverse.
func TestControllerDiesChecksReadTheLastKnownController(t *testing.T) {
	for _, c := range controllerDiesChecks(t) {
		t.Run(c.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			owner, thief := g.Seats[0], g.Seats[1]
			effectstest.PushCatalogPermanent(g, g.Seats[2].ID, "LKI Probe", "Artifact", effectstest.LtbControllerProbeOracle, false)
			subject := seedControllerSubject(t, g, owner.ID, c.subject)
			effectstest.StealForTest(t, g, subject, thief.ID)
			g.WithWriteLock(func() {
				for i := range g.Battlefield.Cards {
					bc := &g.Battlefield.Cards[i]
					if bc.InstanceID != subject {
						continue
					}
					switch c.subject {
					case subjectAttacking:
						bc.AttackingTarget = owner.ID
					case subjectBlocking:
						bc.BlockingTarget = uuid.New()
					}
				}
			})

			thiefSource := &game.Card{InstanceID: uuid.New(), Controller: thief.ID, Owner: thief.ID}
			ownerSource := &game.Card{InstanceID: uuid.New(), Controller: owner.ID, Owner: owner.ID}
			var got map[string]bool
			effectstest.LtbControllerProbeHook = func(ev game.Event, g *game.Game) {
				if ev.CardID != subject {
					return
				}
				if z := g.FindCardZoneForEffect(subject); z != nil {
					for i := range z.Cards {
						if z.Cards[i].InstanceID == subject {
							z.Cards[i].Controller = owner.ID
						}
					}
				}
				got = map[string]bool{
					"thief's source": c.check(ev, thiefSource, g),
					"owner's source": c.check(ev, ownerSource, g),
				}
			}
			t.Cleanup(func() { effectstest.LtbControllerProbeHook = nil })

			g.WithWriteLock(func() {
				var err error
				switch c.subject {
				case subjectSacrificed:
					err = g.SacrificePermanentForEffect(subject)
				case subjectBounced:
					err = g.BounceToHandForEffect(subject)
				default:
					err = g.DestroyPermanentForEffect(subject)
				}
				if err != nil {
					t.Fatalf("exit: %v", err)
				}
			})
			if got == nil {
				t.Fatal("the probe never saw the subject leave")
			}
			if who, _ := lastLTBOf(t, g, subject).LeftUnderControlOf(); who != thief.ID {
				t.Fatalf("setup: the subject left under %s, want the thief", who)
			}
			for label, want := range map[string]bool{"thief's source": !c.opponents, "owner's source": c.opponents} {
				if got[label] != want {
					t.Errorf("%s: %s = %v, want %v", label, c.name, got[label], want)
				}
			}
		})
	}
}
