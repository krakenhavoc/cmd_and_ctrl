package testsds

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestEvasionWaveCardsPublishTheirPrintedKeywords(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, typeLine, keyword string
	}{
		{"Bladetusk Boar", "f640a102-7ee8-4605-9e15-09938fb229b2", "Creature — Boar", "intimidate"},
		{"Looter il-Kor", "c87e3b1d-2d24-4f1c-84e8-1c21a347cdc2", "Creature — Kor Rogue", "shadow"},
		{"Lu Xun, Scholar General", "ea658352-abef-4201-b20c-f5c5809d1d3e", "Legendary Creature — Human Soldier", "horsemanship"},
		{"Furtive Homunculus", "6b59ffa2-5c8b-4549-aa09-2f2a5cabc5af", "Creature — Homunculus", "skulk"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			id := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
				InstanceID: uuid.New(), Name: tc.name, OracleID: tc.oracle,
				TypeLine: tc.typeLine, Power: 2, Toughness: 2,
				Owner: me.ID, Controller: me.ID,
			})
			if !effectstest.HasEffectiveKeyword(t, g, id, tc.keyword) {
				t.Fatalf("%s did not publish printed keyword %q", tc.name, tc.keyword)
			}
		})
	}
}

func TestEvasionWaveCardsChangeBlockLegality(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, typeLine, keyword string
		attackerPower                   int
		makeBlocker                     func(*game.Game, uuid.UUID) uuid.UUID
		makeIllegalBlocker              func(*game.Game, uuid.UUID) uuid.UUID
	}{
		{
			name: "Bladetusk Boar", oracle: "f640a102-7ee8-4605-9e15-09938fb229b2",
			typeLine: "Creature — Boar", keyword: "intimidate", attackerPower: 3,
			makeBlocker: func(g *game.Game, owner uuid.UUID) uuid.UUID {
				return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
					InstanceID: uuid.New(), Name: "Artifact blocker", TypeLine: "Artifact Creature — Golem",
					Power: 2, Toughness: 2, Owner: owner, Controller: owner,
				})
			},
			makeIllegalBlocker: func(g *game.Game, owner uuid.UUID) uuid.UUID {
				return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
					InstanceID: uuid.New(), Name: "Green blocker", TypeLine: "Creature — Elf",
					Power: 2, Toughness: 2, Colors: []string{"G"}, Owner: owner, Controller: owner,
				})
			},
		},
		{
			name: "Looter il-Kor", oracle: "c87e3b1d-2d24-4f1c-84e8-1c21a347cdc2",
			typeLine: "Creature — Kor Rogue", keyword: "shadow", attackerPower: 1,
			makeBlocker: func(g *game.Game, owner uuid.UUID) uuid.UUID {
				return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
					InstanceID: uuid.New(), Name: "Shadow blocker", TypeLine: "Creature — Rogue",
					Power: 2, Toughness: 2, Keywords: []string{"shadow"}, Owner: owner, Controller: owner,
				})
			},
			makeIllegalBlocker: func(g *game.Game, owner uuid.UUID) uuid.UUID {
				return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
					InstanceID: uuid.New(), Name: "Plain blocker", TypeLine: "Creature — Bear",
					Power: 2, Toughness: 2, Owner: owner, Controller: owner,
				})
			},
		},
		{
			name: "Lu Xun, Scholar General", oracle: "ea658352-abef-4201-b20c-f5c5809d1d3e",
			typeLine: "Legendary Creature — Human Soldier", keyword: "horsemanship", attackerPower: 1,
			makeBlocker: func(g *game.Game, owner uuid.UUID) uuid.UUID {
				return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
					InstanceID: uuid.New(), Name: "Horse blocker", TypeLine: "Creature — Horse",
					Power: 2, Toughness: 2, Keywords: []string{"horsemanship"}, Owner: owner, Controller: owner,
				})
			},
			makeIllegalBlocker: func(g *game.Game, owner uuid.UUID) uuid.UUID {
				return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
					InstanceID: uuid.New(), Name: "Plain blocker", TypeLine: "Creature — Bear",
					Power: 2, Toughness: 2, Owner: owner, Controller: owner,
				})
			},
		},
		{
			name: "Furtive Homunculus", oracle: "6b59ffa2-5c8b-4549-aa09-2f2a5cabc5af",
			typeLine: "Creature — Homunculus", keyword: "skulk", attackerPower: 2,
			makeBlocker: func(g *game.Game, owner uuid.UUID) uuid.UUID {
				return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
					InstanceID: uuid.New(), Name: "Small blocker", TypeLine: "Creature — Bear",
					Power: 1, Toughness: 2, Owner: owner, Controller: owner,
				})
			},
			makeIllegalBlocker: func(g *game.Game, owner uuid.UUID) uuid.UUID {
				return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
					InstanceID: uuid.New(), Name: "Large blocker", TypeLine: "Creature — Bear",
					Power: 3, Toughness: 3, Owner: owner, Controller: owner,
				})
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, them := effectstest.SeatsForCombat(g)
			attacker := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
				InstanceID: uuid.New(), Name: tc.name, OracleID: tc.oracle,
				TypeLine: tc.typeLine, Power: tc.attackerPower, Toughness: 3,
				Colors: []string{"R"}, Owner: me.ID, Controller: me.ID,
			})
			if !effectstest.HasEffectiveKeyword(t, g, attacker, tc.keyword) {
				t.Fatalf("%s did not publish %q", tc.name, tc.keyword)
			}
			blocker := tc.makeBlocker(g, them.ID)
			effectstest.AttackIntoBlocks(t, g, them.ID, attacker)
			if got := effectstest.BlockRefusal(t, g, blocker, attacker); got != "" {
				t.Errorf("legal evasion exception was refused: %q", got)
			}
			if illegal := tc.makeIllegalBlocker(g, them.ID); effectstest.BlockRefusal(t, g, illegal, attacker) == "" {
				t.Error("ordinary blocker was accepted despite the evasion keyword")
			}
		})
	}
}

func TestShizoGrantsFearToATargetLegendaryCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	shizo := effectstest.PushPermanentForTest(g, me.ID, "Shizo, Death's Storehouse", "008f2698-1721-45a3-8353-10f2f400dc8f", "Legendary Land")
	legend := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Legend", TypeLine: "Legendary Creature — Test",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	effectstest.AdvanceToMain(t, g)
	me.ManaPool.AddMana(game.ManaToken{Color: "B"})
	ordinary := effectstest.PushCreatureToBattlefieldForTest(g, me.ID, "Not legendary")
	if err := g.ActivateCatalogAbility(me.ID, shizo, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: ordinary}},
	}); !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("nonlegendary target accepted: %v", err)
	}
	if c, _ := effectstest.BattlefieldCard(g, shizo); c.Tapped {
		t.Fatal("rejected target paid Shizo's tap cost")
	}
	if err := g.ActivateCatalogAbility(me.ID, shizo, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: legend}},
	}); err != nil {
		t.Fatalf("activate Shizo: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, legend, "fear") {
		t.Error("Shizo did not grant fear until end of turn")
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	if effectstest.HasEffectiveKeyword(t, g, legend, "fear") {
		t.Error("Shizo's fear grant survived cleanup")
	}
}

func TestLooterAndLuXunResolveTheirDamageTriggers(t *testing.T) {
	t.Run("looter draw then discard", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, them := effectstest.SeatsForCombat(g)
		looter := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: "Looter il-Kor", OracleID: "c87e3b1d-2d24-4f1c-84e8-1c21a347cdc2",
			TypeLine: "Creature — Kor Rogue", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
		})
		handBefore := me.Hand.Size()
		effectstest.AttackWith(t, g, them.ID, looter)
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.DiscardOwed(g, me.ID) != 1 {
			t.Fatal("Looter did not queue one discard after drawing")
		}
		effectstest.DiscardFromHand(t, g, me.ID)
		effectstest.PassPriorityAroundTable(t, g)
		if me.Hand.Size() != handBefore || them.Life != 39 {
			t.Errorf("Looter result: hand=%d (started %d), opponent life=%d", me.Hand.Size(), handBefore, them.Life)
		}
	})

	t.Run("Lu Xun optional draw", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, them := effectstest.SeatsForCombat(g)
		luXun := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: "Lu Xun, Scholar General", OracleID: "ea658352-abef-4201-b20c-f5c5809d1d3e",
			TypeLine: "Legendary Creature — Human Soldier", Power: 1, Toughness: 3, Owner: me.ID, Controller: me.ID,
		})
		handBefore := me.Hand.Size()
		g.WithWriteLock(func() {
			if err := g.DealDamageToPlayerForEffect(luXun, them.ID, 1); err != nil {
				t.Error(err)
			}
		})
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
		effectstest.PassPriorityAroundTable(t, g)
		if me.Hand.Size() != handBefore+1 {
			t.Errorf("Lu Xun hand=%d, want %d", me.Hand.Size(), handBefore+1)
		}
	})
}
