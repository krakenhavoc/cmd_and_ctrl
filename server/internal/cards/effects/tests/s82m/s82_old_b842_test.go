package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	b842StenchOracle     = "14062abe-14df-4dbe-ad7b-530e5f6c4988"
	b842ScepterOracle    = "34a85d7f-d4ea-4a0f-aa4c-bf0b0f4987bf"
	b842FrostbeastOracle = "39050823-ccda-43c0-818d-d5db15d82b99"
	b842AuraOracle       = "465843dc-57d0-46fd-ac47-238723034563"
	b842JokulOracle      = "7bdfe368-353c-4f8d-851f-05b7ea9f905d"
	b842TotalWarOracle   = "8266dd1c-e262-4467-a893-6b3b5eedb427"
	b842MinionOracle     = "8af3b981-6eda-46b0-a7b8-226bc719d137"
	b842IcequakeOracle   = "8c0a6ac1-b1cc-405a-95fb-f06b7fecbb54"
	b842FumaroleOracle   = "a9e33138-e1c0-4b69-b868-1ce8493c073d"
	b842AnarchyOracle    = "b91ed762-9232-4b77-abf4-97f66637f49c"
	b842FilterOracle     = "b9f63388-d1ac-4895-84c5-e5b1db187463"
	b842VortexOracle     = "ea38a211-a7b0-49fe-b5dc-7eca7610cccb"
	b842ThermoOracle     = "f3b071a1-7501-4424-979a-780fa90eecb8"
)

func b842Cards(ids ...uuid.UUID) []game.TargetRef {
	var out []game.TargetRef
	for _, id := range ids {
		out = append(out, game.TargetRef{Kind: game.TargetCard, ID: id})
	}
	return out
}

// Stench of Evil destroys every Plains and asks each destroyed land's controller to pay {2}; a decline is 1 damage.
func TestB842StenchOfEvilTaxesEachPlainsController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, a, b := g.Seats[0], g.Seats[1], g.Seats[2]
	mine := effectstest.PushLandFor(g, me.ID, "Plains", "Basic Land — Plains")
	pa := effectstest.PushLandFor(g, a.ID, "Plains", "Basic Land — Plains")
	pb := effectstest.PushLandFor(g, b.ID, "Plains", "Basic Land — Plains")
	island := effectstest.PushLandFor(g, a.ID, "Island", "Basic Land — Island")
	lives := map[uuid.UUID]int{me.ID: me.Life, a.ID: a.Life, b.ID: b.Life}
	effectstest.CastCatalogSpell(t, g, "Stench of Evil", "Sorcery", b842StenchOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{mine, pa, pb} {
		if g.Battlefield.Contains(id) {
			t.Error("a Plains survived")
		}
	}
	if !g.Battlefield.Contains(island) {
		t.Error("an Island was destroyed")
	}
	effectstest.ApaMana(a, "C", "C")
	effectstest.AnswerPayUnless(t, g, a.ID, true)
	effectstest.AnswerPayUnless(t, g, b.ID, false)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if a.Life != lives[a.ID] || b.Life != lives[b.ID]-1 || me.Life != lives[me.ID]-1 {
		t.Errorf("lives me %d a %d b %d, want -1 for me and b, none for a (who paid)", me.Life-lives[me.ID], a.Life-lives[a.ID], b.Life-lives[b.ID])
	}
}

// Despotic Scepter: only a permanent you OWN, and the destruction can't be regenerated.
func TestB842DespoticScepterDestroysOnlyAPermanentYouOwn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	scepter := effectstest.PushCatalogPermanent(g, me.ID, "Despotic Scepter", "Artifact", b842ScepterOracle, true)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, scepter, 0, game.ActivateAbilityParams{Strict: true, Targets: b842Cards(theirs)}); err == nil {
		t.Fatal("targeted a permanent the controller does not own")
	}
	if err := g.ActivateCatalogAbility(me.ID, scepter, 0, game.ActivateAbilityParams{Strict: true, Targets: b842Cards(mine)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) {
		t.Error("the owned creature was not destroyed")
	}
}

// Kjeldoran Frostbeast: at end of combat the creature that blocked it is destroyed.
func TestB842KjeldoranFrostbeastDestroysItsBlockerAtEndOfCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	beast := effectstest.B12Push(g, me.ID, "Kjeldoran Frostbeast", "Creature — Elemental Beast", b842FrostbeastOracle, 2, 4)
	wall := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 9)
	bystander := effectstest.B12Creature(g, opp.ID, "Bystander", "Creature — Bear", 2, 2)
	effectstest.Pr7aBlockedCombat(t, g, opp.ID, beast, wall)
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(wall) {
		t.Error("the blocker survived the end of combat")
	}
	if !g.Battlefield.Contains(bystander) || !g.Battlefield.Contains(beast) {
		t.Error("a creature not in the combat, or the Frostbeast, was destroyed")
	}
}

