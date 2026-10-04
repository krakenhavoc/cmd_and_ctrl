package game

import (
	"testing"

	"github.com/google/uuid"
)

// cost_commander_choice_test.go — #1397: a commander moved to PAY A
// COST is offered the command zone, and the payment never pauses to
// ask.
//
// Every cost component that can move a card is a row in one table,
// and every row runs the same checks, because the bug was a matrix:
// the discard, return and exile payers skipped the question
// (MustSettleNow) while the sacrifice and alternative-cost payers
// paused half way through the payment for it. One driver means a
// component added later is one row, not a fresh set of assertions to
// forget.
//
// ADR 0115 split the table by destination. A row whose cost puts the
// card into a HAND or a LIBRARY is CR 903.9b's, and #1397's
// ask-before-paying still holds for it:
//
//  1. a NON-commander pays without any prompt, as before;
//  2. a commander parks the announcement: exactly one CR 903.9b prompt,
//     addressed to the commander's OWNER, and NOTHING paid — the card
//     is where it was and the announcement has not been made;
//  3. "yes" makes the announcement and the commander lands in its
//     owner's command zone;
//  4. "no" makes the announcement and the commander lands where the
//     cost sends it;
//  5. undo across the open prompt replays: rewound into the prompt,
//     the other answer lands the other way and the announcement is
//     made once.
//
// A row whose cost puts the card into a GRAVEYARD or EXILE (discard,
// sacrifice, exile) is CR 903.9a's: the commander is paid like any
// other card and offered afterwards.
//
//  1. a NON-commander pays without any prompt, as before;
//  2. a commander pays at once: no prompt before or during the payment,
//     the announcement is made and the card is where the cost sends it;
//  3. the state-based action then asks the OWNER, and "yes" sends the
//     card from there to the command zone;
//  4. "no" leaves it where the cost sent it;
//  5. undo into the open question replays either answer.

// costCommanderSetup is one seeded board for a row.
type costCommanderSetup struct {
	// ownerSeat is the seat that owns the card — the chooser.
	ownerSeat int
	// card is the card the cost moves.
	card uuid.UUID
	// from is where the card sits before the payment; declined is
	// where the cost sends it when the owner says no.
	from, declined func(g *Game) *Zone
	// announce makes the announcement under test.
	announce func(g *Game) error
	// made reports whether the announcement landed: the spell or the
	// ability is on the stack, or the mana is in the pool.
	made func(g *Game) bool
}

type costCommanderCase struct {
	name  string
	build func(t *testing.T, g *Game, commander bool) costCommanderSetup
}

// seedCostCard puts a legendary creature card into z, owned by
// `owner` and controlled by `controller`, marked a commander when
// `commander` is set. `edit` adds the abilities a row needs.
func seedCostCard(z *Zone, owner, controller uuid.UUID, commander bool, edit func(*Card)) uuid.UUID {
	c := NewCard("Atraxa", owner)
	c.TypeLine = "Legendary Creature — Angel"
	c.Power, c.Toughness = 4, 4
	c.Controller = controller
	c.IsCommander = commander
	if edit != nil {
		edit(&c)
	}
	z.PushTop(c)
	return c.InstanceID
}

// creatureCostSpec ("a creature you control") is counter_cost_test.go's.

func handOf(seat int) func(*Game) *Zone { return func(g *Game) *Zone { return g.Seats[seat].Hand } }
func graveyardOf(seat int) func(*Game) *Zone {
	return func(g *Game) *Zone { return g.Seats[seat].Graveyard }
}
func commandOf(seat int) func(*Game) *Zone {
	return func(g *Game) *Zone { return g.Seats[seat].Command }
}
func battlefieldZone(g *Game) *Zone { return g.Battlefield }
func exileZone(g *Game) *Zone       { return g.Exile }

func oneAbilityOnStack(g *Game) bool { return len(g.StackMeta) == 1 }
func manaInPool(g *Game) bool        { return len(g.Seats[0].ManaPool) > 0 }
func spellOnStack(id uuid.UUID) func(*Game) bool {
	return func(g *Game) bool { return g.Stack.Contains(id) }
}

func noopEffect(*Game, *StackItem) error { return nil }

