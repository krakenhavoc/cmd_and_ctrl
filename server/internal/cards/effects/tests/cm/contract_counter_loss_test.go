package testscm

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #6123: Miss Highwater's "… they draw seven cards and get a contract
// counter. For as long as they have a contract counter, when they lose the
// game, for each artifact and creature they controlled, create a token
// that's a copy of it." A delayed trigger (CR 603.7a) on the player, read
// as they leave, resolving after CR 800.4a has taken their objects away.

const contractOracle = "test-6123-contract"

// registerContract is the clause on an activated ability: "{0}: Target
// player gets a contract counter. For as long as they have a contract
// counter, when they lose the game, for each artifact and creature they
// controlled, create a token that's a copy of it."
func registerContract(t *testing.T) {
	effectstest.RegisterForTest(t, effects.Spec{
		OracleID: contractOracle,
		Name:     contractOracle,
		Activated: []effects.ActivatedAbility{{
			Label:   "{0}: Target player gets a contract counter",
			Targets: effects.TargetPlayer("target player"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				who := effects.TargetedPlayerOf(ctx)
				if who == uuid.Nil {
					return nil
				}
				if err := g.AddPlayerCounterByForEffect(item.Controller, who, game.CounterContract, 1); err != nil {
					return err
				}
				return effects.WhenPlayerLosesWithCounter{
					Player:  who,
					Counter: game.CounterContract,
					Label:   "Contract — copies of each artifact and creature they controlled",
					Body:    effects.TokenCopiesOfDepartedArtifactsAndCreaturesBody,
				}.Apply(ctx)
			},
		}},
	})
}

func contractPermanent(g *game.Game, owner uuid.UUID, name, typeLine string) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{
			InstanceID: id, Name: name, TypeLine: typeLine,
			Power: 3, Toughness: 3, Owner: owner, Controller: owner,
		})
	})
	return id
}

func countNamed(g *game.Game, controller uuid.UUID, name string) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.Name == name {
			n++
		}
	}
	return n
}

func giveContract(t *testing.T, g *game.Game, me, them *game.Player) uuid.UUID {
	t.Helper()
	src := effectstest.PushCatalogPermanent(g, me.ID, "Highwater", "Legendary Creature — Human", contractOracle, false)
	if err := g.ActivateCatalogAbility(me.ID, src, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: them.ID}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if them.Counters[game.CounterContract] != 1 {
		t.Fatalf("contract counters = %d, want 1", them.Counters[game.CounterContract])
	}
	return src
}

// Four seats: the marked player concedes, and their artifact and creature
// come back as tokens under the trigger's controller; their land does not.
// The source left the battlefield first, and it still triggers.
func TestContractLossCopiesTheirArtifactsAndCreatures(t *testing.T) {
	registerContract(t)
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[2]
	src := giveContract(t, g, me, them)
	contractPermanent(g, them.ID, "Their Golem", "Artifact Creature — Golem")
	contractPermanent(g, them.ID, "Their Rock", "Artifact")
	contractPermanent(g, them.ID, "Their Bear", "Creature — Bear")
	contractPermanent(g, them.ID, "Their Forest", "Basic Land — Forest")
	// The source leaves first: the delayed trigger owes it nothing.
	g.WithWriteLock(func() { _, _ = g.Battlefield.Remove(src) })

	if err := g.Concede(them.ID); err != nil {
		t.Fatalf("concede: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, name := range []string{"Their Golem", "Their Rock", "Their Bear"} {
		if n := countNamed(g, me.ID, name); n != 1 {
			t.Errorf("%s copies under the trigger's controller = %d, want 1", name, n)
		}
	}
	if n := countNamed(g, me.ID, "Their Forest"); n != 0 {
		t.Errorf("a land was copied (%d) — only artifacts and creatures are", n)
	}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == me.ID && c.Name == "Their Bear" && !c.IsToken() {
			t.Error("the copy of the bear is not a token")
		}
	}
}