// Arenson's Aura: sacrifice an enchantment to destroy an enchantment.
func TestB842ArensonsAuraSacrificesAnEnchantmentToDestroyOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	aura := effectstest.PushCatalogPermanent(g, me.ID, "Arenson's Aura", "Enchantment", b842AuraOracle, false)
	theirs := effectstest.B12Push(g, opp.ID, "Their Enchantment", "Enchantment", "", 0, 0)
	if err := g.ActivateCatalogAbility(me.ID, aura, 0, game.ActivateAbilityParams{Strict: true, Targets: b842Cards(theirs), SacrificeIDs: []uuid.UUID{aura}}); err == nil {
		t.Fatal("activated without {W}")
	}
	effectstest.FloatForTest(g, me, "W")
	if err := g.ActivateCatalogAbility(me.ID, aura, 0, game.ActivateAbilityParams{Strict: true, Targets: b842Cards(theirs), SacrificeIDs: []uuid.UUID{aura}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) || g.Battlefield.Contains(aura) {
		t.Error("the target and the sacrificed Aura should both be gone")
	}
}

// Jokulhaups destroys artifacts, creatures and lands, and nothing else.
func TestB842JokulhaupsDestroysArtifactsCreaturesAndLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	bear := effectstest.B12Creature(g, g.Seats[1].ID, "Bear", "Creature — Bear", 2, 2)
	land := effectstest.PushLandFor(g, g.Seats[2].ID, "Island", "Basic Land — Island")
	art := effectstest.B12Push(g, g.Seats[0].ID, "Rock", "Artifact", "", 0, 0)
	ench := effectstest.B12Push(g, g.Seats[0].ID, "Aura", "Enchantment", "", 0, 0)
	effectstest.CastCatalogSpell(t, g, "Jokulhaups", "Sorcery", b842JokulOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{bear, land, art} {
		if g.Battlefield.Contains(id) {
			t.Error("a creature, artifact or land survived")
		}
	}
	if !g.Battlefield.Contains(ench) {
		t.Error("an enchantment was destroyed")
	}
}

// Total War: the attacking player's untapped, non-Wall creatures that did not attack and have been theirs since the turn began die.
func TestB842TotalWarDestroysTheAttackersIdleCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.PushCatalogPermanent(g, opp.ID, "Total War", "Enchantment", b842TotalWarOracle, false)
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	idle := effectstest.B12Creature(g, me.ID, "Idle", "Creature — Bear", 2, 2)
	wall := effectstest.B12Creature(g, me.ID, "Wall", "Creature — Wall", 0, 4)
	fresh := effectstest.PushCatalogPermanent(g, me.ID, "Fresh", "Creature — Bear", "", true)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(idle) {
		t.Error("an untapped creature that did not attack survived")
	}
	for name, id := range map[string]uuid.UUID{"the attacker": attacker, "a Wall": wall, "a creature that arrived this turn": fresh, "an opponent's creature": theirs} {
		if !g.Battlefield.Contains(id) {
			t.Errorf("%s was destroyed", name)
		}
	}
}

// Minion of Leshrac: at its controller's upkeep, sacrificing another creature avoids the 5 damage; otherwise it deals 5 and taps.
func TestB842MinionOfLeshracUpkeep(t *testing.T) {
	for _, sacrifice := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[1]
		minion := effectstest.B12Push(g, me.ID, "Minion of Leshrac", "Creature — Demon Minion", b842MinionOracle, 5, 5)
		fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Rat", 1, 1)
		life := me.Life
		effectstest.AdvanceTo(t, g, game.StepUpkeep)
		effectstest.PassPriorityAroundTable(t, g)
		prompt := effectstest.PayUnlessFor(g, me.ID)
		if prompt == nil {
			t.Fatal("no upkeep prompt")
		}
		if sacrifice {
			if err := g.ResolvePayUnlessWithCards(prompt.ID, me.ID, true, []uuid.UUID{fodder}); err != nil {
				t.Fatalf("sacrifice: %v", err)
			}
			if me.Life != life || g.Battlefield.Contains(fodder) || effectstest.B13Tapped(t, g, minion) {
				t.Errorf("sacrificed: life %d->%d, fodder alive %v", life, me.Life, g.Battlefield.Contains(fodder))
			}
			continue
		}
		effectstest.AnswerPayUnless(t, g, me.ID, false)
		if me.Life != life-5 || !effectstest.B13Tapped(t, g, minion) || !g.Battlefield.Contains(fodder) {
			t.Errorf("declined: life %d->%d, minion tapped %v", life, me.Life, effectstest.B13Tapped(t, g, minion))
		}
	}
}