// abilitySource seats an artifact carrying one ability with `cost`.
func abilitySource(g *Game, owner *Player, cost AbilityCost) uuid.UUID {
	c := NewCard("Cost Outlet", owner.ID)
	c.TypeLine = "Artifact"
	c.Controller = owner.ID
	c.ActivatedAbilities = []ActivatedAbilityShape{{Label: "pay: nothing", Cost: cost, Effect: noopEffect}}
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

// manaSource seats an artifact carrying one mana ability.
func manaSource(g *Game, owner *Player, ab ManaAbilityShape) uuid.UUID {
	c := NewCard("Mana Outlet", owner.ID)
	c.TypeLine = "Artifact"
	c.Controller = owner.ID
	c.ManaAbilities = []ManaAbilityShape{ab}
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

func costCommanderCases() []costCommanderCase {
	return []costCommanderCase{
		// --- return to hand (#1213, ninjutsu's cost) -------------------
		{"ability: return your own commander to hand", func(t *testing.T, g *Game, cmd bool) costCommanderSetup {
			advanceTo(t, g, StepPrecombatMain)
			me := g.Seats[0]
			src := abilitySource(g, me, AbilityCost{ReturnToHand: &ReturnToHandCost{Count: 1, Filter: creatureCostSpec(), Label: "a creature you control"}})
			id := seedCostCard(g.Battlefield, me.ID, me.ID, cmd, nil)
			return costCommanderSetup{0, id, battlefieldZone, handOf(0),
				func(g *Game) error {
					return g.ActivateCatalogAbility(g.Seats[0].ID, src, 0, ActivateAbilityParams{ReturnIDs: []uuid.UUID{id}})
				}, oneAbilityOnStack}
		}},
		{"ability: return a STOLEN commander to its owner's hand", func(t *testing.T, g *Game, cmd bool) costCommanderSetup {
			advanceTo(t, g, StepPrecombatMain)
			me, opp := g.Seats[0], g.Seats[1]
			src := abilitySource(g, me, AbilityCost{ReturnToHand: &ReturnToHandCost{Count: 1, Filter: creatureCostSpec(), Label: "a creature you control"}})
			id := seedCostCard(g.Battlefield, opp.ID, me.ID, cmd, nil)
			return costCommanderSetup{1, id, battlefieldZone, handOf(1),
				func(g *Game) error {
					return g.ActivateCatalogAbility(g.Seats[0].ID, src, 0, ActivateAbilityParams{ReturnIDs: []uuid.UUID{id}})
				}, oneAbilityOnStack}
		}},

		// --- discard -----------------------------------------------------
		{"ability: discard a card", func(t *testing.T, g *Game, cmd bool) costCommanderSetup {
			advanceTo(t, g, StepPrecombatMain)
			me := g.Seats[0]
			src := pushDiscardOutlet(g, me, discardOutletAbility(1, "a card", nil))
			id := seedCostCard(me.Hand, me.ID, me.ID, cmd, nil)
			return costCommanderSetup{0, id, handOf(0), graveyardOf(0),
				func(g *Game) error {
					return g.ActivateCatalogAbility(g.Seats[0].ID, src, 0, ActivateAbilityParams{DiscardIDs: []uuid.UUID{id}})
				}, oneAbilityOnStack}
		}},
		{"ability: cycling (discard this card)", func(t *testing.T, g *Game, cmd bool) costCommanderSetup {
			advanceTo(t, g, StepPrecombatMain)
			me := g.Seats[0]
			id := seedCostCard(me.Hand, me.ID, me.ID, cmd, func(c *Card) {
				c.ActivatedAbilities = []ActivatedAbilityShape{cyclingAbility("{2}")}
			})
			me.ManaPool.AddMana(ManaToken{Color: "C"})
			me.ManaPool.AddMana(ManaToken{Color: "C"})
			return costCommanderSetup{0, id, handOf(0), graveyardOf(0),
				func(g *Game) error {
					return g.ActivateCatalogAbility(g.Seats[0].ID, id, 0, ActivateAbilityParams{})
				}, oneAbilityOnStack}
		}},
		{"spell: additional cost, discard a card", func(t *testing.T, g *Game, cmd bool) costCommanderSetup {
			const oracle = "test-1397-thrill"
			withCatalogAdditionalCost(t, func(o string) *AdditionalCost {
				if o == oracle {
					return &AdditionalCost{DiscardCards: 1, Label: "Discard a card"}
				}
				return nil
			})
			me := g.Seats[0]
			spell, _ := thrillInHand(t, g, me, oracle, 0)
			id := seedCostCard(me.Hand, me.ID, me.ID, cmd, nil)
			return costCommanderSetup{0, id, handOf(0), graveyardOf(0),
				func(g *Game) error {
					return g.CastSpell(g.Seats[0].ID, spell, CastSpellParams{DiscardIDs: []uuid.UUID{id}})
				}, spellOnStack(spell)}
		}},
		{"mana ability: discard a card", func(t *testing.T, g *Game, cmd bool) costCommanderSetup {
			me := g.Seats[0]
			src := pushDiscardManaSource(g, me, anyCardDiscardCost(), "{B}")
			id := seedCostCard(me.Hand, me.ID, me.ID, cmd, nil)
			return costCommanderSetup{0, id, handOf(0), graveyardOf(0),
				func(g *Game) error {
					return g.ActivateManaAbility(g.Seats[0].ID, src, 0, ManaAbilityParams{DiscardIDs: []uuid.UUID{id}})
				}, manaInPool}
		}},

		// --- exile from hand or graveyard --------------------------------
		{"ability: exile a card from your hand", func(t *testing.T, g *Game, cmd bool) costCommanderSetup {
			advanceTo(t, g, StepPrecombatMain)
			me := g.Seats[0]
			src := abilitySource(g, me, AbilityCost{ExileCards: &ExileCost{N: 1, Label: "a card"}})
			id := seedCostCard(me.Hand, me.ID, me.ID, cmd, nil)
			return costCommanderSetup{0, id, handOf(0), exileZone,
				func(g *Game) error {
					return g.ActivateCatalogAbility(g.Seats[0].ID, src, 0, ActivateAbilityParams{ExileIDs: []uuid.UUID{id}})
				}, oneAbilityOnStack}
		}},
		{"ability: exile a card from your graveyard", func(t *testing.T, g *Game, cmd bool) costCommanderSetup {
			advanceTo(t, g, StepPrecombatMain)
			me := g.Seats[0]
			src := abilitySource(g, me, AbilityCost{ExileCards: &ExileCost{N: 1, Label: "a card", From: ZoneGraveyard}})
			id := seedCostCard(me.Graveyard, me.ID, me.ID, cmd, nil)
			return costCommanderSetup{0, id, graveyardOf(0), exileZone,
				func(g *Game) error {
					return g.ActivateCatalogAbility(g.Seats[0].ID, src, 0, ActivateAbilityParams{ExileIDs: []uuid.UUID{id}})
				}, oneAbilityOnStack}
		}},
		{"ability: exile this card from your graveyard (scavenge)", func(t *testing.T, g *Game, cmd bool) costCommanderSetup {
			advanceTo(t, g, StepPrecombatMain)
			me := g.Seats[0]
			id := seedCostCard(me.Graveyard, me.ID, me.ID, cmd, func(c *Card) {
				c.ActivatedAbilities = []ActivatedAbilityShape{{
					Label:  "Scavenge",
					Cost:   AbilityCost{ExileSelf: true},
					Zones:  []ZoneKind{ZoneGraveyard},
					Effect: noopEffect,
				}}
			})
			return costCommanderSetup{0, id, graveyardOf(0), exileZone,
				func(g *Game) error {
					return g.ActivateCatalogAbility(g.Seats[0].ID, id, 0, ActivateAbilityParams{})
				}, oneAbilityOnStack}
		}},
		{"mana ability: exile this card from your hand (Spirit Guide)", func(t *testing.T, g *Game, cmd bool) costCommanderSetup {
			me := g.Seats[0]
			id := seedCostCard(me.Hand, me.ID, me.ID, cmd, func(c *Card) {
				c.ManaAbilities = spiritGuideAbility("{G}")
			})
			return costCommanderSetup{0, id, handOf(0), exileZone,
				func(g *Game) error {
					return g.ActivateManaAbility(g.Seats[0].ID, id, 0, ManaAbilityParams{})
				}, manaInPool}
		}},
		{"mana ability: exile a card from your hand (Cadaverous Bloom)", func(t *testing.T, g *Game, cmd bool) costCommanderSetup {
			me := g.Seats[0]
			src := manaSource(g, me, ManaAbilityShape{
				ExileCards: &ExileCost{N: 1, Label: "a card"},
				Produced:   "{B}{B}",
				Label:      "Exile a card from your hand: Add {B}{B}",
			})
			id := seedCostCard(me.Hand, me.ID, me.ID, cmd, nil)
			return costCommanderSetup{0, id, handOf(0), exileZone,
				func(g *Game) error {
					return g.ActivateManaAbility(g.Seats[0].ID, src, 0, ManaAbilityParams{ExileIDs: []uuid.UUID{id}})
				}, manaInPool}
		}},

		// --- exile a permanent you control (#1600) ----------------------
		{"ability: exile a creature you control (The Soul Stone)", func(t *testing.T, g *Game, cmd bool) costCommanderSetup {
			advanceTo(t, g, StepPrecombatMain)
			me := g.Seats[0]
			src := abilitySource(g, me, AbilityCost{ExilePermanents: creatureExileCost()})
			id := seedCostCard(g.Battlefield, me.ID, me.ID, cmd, nil)
			return costCommanderSetup{0, id, battlefieldZone, exileZone,
				func(g *Game) error {
					return g.ActivateCatalogAbility(g.Seats[0].ID, src, 0, ActivateAbilityParams{ExilePermanentIDs: []uuid.UUID{id}})
				}, oneAbilityOnStack}
		}},
		{"ability: exile a STOLEN commander you control", func(t *testing.T, g *Game, cmd bool) costCommanderSetup {
			advanceTo(t, g, StepPrecombatMain)
			me, opp := g.Seats[0], g.Seats[1]
			src := abilitySource(g, me, AbilityCost{ExilePermanents: creatureExileCost()})
			id := seedCostCard(g.Battlefield, opp.ID, me.ID, cmd, nil)
			return costCommanderSetup{1, id, battlefieldZone, exileZone,
				func(g *Game) error {
					return g.ActivateCatalogAbility(g.Seats[0].ID, src, 0, ActivateAbilityParams{ExilePermanentIDs: []uuid.UUID{id}})
				}, oneAbilityOnStack}
		}},
		{"mana ability: exile a creature you control (Food Chain)", func(t *testing.T, g *Game, cmd bool) costCommanderSetup {
			me := g.Seats[0]
			src := manaSource(g, me, ManaAbilityShape{
				ExilePermanents: creatureExileCost(),
				Produced:        "{G}{G}",
				Label:           "Exile a creature you control: Add {G}{G}",
			})
			id := seedCostCard(g.Battlefield, me.ID, me.ID, cmd, nil)
			return costCommanderSetup{0, id, battlefieldZone, exileZone,
				func(g *Game) error {
					return g.ActivateManaAbility(g.Seats[0].ID, src, 0, ManaAbilityParams{ExilePermanentIDs: []uuid.UUID{id}})
				}, manaInPool}
		}},

		// --- sacrifice (paused mid-payment before #1397) ------------------
		{"ability: sacrifice a STOLEN commander", func(t *testing.T, g *Game, cmd bool) costCommanderSetup {
			advanceTo(t, g, StepPrecombatMain)
			me, opp := g.Seats[0], g.Seats[1]
			src := abilitySource(g, me, AbilityCost{SacrificeOther: creatureCostSpec()})
			id := seedCostCard(g.Battlefield, opp.ID, me.ID, cmd, nil)
			return costCommanderSetup{1, id, battlefieldZone, graveyardOf(1),
				func(g *Game) error {
					return g.ActivateCatalogAbility(g.Seats[0].ID, src, 0, ActivateAbilityParams{SacrificeIDs: []uuid.UUID{id}})
				}, oneAbilityOnStack}
		}},
		{"ability: sacrifice this", func(t *testing.T, g *Game, cmd bool) costCommanderSetup {
			advanceTo(t, g, StepPrecombatMain)
			me := g.Seats[0]
			id := seedCostCard(g.Battlefield, me.ID, me.ID, cmd, func(c *Card) {
				c.ActivatedAbilities = []ActivatedAbilityShape{{
					Label: "Sacrifice this: nothing", Cost: AbilityCost{SacrificeSelf: true}, Effect: noopEffect,
				}}
			})
			return costCommanderSetup{0, id, battlefieldZone, graveyardOf(0),
				func(g *Game) error {
					return g.ActivateCatalogAbility(g.Seats[0].ID, id, 0, ActivateAbilityParams{})
				}, oneAbilityOnStack}
		}},
		{"spell: additional cost, sacrifice a creature", func(t *testing.T, g *Game, cmd bool) costCommanderSetup {
			const oracle = "test-1397-rites"
			withCatalogAdditionalCost(t, func(o string) *AdditionalCost {
				if o == oracle {
					return &AdditionalCost{Sacrifice: creatureCostSpec(), Label: "Sacrifice a creature"}
				}
				return nil
			})
			me := g.Seats[0]
			spell, _ := thrillInHand(t, g, me, oracle, 0)
			id := seedCostCard(g.Battlefield, me.ID, me.ID, cmd, nil)
			return costCommanderSetup{0, id, battlefieldZone, graveyardOf(0),
				func(g *Game) error {
					return g.CastSpell(g.Seats[0].ID, spell, CastSpellParams{SacrificeIDs: []uuid.UUID{id}})
				}, spellOnStack(spell)}
		}},
		{"mana ability: sacrifice a STOLEN commander (Ashnod's Altar)", func(t *testing.T, g *Game, cmd bool) costCommanderSetup {
			me, opp := g.Seats[0], g.Seats[1]
			src := manaSource(g, me, ManaAbilityShape{
				SacrificeOther: creatureCostSpec(),
				Produced:       "{C}{C}",
				Label:          "Sacrifice a creature: Add {C}{C}",
			})
			id := seedCostCard(g.Battlefield, opp.ID, me.ID, cmd, nil)
			return costCommanderSetup{1, id, battlefieldZone, graveyardOf(1),
				func(g *Game) error {
					return g.ActivateManaAbility(g.Seats[0].ID, src, 0, ManaAbilityParams{SacrificeIDs: []uuid.UUID{id}})
				}, manaInPool}
		}},

		// --- an alternative cost's card (paused mid-payment before #1397) --
		{"spell: alternative cost, exile a blue card from your hand", func(t *testing.T, g *Game, cmd bool) costCommanderSetup {
			const oracle = "test-1397-force"
			withCatalogAlternativeCosts(t, altCostFor(oracle, AlternativeCost{
				Key: "pitch", Label: "Exile a blue card", ExileFromHand: blueCardSpec(),
			}))
			advanceTo(t, g, StepPrecombatMain)
			me := g.Seats[0]
			spell := NewCard("Test Force", me.ID)
			spell.TypeLine = "Instant"
			spell.OracleID = oracle
			me.Hand.PushTop(spell)
			id := seedCostCard(me.Hand, me.ID, me.ID, cmd, func(c *Card) { c.Colors = []string{"U"} })
			return costCommanderSetup{0, id, handOf(0), exileZone,
				func(g *Game) error {
					return g.CastSpell(g.Seats[0].ID, spell.InstanceID, CastSpellParams{AlternativeCost: "pitch", AltCostIDs: []uuid.UUID{id}})
				}, spellOnStack(spell.InstanceID)}
		}},
		{"spell: alternative cost, return a permanent you control", func(t *testing.T, g *Game, cmd bool) costCommanderSetup {
			const oracle = "test-1397-daze"
			withCatalogAlternativeCosts(t, altCostFor(oracle, AlternativeCost{
				Key: "daze", Label: "Return a creature you control", ReturnToHand: creatureCostSpec(),
			}))
			advanceTo(t, g, StepPrecombatMain)
			me := g.Seats[0]
			spell := NewCard("Test Daze", me.ID)
			spell.TypeLine = "Instant"
			spell.OracleID = oracle
			me.Hand.PushTop(spell)
			id := seedCostCard(g.Battlefield, me.ID, me.ID, cmd, nil)
			return costCommanderSetup{0, id, battlefieldZone, handOf(0),
				func(g *Game) error {
					return g.CastSpell(g.Seats[0].ID, spell.InstanceID, CastSpellParams{AlternativeCost: "daze", AltCostIDs: []uuid.UUID{id}})
				}, spellOnStack(spell.InstanceID)}
		}},
		{"spell: flashback, sacrifice a creature (#1727)", func(t *testing.T, g *Game, cmd bool) costCommanderSetup {
			const oracle = "test-1727-flashback-sacrifice"
			me := g.Seats[0]
			withCatalogCastableZones(t, castableZonesFor(oracle, ZoneGraveyard))
			withCatalogAlternativeCosts(t, altCostFor(oracle, sacrificeFlashback(1)))
			spell := looterInGraveyard(t, g, me, oracle)
			id := seedCostCard(g.Battlefield, me.ID, me.ID, cmd, nil)
			return costCommanderSetup{0, id, battlefieldZone, graveyardOf(0),
				func(g *Game) error { return castSacrificeFlashback(g, spell, []uuid.UUID{id}) },
				spellOnStack(spell)}
		}},
		{"spell: escape, exile another card from your graveyard", func(t *testing.T, g *Game, cmd bool) costCommanderSetup {
			const oracle = "test-1397-escape"
			me := g.Seats[0]
			spell, _ := seedEscapeSpell(t, g, me, oracle, 1)
			id := seedCostCard(me.Graveyard, me.ID, me.ID, cmd, nil)
			return costCommanderSetup{0, id, graveyardOf(0), exileZone,
				func(g *Game) error { return castEscape(g, g.Seats[0], spell, []uuid.UUID{id}) },
				spellOnStack(spell)}
		}},
	}
}

// parkedCostPrompt asserts the table holds exactly one CR 903.9
// prompt about `card`, addressed to `owner`, parked on an announcement,
// and returns it.
func parkedCostPrompt(t *testing.T, g *Game, owner, card uuid.UUID) *PendingChoice {
	t.Helper()
	if len(g.PendingChoices) != 1 {
		t.Fatalf("pending choices = %d, want the one CR 903.9 prompt", len(g.PendingChoices))
	}
	c := g.PendingChoices[0]
	if c.Kind != PendingChoiceOptionalReplacement {
		t.Fatalf("prompt kind = %q, want %q", c.Kind, PendingChoiceOptionalReplacement)
	}
	if c.Chooser != owner {
		t.Fatalf("the prompt asks %s, want the commander's OWNER %s", c.Chooser, owner)
	}
	if c.costCommanderResume == nil || c.costCommanderResume.card != card {
		t.Fatal("the prompt is not parked on the announcement for this card")
	}
	if c.Source != card {
		t.Errorf("prompt Source = %s, want the commander", c.Source)
	}
	return c
}

// asksFirst reports whether a row's cost puts the card into a hand or
// a library — CR 903.9b, asked before paying — rather than a graveyard
// or exile, where CR 903.9a asks after (ADR 0115).
func asksFirst(g *Game, s costCommanderSetup) bool {
	k := s.declined(g).Kind
	return k == ZoneHand || k == ZoneLibrary
}

func TestCostCommanderChoice(t *testing.T) {
	for _, tc := range costCommanderCases() {
		t.Run(tc.name, func(t *testing.T) {
			{
				g := newActiveGame(t)
				if !asksFirst(g, tc.build(t, g, true)) {
					testCostCommanderPaidThenOffered(t, tc)
					return
				}
			}
			t.Run("a non-commander pays without asking", func(t *testing.T) {
				g := newActiveGame(t)
				s := tc.build(t, g, false)
				if err := s.announce(g); err != nil {
					t.Fatalf("announce: %v", err)
				}
				if len(g.PendingChoices) != 0 {
					t.Fatalf("a non-commander queued %d prompt(s)", len(g.PendingChoices))
				}
				if !s.made(g) {
					t.Error("the announcement was not made")
				}
				if !s.declined(g).Contains(s.card) {
					t.Errorf("the card is not in %s", s.declined(g).Kind)
				}
			})

			for _, apply := range []bool{true, false} {
				name := "no: the card goes where the cost sends it"
				if apply {
					name = "yes: the command zone, and the announcement is made"
				}
				t.Run(name, func(t *testing.T) {
					g := newActiveGame(t)
					s := tc.build(t, g, true)
					owner := g.Seats[s.ownerSeat].ID
					if err := s.announce(g); err != nil {
						t.Fatalf("announce: %v", err)
					}
					c := parkedCostPrompt(t, g, owner, s.card)
					// Nothing is paid while the owner decides.
					if !s.from(g).Contains(s.card) {
						t.Fatalf("the card left %s before its owner answered", s.from(g).Kind)
					}
					if s.made(g) {
						t.Fatal("the announcement was made before the owner answered")
					}
					if err := g.ResolveOptionalReplacement(c.ID, owner, apply); err != nil {
						t.Fatalf("answer: %v", err)
					}
					if len(g.PendingChoices) != 0 {
						t.Fatalf("%d prompt(s) left after the answer — the payment must not ask again", len(g.PendingChoices))
					}
					if !s.made(g) {
						t.Fatal("the announcement was not made after the answer")
					}
					want, other := s.declined(g), commandOf(s.ownerSeat)(g)
					if apply {
						want, other = other, want
					}
					assertOnlyIn(t, s.card, want, other, s.from(g))
				})
			}

			t.Run("undo across the open prompt replays either way", func(t *testing.T) {
				g := newActiveGame(t)
				s := tc.build(t, g, true)
				owner := g.Seats[s.ownerSeat].ID
				if err := s.announce(g); err != nil {
					t.Fatalf("announce: %v", err)
				}
				promptOpen := g.Clone()
				c := parkedCostPrompt(t, g, owner, s.card)
				if err := g.ResolveOptionalReplacement(c.ID, owner, true); err != nil {
					t.Fatalf("first answer: %v", err)
				}
				assertOnlyIn(t, s.card, commandOf(s.ownerSeat)(g), s.declined(g), s.from(g))

				g.WithWriteLock(func() { g.RestoreFrom(promptOpen) })
				c = parkedCostPrompt(t, g, owner, s.card)
				if !s.from(g).Contains(s.card) || s.made(g) {
					t.Fatal("the rewind did not put the table back on the open prompt with nothing paid")
				}
				if err := g.ResolveOptionalReplacement(c.ID, owner, false); err != nil {
					t.Fatalf("replayed answer: %v", err)
				}
				if !s.made(g) {
					t.Fatal("the replayed answer did not make the announcement")
				}
				assertOnlyIn(t, s.card, s.declined(g), commandOf(s.ownerSeat)(g), s.from(g))
			})
		})
	}
}

// testCostCommanderPaidThenOffered runs a graveyard-or-exile row
// (ADR 0115, CR 903.9a): the commander is paid like any other card,
// and its owner is asked afterwards.
func testCostCommanderPaidThenOffered(t *testing.T, tc costCommanderCase) {
	t.Run("a non-commander pays without asking", func(t *testing.T) {
		g := newActiveGame(t)
		s := tc.build(t, g, false)
		if err := s.announce(g); err != nil {
			t.Fatalf("announce: %v", err)
		}
		runChecks(g)
		if len(g.PendingChoices) != 0 {
			t.Fatalf("a non-commander queued %d prompt(s)", len(g.PendingChoices))
		}
		if !s.made(g) {
			t.Error("the announcement was not made")
		}
		if !s.declined(g).Contains(s.card) {
			t.Errorf("the card is not in %s", s.declined(g).Kind)
		}
	})

	// payThenAsk announces, checks the commander was paid at once, and
	// returns the CR 903.9a question that follows.
	payThenAsk := func(t *testing.T, g *Game, s costCommanderSetup) (*Player, *PendingChoice) {
		t.Helper()
		owner := g.Seats[s.ownerSeat]
		if err := s.announce(g); err != nil {
			t.Fatalf("announce: %v", err)
		}
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == PendingChoiceOptionalReplacement {
				t.Fatal("a graveyard or exile cost asked CR 903.9b before paying")
			}
		}
		if !s.made(g) {
			t.Fatal("the announcement was not made")
		}
		assertOnlyIn(t, s.card, s.declined(g), s.from(g), commandOf(s.ownerSeat)(g))
		return owner, expectCommanderReturn(t, g, owner, s.card)
	}

	for _, apply := range []bool{true, false} {
		name := "paid, then no: the card stays where the cost sent it"
		if apply {
			name = "paid, then yes: the command zone"
		}
		t.Run(name, func(t *testing.T) {
			g := newActiveGame(t)
			s := tc.build(t, g, true)
			owner, c := payThenAsk(t, g, s)
			if err := g.ResolveCommanderReturn(c.ID, owner.ID, apply); err != nil {
				t.Fatalf("answer: %v", err)
			}
			if len(g.PendingChoices) != 0 {
				t.Fatalf("%d prompt(s) left after the answer", len(g.PendingChoices))
			}
			if !s.made(g) {
				t.Fatal("answering the question undid the announcement")
			}
			want, other := s.declined(g), commandOf(s.ownerSeat)(g)
			if apply {
				want, other = other, want
			}
			assertOnlyIn(t, s.card, want, other, s.from(g))
		})
	}

	t.Run("undo into the open question replays either way", func(t *testing.T) {
		g := newActiveGame(t)
		s := tc.build(t, g, true)
		owner, c := payThenAsk(t, g, s)
		promptOpen := g.Clone()
		if err := g.ResolveCommanderReturn(c.ID, owner.ID, true); err != nil {
			t.Fatalf("first answer: %v", err)
		}
		assertOnlyIn(t, s.card, commandOf(s.ownerSeat)(g), s.declined(g), s.from(g))

		g.WithWriteLock(func() { g.RestoreFrom(promptOpen) })
		owner = g.Seats[s.ownerSeat]
		c = expectCommanderReturn(t, g, owner, s.card)
		if !s.declined(g).Contains(s.card) || !s.made(g) {
			t.Fatal("the rewind did not put the table back on the open question with the cost paid")
		}
		if err := g.ResolveCommanderReturn(c.ID, owner.ID, false); err != nil {
			t.Fatalf("replayed answer: %v", err)
		}
		assertOnlyIn(t, s.card, s.declined(g), commandOf(s.ownerSeat)(g), s.from(g))
	})
}

// The double spend the mid-payment pause allowed: an Ashnod's Altar
// activation naming a stolen commander paused on its owner's prompt
// with the commander still on the battlefield, so a second activation
// naming the SAME commander was accepted and paid for. #1397 asked
// first. Since ADR 0115 the sacrifice is paid at once: the commander is
// in its owner's graveyard after the first activation, the second one
// has nothing to sacrifice, and the owner is asked once, afterwards.
func TestACommanderSacrificedForACostCannotPayTwice(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	altar := manaSource(g, me, ManaAbilityShape{
		SacrificeOther: creatureCostSpec(),
		Produced:       "{C}{C}",
		Label:          "Sacrifice a creature: Add {C}{C}",
	})
	cmd := seedCostCard(g.Battlefield, opp.ID, me.ID, true, nil)
	sac := func() error {
		return g.ActivateManaAbility(me.ID, altar, 0, ManaAbilityParams{SacrificeIDs: []uuid.UUID{cmd}})
	}

	if err := sac(); err != nil {
		t.Fatalf("first activation: %v", err)
	}
	if got := len(me.ManaPool); got != 2 {
		t.Fatalf("mana pool = %d after the first sacrifice, want {C}{C}", got)
	}
	assertOnlyIn(t, cmd, opp.Graveyard, g.Battlefield, opp.Command)
	if err := sac(); err == nil {
		t.Fatal("a second activation sacrificed a commander that is already in the graveyard")
	}
	if got := len(me.ManaPool); got != 2 {
		t.Errorf("mana pool = %d, want the {C}{C} of ONE sacrifice", got)
	}
	answerCommanderReturn(t, g, opp, cmd, true)
	assertOnlyIn(t, cmd, opp.Command, opp.Graveyard, g.Battlefield)
}

// returnOutlet seats an artifact whose ability costs "return N
// creatures you control to their owners' hands".
func returnOutlet(g *Game, owner *Player, n int) uuid.UUID {
	return abilitySource(g, owner, AbilityCost{ReturnToHand: &ReturnToHandCost{Count: n, Filter: creatureCostSpec(), Label: "a creature you control"}})
}

// When the payer is the one answering, the refusal of a re-run that no
// longer validates comes back to them, because they can act on it. A
// return-to-hand cost since ADR 0115: it is the kind still asked first
// (CR 903.9b).
func TestAParkedCostTheBoardNoLongerPaysIsRefusedToThePayer(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	src := returnOutlet(g, me, 1)
	cmd := seedCostCard(g.Battlefield, me.ID, me.ID, true, nil)
	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{ReturnIDs: []uuid.UUID{cmd}}); err != nil {
		t.Fatalf("announce: %v", err)
	}
	c := parkedCostPrompt(t, g, me.ID, cmd)
	// The card leaves the battlefield while the question is open.
	g.WithWriteLock(func() {
		if _, err := MoveCard(g.Battlefield, me.Hand, cmd); err != nil {
			t.Fatal(err)
		}
	})
	if err := g.ResolveOptionalReplacement(c.ID, me.ID, true); err == nil {
		t.Fatal("answer = nil, want the refusal — the return can no longer be paid")
	}
	if len(g.StackMeta) != 0 {
		t.Error("an announcement whose cost could not be paid reached the stack")
	}
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == src && g.Battlefield.Cards[i].Tapped {
			t.Error("the refused activation tapped its source — something was paid")
		}
	}
}

