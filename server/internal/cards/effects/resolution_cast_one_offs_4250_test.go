package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// resolution_cast_one_offs_4250_test.go — #4250, part 2. Every one-off
// "you may cast it" a resolution offers, and every shared door those cards
// use, asks the resolution_cast prompt inside that resolution (CR 608.2g).
// One table, three checks per card:
//
//   - cast: the card is cast through the prompt, and the cast is logged
//     before the resolution that offered it finishes;
//   - decline: the empty answer leaves the card where it was, with no
//     permission left over, and lets the resolution finish;
//   - restore: a STRICT restore while the prompt is open (the table is a
//     restore point) keeps the offer, and the cast is made on the
//     restored table.

// oneOffOffer is a table brought to the open offer: who casts, the card on
// offer, the cast's parameters, and the zone a declined card stays in.
type oneOffOffer struct {
	g      *game.Game
	me     uuid.UUID
	card   uuid.UUID
	params game.CastSpellParams
	stays  game.ZoneKind
}

type oneOffCase struct {
	name    string
	keyword string
	setup   func(t *testing.T) oneOffOffer
}

func oneOffCases() []oneOffCase {
	exileFree := game.CastSpellParams{FromZone: "exile", Strict: true}
	return []oneOffCase{
		{"Torrential Gearhulk", game.MayCastKeywordFreeCast, func(t *testing.T) oneOffOffer {
			g := newCatalogGame(t)
			me := g.Seats[0]
			instant := uuid.New()
			me.Graveyard.PushTop(game.Card{InstanceID: instant, Name: "Old Instant", TypeLine: "Instant",
				ManaCost: "{3}{U}{U}", Owner: me.ID, Controller: me.ID})
			castCatalogSpell(t, g, "Torrential Gearhulk", "Artifact Creature — Construct", ysAGearhulkOracle, nil)
			passPriorityAroundTable(t, g)
			answerAnyPendingTargetPrompts(t, g)
			passPriorityAroundTable(t, g)
			return oneOffOffer{g, me.ID, instant, game.CastSpellParams{FromZone: "graveyard", Strict: true,
				AlternativeCost: game.AltCostKeyGrantedCost}, game.ZoneGraveyard}
		}},
		{"Arcane Heist", game.MayCastKeywordFreeCast, func(t *testing.T) oneOffOffer {
			g := newCatalogGame(t)
			me, opp := ciaSeats(t, g)
			third := g.Seats[(g.Turn.ActiveSeat+2)%4]
			bolt := gyCard(opp, "Lightning Bolt", "Instant", "{R}", lightningBoltOracle)
			castCatalogSpell(t, g, "Arcane Heist", "Sorcery", ciaArcaneHeistOracle,
				[]game.TargetRef{{Kind: game.TargetCard, ID: bolt}})
			passPriorityAroundTable(t, g)
			return oneOffOffer{g, me.ID, bolt, game.CastSpellParams{FromZone: "graveyard", Strict: true,
				AlternativeCost: game.AltCostKeyGrantedCost,
				Targets:         []game.TargetRef{{Kind: game.TargetPlayer, ID: third.ID}}}, game.ZoneGraveyard}
		}},
		{"Memory Plunder", game.MayCastKeywordFreeCast, func(t *testing.T) oneOffOffer {
			g := newCatalogGame(t)
			advanceToMain(t, g)
			me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
			third := g.Seats[(g.Turn.ActiveSeat+2)%4]
			bolt := gyCard(opp, "Lightning Bolt", "Instant", "{R}", lightningBoltOracle)
			castCatalogSpell(t, g, "Memory Plunder", "Instant", nghaMemoryPlunderOracle,
				[]game.TargetRef{{Kind: game.TargetCard, ID: bolt}})
			passPriorityAroundTable(t, g)
			return oneOffOffer{g, me.ID, bolt, game.CastSpellParams{FromZone: "graveyard", Strict: true,
				AlternativeCost: game.AltCostKeyGrantedCost,
				Targets:         []game.TargetRef{{Kind: game.TargetPlayer, ID: third.ID}}}, game.ZoneGraveyard}
		}},
		{"Emet-Selch of the Third Seat", game.MayCastKeywordPaidCast, func(t *testing.T) oneOffOffer {
			g, me, opp, emet := emetSetup(t)
			div := gyCard(me, "Divination", "Sorcery", "{2}{U}", divinationOracle)
			emetOffer(t, g, me, opp, emet, div)
			floatForTest(g, me, "U")
			return oneOffOffer{g, me.ID, div, game.CastSpellParams{FromZone: "graveyard", Strict: true,
				AlternativeCost: game.AltCostKeyGrantedCost}, game.ZoneGraveyard}
		}},
		{"Strago and Relm", game.MayCastKeywordFreeCast, func(t *testing.T) oneOffOffer {
			g, me, _, _, _, bear := stragoGame(t)
			return oneOffOffer{g, me.ID, bear, exileFree, game.ZoneExile}
		}},
		{"Crabomination", game.MayCastKeywordFreeCast, func(t *testing.T) oneOffOffer {
			g, me, opp := emergeTable(t)
			bear := func() game.Card {
				return game.Card{InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}",
					Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID}
			}
			var lib, gy, hand game.Card
			g.WithWriteLock(func() {
				lib, gy, hand = bear(), bear(), bear()
				opp.Library.PushTop(lib)
				opp.Graveyard.PushTop(gy)
				opp.Hand.Cards = []game.Card{hand}
			})
			rock := apaPush(g, me.ID, me.ID, game.Card{Name: "Rock", TypeLine: "Artifact", ManaCost: "{5}"})
			crab := handCardOf(g, me, "Crabomination", "Creature — Crab Demon", "{4}{B}{B}", crabominationOracle)
			if err := castEmerge(g, me, crab, rock, false); err != nil {
				t.Fatalf("emerge from artifact: %v", err)
			}
			for i := 0; i < 8 && findBattlefieldCardForTest(g, crab) == nil; i++ {
				if err := g.PassPriority(); err != nil {
					t.Fatalf("PassPriority: %v", err)
				}
			}
			if p := latestPickTarget(g, me.ID); p != nil {
				if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}); err != nil {
					t.Fatalf("pick the opponent: %v", err)
				}
			}
			passPriorityAroundTable(t, g)
			return oneOffOffer{g, me.ID, hand.InstanceID, exileFree, game.ZoneExile}
		}},
		{"Amped Raptor", game.MayCastKeywordPaidCast, func(t *testing.T) oneOffOffer {
			g, me, _ := p7Table(t)
			hit := uuid.New()
			g.WithWriteLock(func() {
				me.Library.PushTop(game.Card{InstanceID: hit, Name: "Two Drop", TypeLine: "Creature — Test", ManaCost: "{1}{G}",
					Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
			})
			castCatalogSpell(t, g, "Amped Raptor", "Creature — Dinosaur", ampedRaptorOracle, nil)
			passPriorityAroundTable(t, g)
			return oneOffOffer{g, me.ID, hit, game.CastSpellParams{FromZone: "exile", Strict: true,
				AlternativeCost: ampedRaptorAltCostKey}, game.ZoneExile}
		}},
		{"The Infamous Cruelclaw", game.MayCastKeywordPaidCast, func(t *testing.T) oneOffOffer {
			g, me, opp := discardTable(t)
			hit := uuid.New()
			g.WithWriteLock(func() {
				me.Library.PushTop(game.Card{InstanceID: hit, Name: "Two Drop", TypeLine: "Creature — Test", ManaCost: "{1}{G}",
					Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
			})
			spare := handCardOf(g, me, "Spare", "Sorcery", "{2}", "")
			claw := pushCatalogPermanent(g, me.ID, "The Infamous Cruelclaw", "Legendary Creature — Weasel Mercenary", cruelclawOracle, false)
			dealCombatDamageToPlayer(g, claw, opp.ID, 3)
			passPriorityAroundTable(t, g)
			return oneOffOffer{g, me.ID, hit, game.CastSpellParams{FromZone: "exile", Strict: true,
				AlternativeCost: cruelclawAltCostKey, AltCostIDs: []uuid.UUID{spare}}, game.ZoneExile}
		}},
		{"Primeval Spawn", game.MayCastKeywordFreeCast, func(t *testing.T) oneOffOffer {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			me.Library.Cards = nil
			a := pushLibraryCardForTest(me, game.Card{Name: "A", TypeLine: "Artifact", ManaCost: "{2}"})
			spawnLeaves(t, g, me)
			return oneOffOffer{g, me.ID, a, exileFree, game.ZoneExile}
		}},
		{"Fevered Suspicion", game.MayCastKeywordFreeCast, func(t *testing.T) oneOffOffer {
			g := newCatalogGame(t)
			var first uuid.UUID
			for i := 1; i < len(g.Seats); i++ {
				opp := g.Seats[(g.Turn.ActiveSeat+i)%len(g.Seats)]
				spell := cardNamed("Stolen Sorcery", "Sorcery")
				spell.ManaCost = "{5}"
				libraryTop(opp, spell, cardNamed("Exiled Land", "Land"))
				if first == uuid.Nil {
					first = spell.InstanceID
				}
			}
			me, _ := castReboundSpellAtMain(t, g, "Fevered Suspicion", "Sorcery", feveredSuspicionOracle)
			passPriorityAroundTable(t, g)
			return oneOffOffer{g, me.ID, first, exileFree, game.ZoneExile}
		}},
		{"Kefka, Dancing Mad", game.MayCastKeywordFreeCast, func(t *testing.T) oneOffOffer {
			g, _ := kefkaGame(t)
			spell := graveyardSpell(g.Seats[1], "Their Sorcery", "{3}{U}")
			advanceToEndStepOf(t, g, 0)
			passPriorityAroundTable(t, g)
			return oneOffOffer{g, g.Seats[0].ID, spell, exileFree, game.ZoneExile}
		}},
		{"Conduit of Worlds", game.MayCastKeywordPaidCast, func(t *testing.T) oneOffOffer {
			g, me, conduit, bear, _, _ := conduitTable(t)
			conduitMana(t, g, me, "{C}{G}{R}")
			if err := conduitActivate(t, g, me, conduit, bear); err != nil {
				t.Fatalf("activate: %v", err)
			}
			return oneOffOffer{g, me.ID, bear, game.CastSpellParams{FromZone: "graveyard", Strict: true}, game.ZoneGraveyard}
		}},
		{"Hazoret's Undying Fury", game.MayCastKeywordFreeCast, func(t *testing.T) oneOffOffer {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			five, _, _, _ := furyLibrary(me)
			furyCast(t, g)
			return oneOffOffer{g, me.ID, five, exileFree, game.ZoneExile}
		}},
		{"Etali, Primal Conqueror", game.MayCastKeywordFreeCast, func(t *testing.T) oneOffOffer {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			var mine uuid.UUID
			for _, p := range g.Seats {
				id := sweepFTop(p, "Spell of "+p.Name, "Sorcery")
				if p.ID == me.ID {
					mine = id
				}
			}
			importToBattlefield(t, g, etaliRow(), me)
			passPriorityAroundTable(t, g)
			return oneOffOffer{g, me.ID, mine, exileFree, game.ZoneExile}
		}},
		{"Goliath Daydreamer", game.MayCastKeywordFreeCast, func(t *testing.T) oneOffOffer {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			advanceToMain(t, g)
			goliath := pushCatalogPermanent(g, me.ID, "Goliath Daydreamer", "Creature — Giant Wizard", goliathOracle, false)
			dream := uuid.New()
			g.WithWriteLock(func() {
				g.Exile.PushTop(game.Card{InstanceID: dream, Name: "Divination", TypeLine: "Sorcery", ManaCost: "{2}{U}",
					OracleID: spellControlDivinationOracle, Owner: me.ID, Controller: me.ID,
					Counters: map[string]int{game.CounterDream: 1}})
			})
			declareAttack(t, g, opp.ID, goliath)
			passPriorityAroundTable(t, g)
			return oneOffOffer{g, me.ID, dream, exileFree, game.ZoneExile}
		}},
		{"Power Pack", game.MayCastKeywordFreeCast, func(t *testing.T) oneOffOffer {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			advanceToMain(t, g)
			pp := b12Push(g, me.ID, "Power Pack", "Legendary Creature — Human Hero", powerPackOracle, 4, 4)
			div := ppGraveyardCard(g, me, "Divination", "Sorcery", spellControlDivinationOracle)
			dealCombatDamageToPlayer(g, pp, opp.ID, 4)
			passPriorityAroundTable(t, g)
			advanceToUpkeepOfSeat(t, g, 0)
			passPriorityAroundTable(t, g)
			return oneOffOffer{g, me.ID, div, game.CastSpellParams{FromZone: "exile", Strict: true,
				AlternativeCost: game.AltCostKeyGrantedCost}, game.ZoneExile}
		}},
		{"Summon: Esper Valigarmanda", game.MayCastKeywordPaidCast, func(t *testing.T) oneOffOffer {
			g := newCatalogGame(t)
			seat := g.Turn.ActiveSeat
			me := g.Seats[seat]
			valigarmandaGraveCard(me, "Blue Spell", "Instant", "{U}{U}")
			saga := valigarmandaCast(t, g)
			passPriorityAroundTable(t, g)
			var spell uuid.UUID
			for _, c := range g.Exile.Cards {
				if c.Name == "Blue Spell" {
					spell = c.InstanceID
				}
			}
			_ = saga
			advanceToPrecombatMainOf(t, g, seat)
			passPriorityAroundTable(t, g)
			return oneOffOffer{g, me.ID, spell, exileFree, game.ZoneExile}
		}},
		{"Invasion of Karsus (a Siege)", game.MayCastKeywordFreeCast, func(t *testing.T) oneOffOffer {
			g := newCatalogGame(t)
			seat := g.Turn.ActiveSeat
			owner := g.Seats[(seat+1)%len(g.Seats)]
			protector := g.Seats[(seat+2)%len(g.Seats)]
			row := siegeRow(invasionOfKarsusOracleID, "Invasion of Karsus", "Refraction Elemental",
				"{2}{R}{R}", "4", "Creature — Elemental", "4", "4")
			id := importToHand(row, owner)
			if err := g.MoveCardByID(
				game.ZoneRef{Kind: game.ZoneHand, Owner: owner.ID},
				game.ZoneRef{Kind: game.ZoneBattlefield}, id,
			); err != nil {
				t.Fatalf("move battle to battlefield: %v", err)
			}
			answerProtectorPrompt(t, g, owner.ID, protector.ID)
			passPriorityAroundTable(t, g)
			advanceTo(t, g, game.StepBeginCombat)
			g.WithWriteLock(func() {
				if err := g.DealDamageToCreatureForEffect(uuid.Nil, id, 4); err != nil {
					t.Fatalf("damage the battle: %v", err)
				}
			})
			g.RunStateChecksForTest()
			passPriorityAroundTable(t, g)
			return oneOffOffer{g, owner.ID, id, exileFree, game.ZoneExile}
		}},
		{"Sunbird's Invocation", game.MayCastKeywordFreeCast, func(t *testing.T) oneOffOffer {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			pushCatalogPermanent(g, me.ID, "Sunbird's Invocation", "Enchantment", sunbirdsInvocationOracle, false)
			ids := sunbirdLibrary(me,
				game.Card{Name: "A Forest", TypeLine: "Basic Land — Forest"},
				game.Card{Name: "Cheap Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2},
			)
			castSpellWithCost(t, g, "Three Drop", "Sorcery", "", "{2}{R}", nil)
			passPriorityUntilChoice(t, g)
			return oneOffOffer{g, me.ID, ids[1], game.CastSpellParams{FromZone: "library", Strict: true}, game.ZoneLibrary}
		}},
		{"The Great Synthesis (chapter III)", game.MayCastKeywordHandCast, func(t *testing.T) oneOffOffer {
			g := newCatalogGame(t)
			seat := g.Turn.ActiveSeat
			me := g.Seats[seat]
			saga := jinCard(me.ID, true)
			saga.Counters = map[string]int{game.CounterLore: 2}
			pushBattlefieldCardWithTimestamp(g, saga)
			g.WithWriteLock(func() { me.Hand.Cards = nil })
			spell := handCardFull(me, "Pricey Sorcery", "Sorcery", "{5}{U}{U}", "", []string{"U"})
			advanceToPrecombatMainOf(t, g, seat)
			passPriorityAroundTable(t, g)
			return oneOffOffer{g, me.ID, spell, game.CastSpellParams{Strict: true}, game.ZoneHand}
		}},
	}
}

func TestOneOffCastOffersAreMadeDuringTheResolution(t *testing.T) {
	for _, tc := range oneOffCases() {
		t.Run(tc.name+"/cast", func(t *testing.T) {
			o := tc.setup(t)
			mustResolutionCastOffer(t, o.g, o.me, tc.keyword, o.card)
			if err := castThroughResolutionCast(t, o.g, o.me, o.card, o.params); err != nil {
				t.Fatalf("the cast through the offer: %v", err)
			}
			if !o.g.Stack.Contains(o.card) {
				t.Fatal("the card is not on the stack")
			}
			if resolutionCastOfferFor(o.g, o.me) != nil {
				declineResolutionCast(t, o.g, o.me)
			}
			castInsideTheResolution(t, o.g, o.card)
		})
		t.Run(tc.name+"/decline", func(t *testing.T) {
			o := tc.setup(t)
			mustResolutionCastOffer(t, o.g, o.me, tc.keyword, o.card)
			declineResolutionCast(t, o.g, o.me)
			if z := o.g.FindCardZoneForEffect(o.card); z == nil || z.Kind != o.stays {
				t.Fatalf("the declined card is in %v, want %s", z, o.stays)
			}
			if perm := o.g.CastPermissionOnCardByIDForEffect(o.card); perm != nil {
				t.Errorf("the decline left a permission: %+v", perm)
			}
			if o.g.ResolutionPaused() {
				t.Error("the decline did not let the resolution finish")
			}
		})
		t.Run(tc.name+"/restore", func(t *testing.T) {
			o := tc.setup(t)
			r := restoreRoundTrip(t, o.g, true)
			mustResolutionCastOffer(t, r, o.me, tc.keyword, o.card)
			if err := castThroughResolutionCast(t, r, o.me, o.card, o.params); err != nil {
				t.Fatalf("the cast on the restored table: %v", err)
			}
			if !r.Stack.Contains(o.card) {
				t.Fatal("the restored cast did not reach the stack")
			}
		})
	}
}
