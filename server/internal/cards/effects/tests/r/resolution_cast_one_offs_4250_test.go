package testsr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardsa "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/a"
	cardsth "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/th"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			instant := uuid.New()
			me.Graveyard.PushTop(game.Card{InstanceID: instant, Name: "Old Instant", TypeLine: "Instant",
				ManaCost: "{3}{U}{U}", Owner: me.ID, Controller: me.ID})
			effectstest.CastCatalogSpell(t, g, "Torrential Gearhulk", "Artifact Creature — Construct", effectstest.YsAGearhulkOracle, nil)
			effectstest.PassPriorityAroundTable(t, g)
			effectstest.AnswerAnyPendingTargetPrompts(t, g)
			effectstest.PassPriorityAroundTable(t, g)
			return oneOffOffer{g, me.ID, instant, game.CastSpellParams{FromZone: "graveyard", Strict: true,
				AlternativeCost: game.AltCostKeyGrantedCost}, game.ZoneGraveyard}
		}},
		{"Arcane Heist", game.MayCastKeywordFreeCast, func(t *testing.T) oneOffOffer {
			g := effectstest.NewCatalogGame(t)
			me, opp := effectstest.CiaSeats(t, g)
			third := g.Seats[(g.Turn.ActiveSeat+2)%4]
			bolt := effectstest.GyCard(opp, "Lightning Bolt", "Instant", "{R}", effectstest.LightningBoltOracle)
			effectstest.CastCatalogSpell(t, g, "Arcane Heist", "Sorcery", effectstest.CiaArcaneHeistOracle,
				[]game.TargetRef{{Kind: game.TargetCard, ID: bolt}})
			effectstest.PassPriorityAroundTable(t, g)
			return oneOffOffer{g, me.ID, bolt, game.CastSpellParams{FromZone: "graveyard", Strict: true,
				AlternativeCost: game.AltCostKeyGrantedCost,
				Targets:         []game.TargetRef{{Kind: game.TargetPlayer, ID: third.ID}}}, game.ZoneGraveyard}
		}},
		{"Memory Plunder", game.MayCastKeywordFreeCast, func(t *testing.T) oneOffOffer {
			g := effectstest.NewCatalogGame(t)
			effectstest.AdvanceToMain(t, g)
			me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
			third := g.Seats[(g.Turn.ActiveSeat+2)%4]
			bolt := effectstest.GyCard(opp, "Lightning Bolt", "Instant", "{R}", effectstest.LightningBoltOracle)
			effectstest.CastCatalogSpell(t, g, "Memory Plunder", "Instant", effectstest.NghaMemoryPlunderOracle,
				[]game.TargetRef{{Kind: game.TargetCard, ID: bolt}})
			effectstest.PassPriorityAroundTable(t, g)
			return oneOffOffer{g, me.ID, bolt, game.CastSpellParams{FromZone: "graveyard", Strict: true,
				AlternativeCost: game.AltCostKeyGrantedCost,
				Targets:         []game.TargetRef{{Kind: game.TargetPlayer, ID: third.ID}}}, game.ZoneGraveyard}
		}},
		{"Emet-Selch of the Third Seat", game.MayCastKeywordPaidCast, func(t *testing.T) oneOffOffer {
			g, me, opp, emet := effectstest.EmetSetup(t)
			div := effectstest.GyCard(me, "Divination", "Sorcery", "{2}{U}", effectstest.DivinationOracle)
			effectstest.EmetOffer(t, g, me, opp, emet, div)
			effectstest.FloatForTest(g, me, "U")
			return oneOffOffer{g, me.ID, div, game.CastSpellParams{FromZone: "graveyard", Strict: true,
				AlternativeCost: game.AltCostKeyGrantedCost}, game.ZoneGraveyard}
		}},
		{"Strago and Relm", game.MayCastKeywordFreeCast, func(t *testing.T) oneOffOffer {
			g, me, _, _, _, bear := effectstest.StragoGame(t)
			return oneOffOffer{g, me.ID, bear, exileFree, game.ZoneExile}
		}},
		{"Crabomination", game.MayCastKeywordFreeCast, func(t *testing.T) oneOffOffer {
			g, me, opp := effectstest.EmergeTable(t)
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
			rock := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Rock", TypeLine: "Artifact", ManaCost: "{5}"})
			crab := effectstest.HandCardOf(g, me, "Crabomination", "Creature — Crab Demon", "{4}{B}{B}", effectstest.CrabominationOracle)
			if err := effectstest.CastEmerge(g, me, crab, rock, false); err != nil {
				t.Fatalf("emerge from artifact: %v", err)
			}
			for i := 0; i < 8 && effectstest.FindBattlefieldCardForTest(g, crab) == nil; i++ {
				if err := g.PassPriority(); err != nil {
					t.Fatalf("PassPriority: %v", err)
				}
			}
			if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
				if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}); err != nil {
					t.Fatalf("pick the opponent: %v", err)
				}
			}
			effectstest.PassPriorityAroundTable(t, g)
			return oneOffOffer{g, me.ID, hand.InstanceID, exileFree, game.ZoneExile}
		}},
		{"Amped Raptor", game.MayCastKeywordPaidCast, func(t *testing.T) oneOffOffer {
			g, me, _ := effectstest.P7Table(t)
			hit := uuid.New()
			g.WithWriteLock(func() {
				me.Library.PushTop(game.Card{InstanceID: hit, Name: "Two Drop", TypeLine: "Creature — Test", ManaCost: "{1}{G}",
					Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
			})
			effectstest.CastCatalogSpell(t, g, "Amped Raptor", "Creature — Dinosaur", effectstest.AmpedRaptorOracle, nil)
			effectstest.PassPriorityAroundTable(t, g)
			return oneOffOffer{g, me.ID, hit, game.CastSpellParams{FromZone: "exile", Strict: true,
				AlternativeCost: cardsa.AmpedRaptorAltCostKey}, game.ZoneExile}
		}},
		{"The Infamous Cruelclaw", game.MayCastKeywordPaidCast, func(t *testing.T) oneOffOffer {
			g, me, opp := effectstest.DiscardTable(t)
			hit := uuid.New()
			g.WithWriteLock(func() {
				me.Library.PushTop(game.Card{InstanceID: hit, Name: "Two Drop", TypeLine: "Creature — Test", ManaCost: "{1}{G}",
					Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
			})
			spare := effectstest.HandCardOf(g, me, "Spare", "Sorcery", "{2}", "")
			claw := effectstest.PushCatalogPermanent(g, me.ID, "The Infamous Cruelclaw", "Legendary Creature — Weasel Mercenary", effectstest.CruelclawOracle, false)
			effectstest.DealCombatDamageToPlayer(g, claw, opp.ID, 3)
			effectstest.PassPriorityAroundTable(t, g)
			return oneOffOffer{g, me.ID, hit, game.CastSpellParams{FromZone: "exile", Strict: true,
				AlternativeCost: cardsth.CruelclawAltCostKey, AltCostIDs: []uuid.UUID{spare}}, game.ZoneExile}
		}},
		{"Primeval Spawn", game.MayCastKeywordFreeCast, func(t *testing.T) oneOffOffer {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			me.Library.Cards = nil
			a := effectstest.PushLibraryCardForTest(me, game.Card{Name: "A", TypeLine: "Artifact", ManaCost: "{2}"})
			effectstest.SpawnLeaves(t, g, me)
			return oneOffOffer{g, me.ID, a, exileFree, game.ZoneExile}
		}},
		{"Fevered Suspicion", game.MayCastKeywordFreeCast, func(t *testing.T) oneOffOffer {
			g := effectstest.NewCatalogGame(t)
			var first uuid.UUID
			for i := 1; i < len(g.Seats); i++ {
				opp := g.Seats[(g.Turn.ActiveSeat+i)%len(g.Seats)]
				spell := effectstest.CardNamed("Stolen Sorcery", "Sorcery")
				spell.ManaCost = "{5}"
				effectstest.LibraryTop(opp, spell, effectstest.CardNamed("Exiled Land", "Land"))
				if first == uuid.Nil {
					first = spell.InstanceID
				}
			}
			me, _ := effectstest.CastReboundSpellAtMain(t, g, "Fevered Suspicion", "Sorcery", effectstest.FeveredSuspicionOracle)
			effectstest.PassPriorityAroundTable(t, g)
			return oneOffOffer{g, me.ID, first, exileFree, game.ZoneExile}
		}},
		{"Kefka, Dancing Mad", game.MayCastKeywordFreeCast, func(t *testing.T) oneOffOffer {
			g, _ := effectstest.KefkaGame(t)
			spell := effectstest.GraveyardSpell(g.Seats[1], "Their Sorcery", "{3}{U}")
			effectstest.AdvanceToEndStepOf(t, g, 0)
			effectstest.PassPriorityAroundTable(t, g)
			return oneOffOffer{g, g.Seats[0].ID, spell, exileFree, game.ZoneExile}
		}},
		{"Conduit of Worlds", game.MayCastKeywordPaidCast, func(t *testing.T) oneOffOffer {
			g, me, conduit, bear, _, _ := effectstest.ConduitTable(t)
			effectstest.ConduitMana(t, g, me, "{C}{G}{R}")
			if err := effectstest.ConduitActivate(t, g, me, conduit, bear); err != nil {
				t.Fatalf("activate: %v", err)
			}
			return oneOffOffer{g, me.ID, bear, game.CastSpellParams{FromZone: "graveyard", Strict: true}, game.ZoneGraveyard}
		}},
		{"Hazoret's Undying Fury", game.MayCastKeywordFreeCast, func(t *testing.T) oneOffOffer {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			five, _, _, _ := effectstest.FuryLibrary(me)
			effectstest.FuryCast(t, g)
			return oneOffOffer{g, me.ID, five, exileFree, game.ZoneExile}
		}},
		{"Etali, Primal Conqueror", game.MayCastKeywordFreeCast, func(t *testing.T) oneOffOffer {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			var mine uuid.UUID
			for _, p := range g.Seats {
				id := effectstest.SweepFTop(p, "Spell of "+p.Name, "Sorcery")
				if p.ID == me.ID {
					mine = id
				}
			}
			effectstest.ImportToBattlefield(t, g, effectstest.EtaliRow(), me)
			effectstest.PassPriorityAroundTable(t, g)
			return oneOffOffer{g, me.ID, mine, exileFree, game.ZoneExile}
		}},
		{"Goliath Daydreamer", game.MayCastKeywordFreeCast, func(t *testing.T) oneOffOffer {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			effectstest.AdvanceToMain(t, g)
			goliath := effectstest.PushCatalogPermanent(g, me.ID, "Goliath Daydreamer", "Creature — Giant Wizard", effectstest.GoliathOracle, false)
			dream := uuid.New()
			g.WithWriteLock(func() {
				g.Exile.PushTop(game.Card{InstanceID: dream, Name: "Divination", TypeLine: "Sorcery", ManaCost: "{2}{U}",
					OracleID: effectstest.SpellControlDivinationOracle, Owner: me.ID, Controller: me.ID,
					Counters: map[string]int{game.CounterDream: 1}})
			})
			effectstest.DeclareAttack(t, g, opp.ID, goliath)
			effectstest.PassPriorityAroundTable(t, g)
			return oneOffOffer{g, me.ID, dream, exileFree, game.ZoneExile}
		}},
		{"Power Pack", game.MayCastKeywordFreeCast, func(t *testing.T) oneOffOffer {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			effectstest.AdvanceToMain(t, g)
			pp := effectstest.B12Push(g, me.ID, "Power Pack", "Legendary Creature — Human Hero", effectstest.PowerPackOracle, 4, 4)
			div := effectstest.PpGraveyardCard(g, me, "Divination", "Sorcery", effectstest.SpellControlDivinationOracle)
			effectstest.DealCombatDamageToPlayer(g, pp, opp.ID, 4)
			effectstest.PassPriorityAroundTable(t, g)
			effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
			effectstest.PassPriorityAroundTable(t, g)
			return oneOffOffer{g, me.ID, div, game.CastSpellParams{FromZone: "exile", Strict: true,
				AlternativeCost: game.AltCostKeyGrantedCost}, game.ZoneExile}
		}},
		{"Summon: Esper Valigarmanda", game.MayCastKeywordPaidCast, func(t *testing.T) oneOffOffer {
			g := effectstest.NewCatalogGame(t)
			seat := g.Turn.ActiveSeat
			me := g.Seats[seat]
			effectstest.ValigarmandaGraveCard(me, "Blue Spell", "Instant", "{U}{U}")
			saga := effectstest.ValigarmandaCast(t, g)
			effectstest.PassPriorityAroundTable(t, g)
			var spell uuid.UUID
			for _, c := range g.Exile.Cards {
				if c.Name == "Blue Spell" {
					spell = c.InstanceID
				}
			}
			_ = saga
			effectstest.AdvanceToPrecombatMainOf(t, g, seat)
			effectstest.PassPriorityAroundTable(t, g)
			return oneOffOffer{g, me.ID, spell, exileFree, game.ZoneExile}
		}},
		{"Invasion of Karsus (a Siege)", game.MayCastKeywordFreeCast, func(t *testing.T) oneOffOffer {
			g := effectstest.NewCatalogGame(t)
			seat := g.Turn.ActiveSeat
			owner := g.Seats[(seat+1)%len(g.Seats)]
			protector := g.Seats[(seat+2)%len(g.Seats)]
			row := effectstest.SiegeRow(effects.InvasionOfKarsusOracleID, "Invasion of Karsus", "Refraction Elemental",
				"{2}{R}{R}", "4", "Creature — Elemental", "4", "4")
			id := effectstest.ImportToHand(row, owner)
			if err := g.MoveCardByID(
				game.ZoneRef{Kind: game.ZoneHand, Owner: owner.ID},
				game.ZoneRef{Kind: game.ZoneBattlefield}, id,
			); err != nil {
				t.Fatalf("move battle to battlefield: %v", err)
			}
			effectstest.AnswerProtectorPrompt(t, g, owner.ID, protector.ID)
			effectstest.PassPriorityAroundTable(t, g)
			effectstest.AdvanceTo(t, g, game.StepBeginCombat)
			g.WithWriteLock(func() {
				if err := g.DealDamageToCreatureForEffect(uuid.Nil, id, 4); err != nil {
					t.Fatalf("damage the battle: %v", err)
				}
			})
			g.RunStateChecksForTest()
			effectstest.PassPriorityAroundTable(t, g)
			return oneOffOffer{g, owner.ID, id, exileFree, game.ZoneExile}
		}},
		{"Sunbird's Invocation", game.MayCastKeywordFreeCast, func(t *testing.T) oneOffOffer {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			effectstest.PushCatalogPermanent(g, me.ID, "Sunbird's Invocation", "Enchantment", effectstest.SunbirdsInvocationOracle, false)
			ids := effectstest.SunbirdLibrary(me,
				game.Card{Name: "A Forest", TypeLine: "Basic Land — Forest"},
				game.Card{Name: "Cheap Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2},
			)
			effectstest.CastSpellWithCost(t, g, "Three Drop", "Sorcery", "", "{2}{R}", nil)
			effectstest.PassPriorityUntilChoice(t, g)
			return oneOffOffer{g, me.ID, ids[1], game.CastSpellParams{FromZone: "library", Strict: true}, game.ZoneLibrary}
		}},
		{"The Great Synthesis (chapter III)", game.MayCastKeywordHandCast, func(t *testing.T) oneOffOffer {
			g := effectstest.NewCatalogGame(t)
			seat := g.Turn.ActiveSeat
			me := g.Seats[seat]
			saga := effectstest.JinCard(me.ID, true)
			saga.Counters = map[string]int{game.CounterLore: 2}
			effectstest.PushBattlefieldCardWithTimestamp(g, saga)
			g.WithWriteLock(func() { me.Hand.Cards = nil })
			spell := effectstest.HandCardFull(me, "Pricey Sorcery", "Sorcery", "{5}{U}{U}", "", []string{"U"})
			effectstest.AdvanceToPrecombatMainOf(t, g, seat)
			effectstest.PassPriorityAroundTable(t, g)
			return oneOffOffer{g, me.ID, spell, game.CastSpellParams{Strict: true}, game.ZoneHand}
		}},
	}
}

func TestOneOffCastOffersAreMadeDuringTheResolution(t *testing.T) {
	for _, tc := range oneOffCases() {
		t.Run(tc.name+"/cast", func(t *testing.T) {
			o := tc.setup(t)
			effectstest.MustResolutionCastOffer(t, o.g, o.me, tc.keyword, o.card)
			if err := effectstest.CastThroughResolutionCast(t, o.g, o.me, o.card, o.params); err != nil {
				t.Fatalf("the cast through the offer: %v", err)
			}
			if !o.g.Stack.Contains(o.card) {
				t.Fatal("the card is not on the stack")
			}
			if effectstest.ResolutionCastOfferFor(o.g, o.me) != nil {
				effectstest.DeclineResolutionCast(t, o.g, o.me)
			}
			effectstest.CastInsideTheResolution(t, o.g, o.card)
		})
		t.Run(tc.name+"/decline", func(t *testing.T) {
			o := tc.setup(t)
			effectstest.MustResolutionCastOffer(t, o.g, o.me, tc.keyword, o.card)
			effectstest.DeclineResolutionCast(t, o.g, o.me)
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
			r := effectstest.RestoreRoundTrip(t, o.g, true)
			effectstest.MustResolutionCastOffer(t, r, o.me, tc.keyword, o.card)
			if err := effectstest.CastThroughResolutionCast(t, r, o.me, o.card, o.params); err != nil {
				t.Fatalf("the cast on the restored table: %v", err)
			}
			if !r.Stack.Contains(o.card) {
				t.Fatal("the restored cast did not reach the stack")
			}
		})
	}
}