// Two commanders in one payment are asked about one at a time, and the
// first answer is carried to the second prompt rather than asked again.
// Two returns to hand since ADR 0115 (CR 903.9b).
func TestTwoCommandersInOnePaymentAreAskedInTurn(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	src := returnOutlet(g, me, 2)
	a := seedCostCard(g.Battlefield, me.ID, me.ID, true, nil)
	b := seedCostCard(g.Battlefield, me.ID, me.ID, true, nil)
	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{ReturnIDs: []uuid.UUID{a, b}}); err != nil {
		t.Fatalf("announce: %v", err)
	}
	c := parkedCostPrompt(t, g, me.ID, a)
	if err := g.ResolveOptionalReplacement(c.ID, me.ID, true); err != nil {
		t.Fatalf("first answer: %v", err)
	}
	if len(g.StackMeta) != 0 || !g.Battlefield.Contains(a) {
		t.Fatal("the payment began with one commander still unanswered")
	}
	c = parkedCostPrompt(t, g, me.ID, b)
	if err := g.ResolveOptionalReplacement(c.ID, me.ID, false); err != nil {
		t.Fatalf("second answer: %v", err)
	}
	assertOnlyIn(t, a, me.Command, me.Hand, g.Battlefield)
	assertOnlyIn(t, b, me.Hand, me.Command, g.Battlefield)
	if len(g.StackMeta) != 1 {
		t.Errorf("StackMeta = %d, want the one ability", len(g.StackMeta))
	}
}