// "For as long as they have a contract counter": a player who lost the
// counter before leaving does not trigger it.
func TestContractLossNeedsTheCounterAsTheyLeave(t *testing.T) {
	registerContract(t)
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	giveContract(t, g, me, them)
	contractPermanent(g, them.ID, "Their Bear", "Creature — Bear")
	g.WithWriteLock(func() { them.Counters[game.CounterContract] = 0 })
	if err := g.Concede(them.ID); err != nil {
		t.Fatalf("concede: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := countNamed(g, me.ID, "Their Bear"); n != 0 {
		t.Errorf("copies = %d, want none without a contract counter", n)
	}
}

// Another player losing does not trigger it.
func TestContractLossIsTheMarkedPlayersOnly(t *testing.T) {
	registerContract(t)
	g := effectstest.NewCatalogGame(t)
	me, them, other := g.Seats[0], g.Seats[1], g.Seats[3]
	giveContract(t, g, me, them)
	contractPermanent(g, other.ID, "Other Bear", "Creature — Bear")
	if err := g.Concede(other.ID); err != nil {
		t.Fatalf("concede: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := countNamed(g, me.ID, "Other Bear"); n != 0 {
		t.Errorf("copies = %d, want none — the other player had no contract", n)
	}
}

// The record survives a restore point: what they controlled as they left
// is carried on the player.
func TestDepartedPermanentsSurviveAClone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	them := g.Seats[1]
	contractPermanent(g, them.ID, "Their Bear", "Creature — Bear")
	if err := g.Concede(them.ID); err != nil {
		t.Fatalf("concede: %v", err)
	}
	var got []game.DepartedPermanent
	g.WithWriteLock(func() { got = g.DepartedPermanentsForEffect(them.ID) })
	if len(got) != 1 || got[0].Source.Name != "Their Bear" || !got[0].HasType("creature") {
		t.Fatalf("departed record = %+v, want the bear typed creature", got)
	}
}

// Curse of Vengeance's "When enchanted player loses the game, you gain X
// life … where X is the number of spite counters on this Aura" needs no new
// seam: the Aura's own trigger sees EventPlayerEliminated while the player
// is still attached (it is emitted before CR 800.4a), and X is read off the
// Aura as it last existed once the state-based action has put it into the
// graveyard (CR 608.2h). PermanentForEffect answers both.
func TestAnAuraOnAPlayerTriggersOnTheirLoss(t *testing.T) {
	fired := false
	const oracle = "test-6123-curse-on-loss"
	effectstest.RegisterForTest(t, effects.Spec{
		OracleID: oracle,
		Name:     oracle,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventPlayerEliminated, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return source.AttachedTo.Kind == game.TargetPlayer && source.AttachedTo.ID == ev.Actor
			}, "Curse — gain X life", effects.Do(effects.Instruction(func(ctx *effects.Context) error {
				// The Aura as it is, or as it last existed once a state-based
				// action has taken it (CR 608.2h).
				x := 0
				if ref, ok := ctx.Game.SourceObjectForEffect(ctx.Item); ok {
					if info, ok := ctx.Game.PermanentForEffect(ref); ok {
						x = info.Counters["spite"]
					}
				}
				fired = true
				return effects.GainLife{Amount: x}.Apply(ctx)
			}))),
		},
	})
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	curse := uuid.New()
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{
			InstanceID: curse, Name: "Curse", TypeLine: "Enchantment — Aura Curse", OracleID: oracle,
			Owner: me.ID, Controller: me.ID, Counters: map[string]int{"spite": 3},
			AttachedTo: game.TargetRef{Kind: game.TargetPlayer, ID: them.ID},
		})
	})
	life := me.Life
	if err := g.Concede(them.ID); err != nil {
		t.Fatalf("concede: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !fired {
		t.Fatal("the trigger never resolved")
	}
	if me.Life != life+3 {
		t.Errorf("life = %d, want %d — three spite counters", me.Life, life+3)
	}
}