// Icequake and Thermokarst: the rider happens only when the destroyed land was snow.
func TestB842IcequakeAndThermokarstRidersNeedASnowLand(t *testing.T) {
	for _, snow := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		typeLine := "Basic Land — Swamp"
		if snow {
			typeLine = "Basic Snow Land — Swamp"
		}
		land := effectstest.PushLandFor(g, opp.ID, "Swamp", typeLine)
		oppLife, myLife := opp.Life, me.Life
		effectstest.CastCatalogSpell(t, g, "Icequake", "Sorcery", b842IcequakeOracle, b842Cards(land))
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(land) {
			t.Error("Icequake left the land")
		}
		if want := oppLife; snow {
			want--
			if opp.Life != want {
				t.Errorf("snow Icequake: opponent life %d, want %d", opp.Life, want)
			}
		} else if opp.Life != want {
			t.Errorf("plain Icequake: opponent life %d, want %d", opp.Life, want)
		}
		land2 := effectstest.PushLandFor(g, opp.ID, "Swamp", typeLine)
		effectstest.CastCatalogSpell(t, g, "Thermokarst", "Sorcery", b842ThermoOracle, b842Cards(land2))
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(land2) {
			t.Error("Thermokarst left the land")
		}
		want := myLife
		if snow {
			want++
		}
		if me.Life != want {
			t.Errorf("snow=%v Thermokarst: life %d, want %d", snow, me.Life, want)
		}
	}
}

// Fumarole: pay 3 life, destroy the creature and the land.
func TestB842FumaroleCostsThreeLifeAndDestroysBoth(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	land := effectstest.PushLandFor(g, opp.ID, "Island", "Basic Land — Island")
	other := effectstest.PushLandFor(g, opp.ID, "Swamp", "Basic Land — Swamp")
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Fumarole", "Sorcery", b842FumaroleOracle, b842Cards(bear, land))
	if me.Life != life-3 {
		t.Errorf("life %d -> %d, want -3 as the cost", life, me.Life)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || g.Battlefield.Contains(land) || !g.Battlefield.Contains(other) {
		t.Error("Fumarole should destroy exactly its two targets")
	}
}

// Anarchy destroys white permanents only.
func TestB842AnarchyDestroysWhitePermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	white := effectstest.ApaPush(g, g.Seats[1].ID, g.Seats[1].ID, game.Card{Name: "White Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Colors: []string{"W"}})
	whiteEnch := effectstest.ApaPush(g, g.Seats[2].ID, g.Seats[2].ID, game.Card{Name: "White Ench", TypeLine: "Enchantment", Colors: []string{"W"}})
	green := effectstest.ApaPush(g, g.Seats[1].ID, g.Seats[1].ID, game.Card{Name: "Green Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Colors: []string{"G"}})
	plains := effectstest.PushLandFor(g, g.Seats[1].ID, "Plains", "Basic Land — Plains")
	effectstest.CastCatalogSpell(t, g, "Anarchy", "Sorcery", b842AnarchyOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(white) || g.Battlefield.Contains(whiteEnch) {
		t.Error("a white permanent survived")
	}
	if !g.Battlefield.Contains(green) || !g.Battlefield.Contains(plains) {
		t.Error("a nonwhite permanent was destroyed")
	}
}

// Essence Filter: the caster chooses all enchantments or all nonwhite enchantments.
func TestB842EssenceFilterChoosesBetweenTheTwoSweeps(t *testing.T) {
	for mode, whiteSurvives := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		white := effectstest.ApaPush(g, g.Seats[1].ID, g.Seats[1].ID, game.Card{Name: "White Ench", TypeLine: "Enchantment", Colors: []string{"W"}})
		blue := effectstest.ApaPush(g, g.Seats[2].ID, g.Seats[2].ID, game.Card{Name: "Blue Ench", TypeLine: "Enchantment", Colors: []string{"U"}})
		creature := effectstest.B12Creature(g, g.Seats[1].ID, "Bear", "Creature — Bear", 2, 2)
		effectstest.CastCatalogSpellWithModes(t, g, "Essence Filter", "Sorcery", b842FilterOracle, []int{mode})
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(blue) {
			t.Errorf("mode %d: a nonwhite enchantment survived", mode)
		}
		if g.Battlefield.Contains(white) == !whiteSurvives {
			t.Errorf("mode %d: white enchantment survival = %v, want %v", mode, g.Battlefield.Contains(white), whiteSurvives)
		}
		if !g.Battlefield.Contains(creature) {
			t.Errorf("mode %d: a creature was destroyed", mode)
		}
	}
}

// Essence Vortex: the creature's controller may pay life equal to its toughness; otherwise it is destroyed.
func TestB842EssenceVortexDestroysUnlessItsControllerPaysLife(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		opp := g.Seats[1]
		bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 3)
		life := opp.Life
		effectstest.CastCatalogSpell(t, g, "Essence Vortex", "Instant", b842VortexOracle, b842Cards(bear))
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerPayUnless(t, g, opp.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(bear) != pay {
			t.Errorf("pay=%v: creature on battlefield %v", pay, g.Battlefield.Contains(bear))
		}
		if want := life; pay {
			want -= 3
			if opp.Life != want {
				t.Errorf("paid: life %d, want %d", opp.Life, want)
			}
		} else if opp.Life != want {
			t.Errorf("declined: life %d, want %d", opp.Life, want)
		}
	}
}