// An owner who has left the game cannot be asked (CR 800.4a), so the
// payment goes ahead as a decline rather than waiting on a prompt
// nobody can answer.
func TestACostCommanderWhoseOwnerHasLeftIsNotAsked(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	src := abilitySource(g, me, AbilityCost{SacrificeOther: creatureCostSpec()})
	cmd := seedCostCard(g.Battlefield, opp.ID, me.ID, true, nil)
	g.WithWriteLock(func() { opp.Eliminated = true })
	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{SacrificeIDs: []uuid.UUID{cmd}}); err != nil {
		t.Fatalf("announce: %v", err)
	}
	if len(g.PendingChoices) != 0 {
		t.Fatalf("%d prompt(s) queued for an owner who has left", len(g.PendingChoices))
	}
	if len(g.StackMeta) != 1 {
		t.Error("the ability was not activated")
	}
	if opp.Command.Contains(cmd) {
		t.Error("the commander of a departed owner was sent to the command zone unasked")
	}
}

// The answer rides the route, not the game: a returned commander whose
// owner said yes still meets another replacement on its way out (a
// test "if a card would be put into a hand, exile it instead") in the
// ordinary CR 616 way, and the command zone still wins because the
// owner already chose it. A return to hand since ADR 0115: a cost that
// goes to a graveyard or exile is no longer asked first.
func TestAnAcceptedCostAnswerStillMeetsOtherReplacements(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(ReplacementEffect{
			Watches: []EventKind{EventZoneMove},
			Label:   "Test hand-to-exile",
			AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
				return isExitMove(ev.Kind) && ev.NewZone == ZoneHand
			},
			Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
				ev.NewZone = ZoneExile
				return nil
			},
		})
	})
	cmd := seedCostCard(g.Battlefield, me.ID, me.ID, true, nil)
	src := returnOutlet(g, me, 1)
	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{ReturnIDs: []uuid.UUID{cmd}}); err != nil {
		t.Fatalf("announce: %v", err)
	}
	c := parkedCostPrompt(t, g, me.ID, cmd)
	if err := g.ResolveOptionalReplacement(c.ID, me.ID, true); err != nil {
		t.Fatalf("answer: %v", err)
	}
	if len(g.PendingChoices) != 0 {
		t.Fatalf("a settle-now cost paused on %d prompt(s)", len(g.PendingChoices))
	}
	assertOnlyIn(t, cmd, me.Command, g.Exile, g.Battlefield, me.Hand)
}

// ADR 0115 §5: a discarded commander meets Rest in Peace like any other
// card. It is exiled instead of going to the graveyard, with no CR 616
// ordering prompt (there is no second replacement to order any more),
// and CR 903.9a then offers it from exile.
func TestADiscardedCommanderMeetsRestInPeaceThenIsOffered(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(ReplacementEffect{
			Watches: []EventKind{EventZoneMove, EventDiscardCard},
			Label:   "Test Rest in Peace",
			AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
				return isExitMove(ev.Kind) && ev.NewZone == ZoneGraveyard
			},
			Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
				ev.NewZone = ZoneExile
				return nil
			},
		})
	})
	pitch := seedCostCard(me.Hand, me.ID, me.ID, true, nil)
	src := pushDiscardOutlet(g, me, discardOutletAbility(1, "a card", nil))
	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{DiscardIDs: []uuid.UUID{pitch}}); err != nil {
		t.Fatalf("announce: %v", err)
	}
	if len(g.StackMeta) != 1 {
		t.Fatal("the ability was not activated")
	}
	assertOnlyIn(t, pitch, g.Exile, me.Graveyard, me.Hand, me.Command)
	answerCommanderReturn(t, g, me, pitch, true)
	assertOnlyIn(t, pitch, me.Command, g.Exile, me.Graveyard, me.Hand)
}
